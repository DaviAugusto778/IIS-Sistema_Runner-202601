package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/runtime"
	"github.com/spf13/cobra"
)

const (
	shutdownPath    = "/shutdown"
	shutdownTimeout = 10 * time.Second
	shutdownPoll    = 200 * time.Millisecond
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Encerra a instância em execução (via POST /shutdown)",
	Long: `Lê ~/.hubsaude/simulador.json para descobrir porta/PID, solicita
shutdown gracioso via POST http://localhost:<porta>/shutdown e aguarda
o processo encerrar antes de remover o state file.

Se o state file não existir ou o PID registrado já estiver morto, o
comando trata como sucesso (idempotência) e apenas limpa state órfão.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runStop(cmd.OutOrStdout(), shutdownTimeout)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(out io.Writer, timeout time.Duration) error {
	st, err := runtime.Load(stateName)
	if errors.Is(err, os.ErrNotExist) {
		_, _ = fmt.Fprintln(out, "Simulador ja estava parado (nenhum state file).")
		return nil
	}
	if err != nil {
		return err
	}

	if !runtime.IsAlive(st.PID) {
		_, _ = fmt.Fprintf(out, "Simulador ja estava parado (PID %d nao esta vivo); removendo state orfao.\n", st.PID)
		return runtime.Delete(stateName)
	}

	url := fmt.Sprintf("http://localhost:%d%s", st.Port, shutdownPath)
	if err := postShutdown(url, timeout); err != nil {
		return fmt.Errorf("simulador: solicitar shutdown em %s: %w", url, err)
	}

	if err := waitFor(func() bool { return !runtime.IsAlive(st.PID) }, timeout, shutdownPoll); err != nil {
		return fmt.Errorf("simulador: PID %d ainda vivo apos shutdown: %w", st.PID, err)
	}

	if err := runtime.Delete(stateName); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "Simulador encerrado (PID %d, porta %d).\n", st.PID, st.Port)
	return nil
}

func postShutdown(url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("criar request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status HTTP %d", resp.StatusCode)
	}
	return nil
}

func waitFor(check func() bool, timeout, poll time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return nil
		}
		time.Sleep(poll)
	}
	return fmt.Errorf("timeout apos %s", timeout)
}
