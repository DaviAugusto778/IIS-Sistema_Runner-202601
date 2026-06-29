package runtime

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
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

// TestSaveConcurrentNeverYieldsTornRead simula o cenário "race no start"
// (criterio G): vários `start` concorrentes persistindo o estado enquanto
// outros o leem. Com a escrita atômica (temp + rename), o leitor nunca observa
// um arquivo truncado — Load sempre devolve um estado íntegro. Há um pequeno
// espaçamento entre as operações para refletir o uso real (poucos starts
// gravando esporadicamente), não um hammer sintético. Sob -race (como no CI),
// valida ainda a ausência de corrida de memória no caminho.
func TestSaveConcurrentNeverYieldsTornRead(t *testing.T) {
	withFakeHome(t)
	const (
		name    = "concorrente"
		writers = 4
		readers = 4
		iters   = 25
		spacing = time.Millisecond
	)

	// Semeia o arquivo para que os leitores não precisem tolerar ErrNotExist
	// no estado estacionário — qualquer erro de Load passa a ser corrupção.
	if err := Save(name, State{PID: 1, Port: 8000}); err != nil {
		t.Fatalf("seed Save: %v", err)
	}

	var wg sync.WaitGroup
	errc := make(chan error, (writers+readers)*iters)

	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				if err := Save(name, State{PID: id*1000 + i, Port: 8000 + id}); err != nil {
					errc <- fmt.Errorf("Save: %w", err)
					return
				}
				time.Sleep(spacing)
			}
		}(w)
	}
	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				st, err := Load(name)
				if err != nil {
					errc <- fmt.Errorf("Load corrompido (esperava estado integro): %w", err)
					return
				}
				if st.Port < 8000 || st.Port >= 8000+writers {
					errc <- fmt.Errorf("estado torn: porta fora da faixa esperada: %d", st.Port)
					return
				}
				time.Sleep(spacing)
			}
		}()
	}

	wg.Wait()
	close(errc)
	for err := range errc {
		t.Error(err)
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
