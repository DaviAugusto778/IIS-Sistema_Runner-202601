package main

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

// withFakeHome aponta ~/.hubsaude para um diretório temporário, isolando
// cada teste do estado real do usuário.
func withFakeHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
}

func TestRunStopNoStateFile(t *testing.T) {
	withFakeHome(t)
	var out bytes.Buffer
	if err := runStop(&out, time.Second); err != nil {
		t.Fatalf("runStop: %v", err)
	}
	if !strings.Contains(out.String(), "ja estava parado") {
		t.Errorf("output = %q, queria mencao a 'ja estava parado'", out.String())
	}
}

func TestRunStopDeadPIDCleansState(t *testing.T) {
	withFakeHome(t)
	if err := runtime.Save(stateName, runtime.State{PID: 999999999, Port: 8443}); err != nil {
		t.Fatalf("setup Save: %v", err)
	}
	var out bytes.Buffer
	if err := runStop(&out, time.Second); err != nil {
		t.Fatalf("runStop: %v", err)
	}
	if _, err := runtime.Load(stateName); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("state deveria ter sido removido; Load retornou: %v", err)
	}
	if !strings.Contains(out.String(), "orfao") {
		t.Errorf("output = %q, queria mencao a 'orfao'", out.String())
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
	if err := postShutdown(srv.URL+"/shutdown", time.Second); err != nil {
		t.Fatalf("postShutdown: %v", err)
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
	if err := postShutdown(srv.URL, time.Second); err == nil {
		t.Errorf("queria erro para HTTP 500, got nil")
	}
}

func TestPostShutdownTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(500 * time.Millisecond):
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	if err := postShutdown(srv.URL, 50*time.Millisecond); err == nil {
		t.Errorf("queria timeout, got nil")
	}
}

func TestPostShutdownConnRefused(t *testing.T) {
	// httptest.NewServer + Close devolve uma URL com porta ja liberada,
	// o que reproduz "connection refused" sem hardcodar porta.
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	url := srv.URL
	srv.Close()
	if err := postShutdown(url, time.Second); err == nil {
		t.Errorf("queria erro de conexao recusada, got nil")
	}
}

func TestWaitForReturnsWhenTrue(t *testing.T) {
	var ready atomic.Bool
	go func() {
		time.Sleep(20 * time.Millisecond)
		ready.Store(true)
	}()
	if err := waitFor(ready.Load, time.Second, 5*time.Millisecond); err != nil {
		t.Errorf("waitFor: %v", err)
	}
}

func TestWaitForTimeout(t *testing.T) {
	if err := waitFor(func() bool { return false }, 50*time.Millisecond, 10*time.Millisecond); err == nil {
		t.Errorf("queria timeout, got nil")
	}
}
