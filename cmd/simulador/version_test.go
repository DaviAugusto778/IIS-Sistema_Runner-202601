package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestVersionFlagOutputsTraceable garante que --version (criterio I) funciona no
// CLI simulador e devolve algo rastreável.
func TestVersionFlagOutputsTraceable(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--version"})
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute --version: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "simulador") {
		t.Errorf("--version deve identificar o binario 'simulador'; got %q", out)
	}
	if !strings.Contains(out, Version) {
		t.Errorf("--version deve conter a versao %q; got %q", Version, out)
	}
	if !strings.Contains(out, Commit) {
		t.Errorf("--version deve ser rastreavel (commit %q); got %q", Commit, out)
	}
}
