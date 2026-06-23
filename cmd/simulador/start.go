package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/release"
	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/runtime"
	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/server"
	"github.com/spf13/cobra"
)

const (
	simuladorPort    = 8443
	infoPath         = "/api/info"
	readinessTimeout = 60 * time.Second
	readinessPoll    = 500 * time.Millisecond
	downloadTimeout  = 5 * time.Minute
)

var startSource string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Baixa (se preciso) e inicia o Simulador em background",
	Long: `Garante simulador.jar em cache (~/.hubsaude/) baixando-o do
release.json apontado por --source, verifica se a porta 8443 esta
livre, lanca java -jar em background, persiste PID/porta em
~/.hubsaude/simulador.json e aguarda GET /api/info responder antes
de declarar o servico pronto.

Saidas do JVM sao redirecionadas para ~/.hubsaude/simulador.log.

Pre-requisito: 'java' disponivel no PATH (provisionamento automatico
de JDK fica para US-04.1).`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runStart(cmd.OutOrStdout(), startSource, readinessTimeout)
	},
}

func init() {
	startCmd.Flags().StringVar(&startSource, "source", "",
		"URL do release.json (default: "+release.DefaultManifestURL+")")
	rootCmd.AddCommand(startCmd)
}

func runStart(out io.Writer, source string, timeout time.Duration) error {
	if err := assertNotRunning(out); err != nil {
		return err
	}
	if err := server.AssertPortFree(simuladorPort); err != nil {
		return fmt.Errorf("simulador: porta %d nao disponivel: %w", simuladorPort, err)
	}

	java, err := exec.LookPath("java")
	if err != nil {
		return fmt.Errorf("simulador: 'java' nao encontrado no PATH (US-04.1 cobrira auto-provisao): %w", err)
	}

	_, _ = fmt.Fprintln(out, "Garantindo simulador.jar local...")
	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()
	jar, err := release.EnsureSimulador(ctx, nil, source)
	if err != nil {
		return fmt.Errorf("simulador: obter jar: %w", err)
	}

	logPath, err := simuladorLogPath()
	if err != nil {
		return err
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("simulador: abrir log %s: %w", logPath, err)
	}
	defer func() { _ = logFile.Close() }()

	_, _ = fmt.Fprintf(out, "Iniciando: %s -jar %s (porta %d)\n", java, jar, simuladorPort)
	cmd := exec.Command(java, "-jar", jar, fmt.Sprintf("--server.port=%d", simuladorPort))
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	server.Detach(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("simulador: spawn: %w", err)
	}
	pid := cmd.Process.Pid
	if err := cmd.Process.Release(); err != nil {
		_, _ = fmt.Fprintf(out, "Aviso: cmd.Process.Release: %v\n", err)
	}

	if err := runtime.Save(stateName, runtime.State{PID: pid, Port: simuladorPort}); err != nil {
		_ = server.KillPID(pid)
		return fmt.Errorf("simulador: salvar state: %w", err)
	}
	_, _ = fmt.Fprintf(out, "PID %d gravado; aguardando GET %s ficar pronto (timeout %s)...\n", pid, infoPath, timeout)

	alive := func() bool { return runtime.IsAlive(pid) }
	if err := server.WaitReady(simuladorPort, alive, timeout, readinessPoll, infoPath); err != nil {
		_ = server.KillPID(pid)
		_ = runtime.Delete(stateName)
		return fmt.Errorf("simulador: %w (logs: %s)", err, logPath)
	}

	_, _ = fmt.Fprintf(out, "Simulador pronto em http://localhost:%d (PID %d, logs em %s).\n",
		simuladorPort, pid, logPath)
	return nil
}

func assertNotRunning(out io.Writer) error {
	st, err := runtime.Load(stateName)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !runtime.IsAlive(st.PID) {
		_, _ = fmt.Fprintf(out, "State orfao detectado (PID %d morto); removendo.\n", st.PID)
		return runtime.Delete(stateName)
	}
	return fmt.Errorf("simulador ja em execucao: PID=%d porta=%d (use `simulador stop` antes)",
		st.PID, st.Port)
}

func simuladorLogPath() (string, error) {
	dir, err := release.CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "simulador.log"), nil
}
