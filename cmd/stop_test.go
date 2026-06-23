package cmd

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/runtime"
)

func TestStopRegistrado(t *testing.T) {
	if findSubcommand(rootCmd, "stop") == nil {
		t.Error("subcomando 'stop' nao foi registrado no rootCmd")
	}
}

func TestStopPortFlagDefaultZero(t *testing.T) {
	stop := findSubcommand(rootCmd, "stop")
	if stop == nil {
		t.Fatal("stop nao registrado")
	}
	f := stop.Flags().Lookup("port")
	if f == nil {
		t.Fatal("flag 'port' nao existe em stop")
	}
	if f.DefValue != "0" {
		t.Errorf("default de --port = %q, quer 0", f.DefValue)
	}
}

func TestRunStopNoState(t *testing.T) {
	withFakeHome(t)
	var out bytes.Buffer
	if err := runStop(&out, 0, time.Second); err != nil {
		t.Fatalf("runStop: %v", err)
	}
	if !strings.Contains(out.String(), "ja estava parado") {
		t.Errorf("output = %q, queria 'ja estava parado'", out.String())
	}
}

func TestRunStopDeadPIDCleansState(t *testing.T) {
	withFakeHome(t)
	if err := runtime.Save(stateName, runtime.State{PID: 999999999, Port: defaultServerPort}); err != nil {
		t.Fatalf("setup Save: %v", err)
	}
	var out bytes.Buffer
	if err := runStop(&out, 0, time.Second); err != nil {
		t.Fatalf("runStop: %v", err)
	}
	if _, err := runtime.Load(stateName); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("state deveria ter sido removido; Load retornou: %v", err)
	}
	if !strings.Contains(out.String(), "orfao") {
		t.Errorf("output = %q, queria mencao a 'orfao'", out.String())
	}
}

func TestRunStopByPortEnviaShutdown(t *testing.T) {
	withFakeHome(t)
	var gotShutdown atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == shutdownPath && r.Method == http.MethodPost {
			gotShutdown.Store(true)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	port := portOf(t, srv)

	// timeout curto: o servidor de teste continua "healthy", então o WaitFor
	// interno expira rápido (resultado ignorado no caminho por porta).
	var out bytes.Buffer
	if err := runStop(&out, port, 150*time.Millisecond); err != nil {
		t.Fatalf("runStop: %v", err)
	}
	if !gotShutdown.Load() {
		t.Error("servidor nao recebeu POST /shutdown")
	}
	if !strings.Contains(out.String(), "encerrado na porta") {
		t.Errorf("output = %q, queria confirmacao com a porta", out.String())
	}
}
