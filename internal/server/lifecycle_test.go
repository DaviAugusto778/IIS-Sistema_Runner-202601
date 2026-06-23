package server

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAssertPortFreeWhenFree(t *testing.T) {
	port := freePort(t)
	if err := AssertPortFree(port); err != nil {
		t.Errorf("queria porta livre, got %v", err)
	}
}

func TestAssertPortFreeWhenTaken(t *testing.T) {
	// Bind ao mesmo endereco (":port") que AssertPortFree usa, evitando
	// quirks de Windows onde 127.0.0.1:port nao bloqueia 0.0.0.0:port.
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("setup Listen: %v", err)
	}
	defer func() { _ = l.Close() }()
	port := l.Addr().(*net.TCPAddr).Port
	if err := AssertPortFree(port); err == nil {
		t.Errorf("queria erro de porta tomada, got nil")
	}
}

func TestWaitReadyReturnsOn2xx(t *testing.T) {
	const healthPath = "/health"
	var ready atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != healthPath {
			http.Error(w, "no", http.StatusNotFound)
			return
		}
		if !ready.Load() {
			http.Error(w, "starting", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	port := serverPort(t, srv)

	go func() {
		time.Sleep(30 * time.Millisecond)
		ready.Store(true)
	}()
	if err := WaitReady(port, func() bool { return true }, 2*time.Second, 10*time.Millisecond, healthPath); err != nil {
		t.Errorf("WaitReady: %v", err)
	}
}

func TestWaitReadyTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "starting", http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	port := serverPort(t, srv)

	if err := WaitReady(port, func() bool { return true }, 80*time.Millisecond, 10*time.Millisecond, "/health"); err == nil {
		t.Errorf("queria timeout, got nil")
	}
}

func TestWaitReadyProcessDies(t *testing.T) {
	// Sem servidor: probe falha. alive() false na primeira iteracao
	// para sinalizar morte imediata do JVM.
	err := WaitReady(0, func() bool { return false }, time.Second, 10*time.Millisecond, "/health")
	if err == nil || !strings.Contains(err.Error(), "morreu") {
		t.Errorf("err = %v, queria 'processo morreu'", err)
	}
}

func TestIsHealthyTrueOn2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "no", http.StatusNotFound)
	}))
	defer srv.Close()
	port := serverPort(t, srv)

	if !IsHealthy(port, "/health") {
		t.Error("queria IsHealthy=true para servidor respondendo 200")
	}
	if IsHealthy(port, "/inexistente") {
		t.Error("queria IsHealthy=false para path 404")
	}
}

func TestIsHealthyFalseQuandoSemServidor(t *testing.T) {
	port := freePort(t) // ninguem escutando
	if IsHealthy(port, "/health") {
		t.Error("queria IsHealthy=false quando nada escuta na porta")
	}
}

func TestPostShutdownSuccess(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	if err := PostShutdown(srv.URL+"/shutdown", time.Second); err != nil {
		t.Fatalf("PostShutdown: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("metodo = %q, quer POST", gotMethod)
	}
	if gotPath != "/shutdown" {
		t.Errorf("path = %q, quer /shutdown", gotPath)
	}
}

func TestPostShutdownNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	if err := PostShutdown(srv.URL, time.Second); err == nil {
		t.Errorf("queria erro para HTTP 500, got nil")
	}
}

func TestPostShutdownConnRefused(t *testing.T) {
	// httptest.NewServer + Close devolve uma URL com porta ja liberada,
	// reproduzindo "connection refused" sem hardcodar porta.
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	url := srv.URL
	srv.Close()
	if err := PostShutdown(url, time.Second); err == nil {
		t.Errorf("queria erro de conexao recusada, got nil")
	}
}

func TestWaitForReturnsWhenTrue(t *testing.T) {
	var ready atomic.Bool
	go func() {
		time.Sleep(20 * time.Millisecond)
		ready.Store(true)
	}()
	if err := WaitFor(ready.Load, time.Second, 5*time.Millisecond); err != nil {
		t.Errorf("WaitFor: %v", err)
	}
}

func TestWaitForTimeout(t *testing.T) {
	if err := WaitFor(func() bool { return false }, 50*time.Millisecond, 10*time.Millisecond); err == nil {
		t.Errorf("queria timeout, got nil")
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

func serverPort(t *testing.T, srv *httptest.Server) int {
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
