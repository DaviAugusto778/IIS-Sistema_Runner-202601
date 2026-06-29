package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPkcs11JarArgsSemBibliotecaEhNil(t *testing.T) {
	if got := pkcs11JarArgs("", 0); got != nil {
		t.Errorf("sem biblioteca deve devolver nil, obtido %v", got)
	}
}

func TestPkcs11JarArgsComSlot(t *testing.T) {
	got := strings.Join(pkcs11JarArgs("/lib/x.so", 2), " ")
	want := "--pkcs11-lib /lib/x.so --pkcs11-slot 2"
	if got != want {
		t.Errorf("pkcs11JarArgs = %q, quer %q", got, want)
	}
}

func TestPkcs11JarArgsSlotZeroEhRepassado(t *testing.T) {
	// slot 0 é válido (não é "ausência"): deve aparecer.
	got := strings.Join(pkcs11JarArgs("/lib/x.so", 0), " ")
	if !strings.Contains(got, "--pkcs11-slot 0") {
		t.Errorf("slot 0 deveria ser repassado, obtido %q", got)
	}
}

func TestPkcs11JarArgsSlotNaoInformadoEhOmitido(t *testing.T) {
	got := strings.Join(pkcs11JarArgs("/lib/x.so", pkcs11SlotUnset), " ")
	if strings.Contains(got, "--pkcs11-slot") {
		t.Errorf("slot não informado não deve emitir --pkcs11-slot, obtido %q", got)
	}
	if got != "--pkcs11-lib /lib/x.so" {
		t.Errorf("pkcs11JarArgs = %q, quer só a biblioteca", got)
	}
}

func TestPkcs11ArgsFromFlagsLeDoComando(t *testing.T) {
	c := &cobra.Command{Use: "x", Run: func(*cobra.Command, []string) {}}
	c.Flags().String("pkcs11-lib", "", "")
	c.Flags().Int("pkcs11-slot", pkcs11SlotUnset, "")
	c.SetArgs([]string{"--pkcs11-lib", "/lib/x.so", "--pkcs11-slot", "1"})
	if err := c.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := strings.Join(pkcs11ArgsFromFlags(c), " ")
	want := "--pkcs11-lib /lib/x.so --pkcs11-slot 1"
	if got != want {
		t.Errorf("pkcs11ArgsFromFlags = %q, quer %q", got, want)
	}
}

func TestPkcs11ArgsFromFlagsSeguroSemFlags(t *testing.T) {
	// validate não expõe flags PKCS#11: a leitura deve ser segura e devolver nil.
	c := &cobra.Command{Use: "x", Run: func(*cobra.Command, []string) {}}
	if got := pkcs11ArgsFromFlags(c); got != nil {
		t.Errorf("comando sem flags PKCS#11 deve devolver nil, obtido %v", got)
	}
}

func TestSignTemFlagsPkcs11(t *testing.T) {
	sign := findSubcommand(rootCmd, "sign")
	if sign == nil {
		t.Fatal("sign não registrado")
	}
	for _, name := range []string{"pkcs11-lib", "pkcs11-slot"} {
		if sign.Flags().Lookup(name) == nil {
			t.Errorf("flag %q ausente em sign", name)
			continue
		}
		if isRequired(sign, name) {
			t.Errorf("flag %q em sign não pode ser required (PKCS#11 é opcional)", name)
		}
	}
}

func TestStartTemFlagsPkcs11(t *testing.T) {
	start := findSubcommand(rootCmd, "start")
	if start == nil {
		t.Fatal("start não registrado")
	}
	for _, name := range []string{"pkcs11-lib", "pkcs11-slot"} {
		if start.Flags().Lookup(name) == nil {
			t.Errorf("flag %q ausente em start", name)
		}
	}
}

func TestValidateNaoTemFlagsPkcs11(t *testing.T) {
	// Validação não interage com o dispositivo (ADR-0005): não deve oferecer
	// flags PKCS#11, para não sugerir o contrário.
	validate := findSubcommand(rootCmd, "validate")
	if validate == nil {
		t.Fatal("validate não registrado")
	}
	if validate.Flags().Lookup("pkcs11-lib") != nil {
		t.Error("validate não deveria expor --pkcs11-lib (validação não usa o dispositivo)")
	}
}
