package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
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
