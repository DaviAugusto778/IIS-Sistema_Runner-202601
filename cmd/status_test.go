package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/runtime"
)

func TestStatusRegistrado(t *testing.T) {
	if findSubcommand(rootCmd, "status") == nil {
		t.Error("subcomando 'status' nao foi registrado no rootCmd")
	}
}

func TestRunStatusParado(t *testing.T) {
	withFakeHome(t)
	var out bytes.Buffer
	code, err := runStatus(&out)
	if err != nil {
		t.Fatalf("runStatus: %v", err)
	}
	if code != 1 {
		t.Errorf("exit code = %d, quer 1 (parado)", code)
	}
	if !strings.Contains(out.String(), "parado") {
		t.Errorf("output = %q, queria mencao a 'parado'", out.String())
	}
}

func TestRunStatusEmExecucao(t *testing.T) {
	withFakeHome(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == healthPath {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "no", http.StatusNotFound)
	}))
	defer srv.Close()

	if err := runtime.Save(stateName, runtime.State{PID: os.Getpid(), Port: portOf(t, srv)}); err != nil {
		t.Fatalf("setup Save: %v", err)
	}
	var out bytes.Buffer
	code, err := runStatus(&out)
	if err != nil {
		t.Fatalf("runStatus: %v", err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, quer 0 (em execucao)", code)
	}
	if !strings.Contains(out.String(), "execucao") {
		t.Errorf("output = %q, queria mencao a 'execucao'", out.String())
	}
}

func TestRunStatusRegistradoSemResponder(t *testing.T) {
	withFakeHome(t)
	if err := runtime.Save(stateName, runtime.State{PID: os.Getpid(), Port: freeLocalPort(t)}); err != nil {
		t.Fatalf("setup Save: %v", err)
	}
	var out bytes.Buffer
	code, err := runStatus(&out)
	if err != nil {
		t.Fatalf("runStatus: %v", err)
	}
	if code != 1 {
		t.Errorf("exit code = %d, quer 1 (sem responder)", code)
	}
	if !strings.Contains(out.String(), "nao responde") {
		t.Errorf("output = %q, queria mencao a 'nao responde'", out.String())
	}
}
