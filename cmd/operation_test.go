package cmd

import (
	"strings"
	"testing"
)

func TestLocalArgsSign(t *testing.T) {
	args := localArgs("sign", map[string]string{"content": "ola", "token": "pin"})
	got := strings.Join(args, " ")
	want := "sign --content ola --token pin"
	if got != want {
		t.Errorf("localArgs = %q, quer %q", got, want)
	}
}

func TestLocalArgsValidate(t *testing.T) {
	args := localArgs("validate", map[string]string{"content": "ola", "signature": "SIG=="})
	got := strings.Join(args, " ")
	want := "validate --content ola --signature SIG=="
	if got != want {
		t.Errorf("localArgs = %q, quer %q", got, want)
	}
}

func TestLocalArgsOmiteChavesAusentes(t *testing.T) {
	args := localArgs("sign", map[string]string{"content": "ola"})
	got := strings.Join(args, " ")
	if got != "sign --content ola" {
		t.Errorf("localArgs = %q, quer apenas content", got)
	}
}

func TestSignValidateTemFlagLocal(t *testing.T) {
	for _, name := range []string{"sign", "validate"} {
		c := findSubcommand(rootCmd, name)
		if c == nil {
			t.Fatalf("%s nao registrado", name)
		}
		if c.Flags().Lookup("local") == nil {
			t.Errorf("flag --local ausente em %q", name)
		}
		// --local nao pode ser required (criterio E3 continua valendo)
		if isRequired(c, "local") {
			t.Errorf("flag --local em %q nao pode ser required", name)
		}
	}
}
