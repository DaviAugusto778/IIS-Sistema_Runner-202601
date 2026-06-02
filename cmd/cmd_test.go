package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func findSubcommand(parent *cobra.Command, name string) *cobra.Command {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestRootHasExpectedSubcommands(t *testing.T) {
	wanted := []string{"sign", "validate", "version"}
	for _, name := range wanted {
		if findSubcommand(rootCmd, name) == nil {
			t.Errorf("subcomando %q não foi registrado no rootCmd", name)
		}
	}
}

func TestSignHasOptionalFlags(t *testing.T) {
	sign := findSubcommand(rootCmd, "sign")
	if sign == nil {
		t.Fatal("sign não registrado")
	}
	for _, name := range []string{"content", "token"} {
		f := sign.Flags().Lookup(name)
		if f == nil {
			t.Errorf("flag %q não existe em sign", name)
			continue
		}
		// Critério E3: a validação é do JAR. Nenhuma flag pode ser
		// marcada como required no Cobra.
		if isRequired(sign, name) {
			t.Errorf("flag %q em sign não pode ser required (critério E3)", name)
		}
	}
}

func TestValidateHasOptionalFlags(t *testing.T) {
	validate := findSubcommand(rootCmd, "validate")
	if validate == nil {
		t.Fatal("validate não registrado")
	}
	for _, name := range []string{"content", "signature"} {
		f := validate.Flags().Lookup(name)
		if f == nil {
			t.Errorf("flag %q não existe em validate", name)
			continue
		}
		if isRequired(validate, name) {
			t.Errorf("flag %q em validate não pode ser required (critério E3)", name)
		}
	}
}

func TestVersionDefaultsAreSafe(t *testing.T) {
	// Garante que os placeholders existem; a injeção via ldflags ocorre no CI.
	if Version == "" || Commit == "" || Date == "" {
		t.Errorf("Version/Commit/Date precisam de defaults não-vazios; got %q/%q/%q", Version, Commit, Date)
	}
}

func isRequired(c *cobra.Command, flagName string) bool {
	annotations := c.Flags().Lookup(flagName).Annotations
	required, ok := annotations[cobra.BashCompOneRequiredFlag]
	if !ok || len(required) == 0 {
		return false
	}
	return required[0] == "true"
}
