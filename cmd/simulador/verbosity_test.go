package main

import (
	"io"
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

func TestQuietDiscardsCmdOut(t *testing.T) {
	optQuiet = true
	t.Cleanup(func() { optQuiet = false })
	if cmdOut(&cobra.Command{}) != io.Discard {
		t.Error("sob --quiet, cmdOut deve devolver io.Discard")
	}
}
