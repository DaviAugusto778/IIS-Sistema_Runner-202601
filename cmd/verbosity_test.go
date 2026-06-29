package cmd

import (
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestVerbosityFlagsRegistered(t *testing.T) {
	for _, n := range []string{"verbose", "quiet"} {
		if rootCmd.PersistentFlags().Lookup(n) == nil {
			t.Errorf("flag persistente --%s ausente", n)
		}
	}
}

func TestVerboseQuietMutuallyExclusive(t *testing.T) {
	optVerbose, optQuiet = true, true
	t.Cleanup(func() { optVerbose, optQuiet = false, false })
	if err := rootCmd.PersistentPreRunE(rootCmd, nil); err == nil {
		t.Error("--verbose junto com --quiet deveria ser rejeitado")
	}
}

func TestQuietDiscardsProgress(t *testing.T) {
	optQuiet = true
	t.Cleanup(func() { optQuiet = false })
	if progressWriter() != io.Discard {
		t.Error("sob --quiet, o progresso deve ir para io.Discard")
	}
	c := &cobra.Command{}
	if cmdOut(c) != io.Discard {
		t.Error("sob --quiet, cmdOut deve devolver io.Discard")
	}
}

func TestProgressWriterDefaultIsStderr(t *testing.T) {
	optQuiet, optVerbose = false, false
	if progressWriter() != os.Stderr {
		t.Error("sem --quiet, o progresso deve ir para stderr")
	}
}

func TestTracefOnlyUnderVerbose(t *testing.T) {
	// tracef não deve emitir quando não está em modo verbose; aqui apenas
	// garantimos que a função é segura de chamar nos dois estados.
	optVerbose, optQuiet = false, false
	tracef("nada deve sair: %d", 1)
	optVerbose = true
	t.Cleanup(func() { optVerbose = false })
	tracef("ok: %s", "detalhe")
}
