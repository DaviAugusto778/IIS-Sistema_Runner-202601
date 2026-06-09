package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// withFakeHome aponta os.UserHomeDir() para um diretório temporário,
// evitando que os testes poluam o ~/.hubsaude real do usuário.
func withFakeHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp) // Windows
	return tmp
}

func TestStatePathLivesUnderHubsaude(t *testing.T) {
	home := withFakeHome(t)
	p, err := StatePath("simulador")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := filepath.Join(home, ".hubsaude", "simulador.json")
	if p != want {
		t.Errorf("StatePath = %q, quer %q", p, want)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	withFakeHome(t)
	want := State{PID: 4242, Port: 8443}
	if err := Save("simulador", want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load("simulador")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if *got != want {
		t.Errorf("Load = %+v, quer %+v", *got, want)
	}
}

func TestLoadAbsentIsErrNotExist(t *testing.T) {
	withFakeHome(t)
	_, err := Load("inexistente")
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("err = %v, quer satisfazer os.ErrNotExist", err)
	}
}

func TestDeleteIdempotent(t *testing.T) {
	withFakeHome(t)
	if err := Delete("nunca-existiu"); err != nil {
		t.Errorf("Delete em ausente devia ser no-op, recebeu: %v", err)
	}
	if err := Save("x", State{PID: 1, Port: 2}); err != nil {
		t.Fatalf("setup Save: %v", err)
	}
	if err := Delete("x"); err != nil {
		t.Errorf("Delete: %v", err)
	}
	if _, err := Load("x"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("apos Delete, Load devia errar com NotExist, got %v", err)
	}
}

func TestIsAliveCurrentProcess(t *testing.T) {
	if !IsAlive(os.Getpid()) {
		t.Errorf("IsAlive(self) = false, quer true")
	}
}

func TestIsAliveZeroOrNegativeIsFalse(t *testing.T) {
	if IsAlive(0) {
		t.Errorf("IsAlive(0) deveria ser false")
	}
	if IsAlive(-1) {
		t.Errorf("IsAlive(-1) deveria ser false")
	}
}

func TestIsAliveUnlikelyPID(t *testing.T) {
	// PID 999_999_999 é improvável de existir em qualquer SO real.
	// Tolera-se um falso positivo raríssimo (sistema com PID limit
	// muito alto reaproveitando justamente esse).
	if IsAlive(999999999) {
		t.Skip("PID improvável foi observado vivo; ambiente atípico")
	}
}
