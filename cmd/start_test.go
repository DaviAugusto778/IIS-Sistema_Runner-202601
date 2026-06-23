package cmd

import (
	"bytes"
	"errors"
	"os"
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

func TestAssertNotRunningWhenNoState(t *testing.T) {
	withFakeHome(t)
	var out bytes.Buffer
	if err := assertNotRunning(&out); err != nil {
		t.Errorf("assertNotRunning: %v", err)
	}
}

func TestAssertNotRunningWhenAlive(t *testing.T) {
	withFakeHome(t)
	if err := runtime.Save(stateName, runtime.State{PID: os.Getpid(), Port: defaultServerPort}); err != nil {
		t.Fatalf("setup Save: %v", err)
	}
	var out bytes.Buffer
	err := assertNotRunning(&out)
	if err == nil {
		t.Fatalf("queria erro 'ja em execucao', got nil")
	}
	if !strings.Contains(err.Error(), "ja em execucao") {
		t.Errorf("err = %v, queria mencao a 'ja em execucao'", err)
	}
}

func TestAssertNotRunningClearsOrphanState(t *testing.T) {
	withFakeHome(t)
	if err := runtime.Save(stateName, runtime.State{PID: 999999999, Port: defaultServerPort}); err != nil {
		t.Fatalf("setup Save: %v", err)
	}
	var out bytes.Buffer
	if err := assertNotRunning(&out); err != nil {
		t.Errorf("assertNotRunning: %v", err)
	}
	if _, err := runtime.Load(stateName); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("state orfao deveria ter sido removido, Load retornou: %v", err)
	}
	if !strings.Contains(out.String(), "orfao") {
		t.Errorf("output = %q, queria mencao a 'orfao'", out.String())
	}
}
