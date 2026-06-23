package cmd

import (
	"bytes"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/runtime"
)

// withFakeHome aponta ~/.hubsaude para um diretório temporário, isolando
// cada teste do estado real do usuário.
func withFakeHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
}

// freeLocalPort devolve uma porta TCP livre (nada escutando nela).
func freeLocalPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freeLocalPort: %v", err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

// portOf extrai a porta de um servidor httptest.
func portOf(t *testing.T, srv *httptest.Server) int {
	t.Helper()
	_, portStr, err := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	p, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Atoi: %v", err)
	}
	return p
}

func TestStartRegistrado(t *testing.T) {
	if findSubcommand(rootCmd, "start") == nil {
		t.Error("subcomando 'start' nao foi registrado no rootCmd")
	}
}

func TestStartPortFlagDefault(t *testing.T) {
	start := findSubcommand(rootCmd, "start")
	if start == nil {
		t.Fatal("start nao registrado")
	}
	f := start.Flags().Lookup("port")
	if f == nil {
		t.Fatal("flag 'port' nao existe em start")
	}
	if f.DefValue != "8080" {
		t.Errorf("default de --port = %q, quer 8080", f.DefValue)
	}
}

func TestDetectInstanceNoState(t *testing.T) {
	withFakeHome(t)
	var out bytes.Buffer
	status, _, err := detectInstance(&out)
	if err != nil {
		t.Fatalf("detectInstance: %v", err)
	}
	if status != statusNotRunning {
		t.Errorf("status = %v, quer statusNotRunning", status)
	}
}

func TestDetectInstanceClearsOrphanState(t *testing.T) {
	withFakeHome(t)
	if err := runtime.Save(stateName, runtime.State{PID: 999999999, Port: defaultServerPort}); err != nil {
		t.Fatalf("setup Save: %v", err)
	}
	var out bytes.Buffer
	status, _, err := detectInstance(&out)
	if err != nil {
		t.Fatalf("detectInstance: %v", err)
	}
	if status != statusNotRunning {
		t.Errorf("status = %v, quer statusNotRunning (PID morto)", status)
	}
	if _, err := runtime.Load(stateName); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("state orfao deveria ter sido removido, Load retornou: %v", err)
	}
	if !strings.Contains(out.String(), "orfao") {
		t.Errorf("output = %q, queria mencao a 'orfao'", out.String())
	}
}

func TestDetectInstanceAliveMasSemResponder(t *testing.T) {
	withFakeHome(t)
	// Processo vivo (este teste), mas nada respondendo /health na porta.
	if err := runtime.Save(stateName, runtime.State{PID: os.Getpid(), Port: freeLocalPort(t)}); err != nil {
		t.Fatalf("setup Save: %v", err)
	}
	var out bytes.Buffer
	status, st, err := detectInstance(&out)
	if err != nil {
		t.Fatalf("detectInstance: %v", err)
	}
	if status != statusUnresponsive {
		t.Errorf("status = %v, quer statusUnresponsive", status)
	}
	if st == nil || st.PID != os.Getpid() {
		t.Errorf("esperava devolver o state da instancia viva")
	}
}

func TestDetectInstanceRunning(t *testing.T) {
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
	status, st, err := detectInstance(&out)
	if err != nil {
		t.Fatalf("detectInstance: %v", err)
	}
	if status != statusRunning {
		t.Errorf("status = %v, quer statusRunning", status)
	}
	if st == nil {
		t.Fatal("esperava state nao nulo")
	}
}
