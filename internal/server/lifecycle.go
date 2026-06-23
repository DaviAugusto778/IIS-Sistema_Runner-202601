// Package server reúne primitivas de ciclo de vida compartilhadas pelos
// CLIs (assinatura, simulador) para lançar o respectivo .jar como servidor
// em background: spawn desacoplado do console do CLI (spawn_unix.go,
// spawn_windows.go), verificação de porta livre e espera por readiness HTTP.
//
// O estado persistente (PID/porta) continua no pacote internal/runtime; este
// pacote cuida apenas do lançamento e da prontidão do processo.
package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

// readinessHTTPTimeout limita cada tentativa de health check para não travar
// o laço de readiness quando o servidor aceita a conexão mas não responde.
const readinessHTTPTimeout = 2 * time.Second

// AssertPortFree retorna erro se a porta TCP já estiver ocupada. Faz bind em
// ":port" (todas as interfaces) para evitar quirks de Windows em que
// 127.0.0.1:port não bloqueia 0.0.0.0:port.
func AssertPortFree(port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	return l.Close()
}

// WaitReady bloqueia até que GET http://localhost:<port><healthPath> responda
// 2xx (servidor pronto), até alive() retornar false (processo morreu) ou até
// timeout expirar. poll define o intervalo entre tentativas.
func WaitReady(port int, alive func() bool, timeout, poll time.Duration, healthPath string) error {
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("http://localhost:%d%s", port, healthPath)
	client := &http.Client{Timeout: readinessHTTPTimeout}
	for time.Now().Before(deadline) {
		if !alive() {
			return fmt.Errorf("processo morreu durante startup; verifique logs")
		}
		if probeReady(client, url) {
			return nil
		}
		time.Sleep(poll)
	}
	return fmt.Errorf("timeout %s aguardando %s", timeout, url)
}

func probeReady(client *http.Client, url string) bool {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// KillPID força o encerramento do processo com o PID indicado. Usado como
// rollback quando o startup falha após o spawn.
func KillPID(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
