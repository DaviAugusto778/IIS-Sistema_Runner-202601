package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/invoker"
	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/jdk"
	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/runtime"
	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/server"
	"github.com/spf13/cobra"
)

const (
	// stateName é o nome do arquivo de estado do assinador em
	// ~/.hubsaude/assinador.json (PID/porta da instância em modo servidor).
	stateName = "assinador"

	defaultServerPort = 8080
	healthPath        = "/health"
	readinessTimeout  = 30 * time.Second
	readinessPoll     = 300 * time.Millisecond
	jdkTimeout        = 15 * time.Minute
)

var startPort int

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o assinador.jar no modo servidor (background)",
	Long: `Sobe o assinador.jar como servidor HTTP em segundo plano, na porta
padrão (8080) ou na indicada por --port.

Localiza o assinador.jar (junto ao executável ou via ASSINADOR_JAR),
provisiona o JDK 21 automaticamente se necessário (US-04.1), verifica se
a porta está livre, lança 'java -jar assinador.jar server' desacoplado do
console, persiste PID/porta em ~/.hubsaude/assinador.json e aguarda
GET /health responder antes de declarar o servidor pronto.

Saídas do JVM são redirecionadas para ~/.hubsaude/assinador.log.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runStart(cmd.OutOrStdout(), startPort, readinessTimeout)
	},
}

func init() {
	startCmd.Flags().IntVar(&startPort, "port", defaultServerPort,
		"porta do servidor")
	rootCmd.AddCommand(startCmd)
}

func runStart(out io.Writer, port int, timeout time.Duration) error {
	if err := assertNotRunning(out); err != nil {
		return err
	}
	if err := server.AssertPortFree(port); err != nil {
		return fmt.Errorf("assinatura: porta %d nao disponivel: %w", port, err)
	}

	jar, err := invoker.JarPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(jar); err != nil {
		return fmt.Errorf("assinatura: assinador.jar nao encontrado em %q: %w", jar, err)
	}

	_, _ = fmt.Fprintln(out, "Localizando/provisionando JDK 21...")
	ctx, cancel := context.WithTimeout(context.Background(), jdkTimeout)
	defer cancel()
	java, err := jdk.Ensure(ctx, nil)
	if err != nil {
		return fmt.Errorf("assinatura: localizar ou provisionar JDK 21: %w", err)
	}

	logPath, err := serverLogPath()
	if err != nil {
		return err
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("assinatura: abrir log %s: %w", logPath, err)
	}
	defer func() { _ = logFile.Close() }()

	_, _ = fmt.Fprintf(out, "Iniciando: %s -jar %s server --port %d\n", java, jar, port)
	cmd := exec.Command(java, "-jar", jar, "server", "--port", strconv.Itoa(port))
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	server.Detach(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("assinatura: spawn: %w", err)
	}
	pid := cmd.Process.Pid
	if err := cmd.Process.Release(); err != nil {
		_, _ = fmt.Fprintf(out, "Aviso: cmd.Process.Release: %v\n", err)
	}

	if err := runtime.Save(stateName, runtime.State{PID: pid, Port: port}); err != nil {
		_ = server.KillPID(pid)
		return fmt.Errorf("assinatura: salvar state: %w", err)
	}
	_, _ = fmt.Fprintf(out, "PID %d gravado; aguardando GET %s ficar pronto (timeout %s)...\n", pid, healthPath, timeout)

	alive := func() bool { return runtime.IsAlive(pid) }
	if err := server.WaitReady(port, alive, timeout, readinessPoll, healthPath); err != nil {
		_ = server.KillPID(pid)
		_ = runtime.Delete(stateName)
		return fmt.Errorf("assinatura: %w (logs: %s)", err, logPath)
	}

	_, _ = fmt.Fprintf(out, "Assinador pronto em http://localhost:%d (PID %d, logs em %s).\n",
		port, pid, logPath)
	return nil
}

// assertNotRunning falha se já há uma instância viva registrada; limpa o
// state quando o PID registrado não está mais vivo (state órfão).
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
	return fmt.Errorf("assinador ja em execucao: PID=%d porta=%d (use `assinatura stop` antes)",
		st.PID, st.Port)
}

func serverLogPath() (string, error) {
	dir, err := runtime.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "assinador.log"), nil
}
