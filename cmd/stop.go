package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/runtime"
	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/server"
	"github.com/spf13/cobra"
)

const (
	shutdownPath    = "/shutdown"
	shutdownTimeout = 10 * time.Second
	shutdownPoll    = 200 * time.Millisecond
)

var stopPort int

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Encerra o assinador em execução no modo servidor",
	Long: `Solicita o encerramento gracioso do assinador (POST /shutdown).

Sem --port, lê ~/.hubsaude/assinador.json para descobrir PID/porta, aguarda
o processo encerrar e remove o state file. É idempotente: se não houver
instância registrada (ou o PID já estiver morto), trata como sucesso e
apenas limpa state órfão.

Com --port <porta>, envia o shutdown diretamente àquela porta — útil para
encerrar uma instância específica mesmo sem state file.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runStop(cmdOut(cmd), stopPort, shutdownTimeout)
	},
}

func init() {
	stopCmd.Flags().IntVar(&stopPort, "port", 0,
		"porta do servidor a encerrar (default: a registrada em ~/.hubsaude)")
	rootCmd.AddCommand(stopCmd)
}

func runStop(out io.Writer, port int, timeout time.Duration) error {
	if port != 0 {
		return stopByPort(out, port, timeout)
	}
	return stopByState(out, timeout)
}

// stopByPort encerra a instância que escuta na porta indicada, sem depender
// do state file. Limpa o state se a porta registrada coincidir.
func stopByPort(out io.Writer, port int, timeout time.Duration) error {
	url := fmt.Sprintf("http://localhost:%d%s", port, shutdownPath)
	if err := server.PostShutdown(url, timeout); err != nil {
		return fmt.Errorf("assinatura: solicitar shutdown em %s: %w", url, err)
	}
	// aguarda o servidor parar de responder (não temos o PID neste caminho)
	_ = server.WaitFor(func() bool { return !server.IsHealthy(port, healthPath) }, timeout, shutdownPoll)

	if st, err := runtime.Load(stateName); err == nil && st.Port == port {
		_ = runtime.Delete(stateName)
	}
	_, _ = fmt.Fprintf(out, "Assinador encerrado na porta %d.\n", port)
	return nil
}

// stopByState encerra a instância registrada em ~/.hubsaude/assinador.json.
func stopByState(out io.Writer, timeout time.Duration) error {
	st, err := runtime.Load(stateName)
	if errors.Is(err, os.ErrNotExist) {
		_, _ = fmt.Fprintln(out, "Assinador ja estava parado (nenhum state file).")
		return nil
	}
	if err != nil {
		return err
	}
	if !runtime.IsAlive(st.PID) {
		_, _ = fmt.Fprintf(out, "Assinador ja estava parado (PID %d nao esta vivo); removendo state orfao.\n", st.PID)
		return runtime.Delete(stateName)
	}

	url := fmt.Sprintf("http://localhost:%d%s", st.Port, shutdownPath)
	if err := server.PostShutdown(url, timeout); err != nil {
		return fmt.Errorf("assinatura: solicitar shutdown em %s: %w", url, err)
	}
	if err := server.WaitFor(func() bool { return !runtime.IsAlive(st.PID) }, timeout, shutdownPoll); err != nil {
		return fmt.Errorf("assinatura: PID %d ainda vivo apos shutdown: %w", st.PID, err)
	}
	if err := runtime.Delete(stateName); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "Assinador encerrado (PID %d, porta %d).\n", st.PID, st.Port)
	return nil
}
