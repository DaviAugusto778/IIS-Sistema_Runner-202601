package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
)

// pkcs11SlotUnset é o sentinela de "slot não informado" para as flags
// --pkcs11-slot. O slot 0 é válido, então não serve como ausência.
const pkcs11SlotUnset = -1

// pkcs11JarArgs traduz a configuração PKCS#11 em argumentos para o
// assinador.jar (US-02.5). Sem biblioteca, devolve nil — o JAR usa o
// FakeSignatureService e o dispositivo nunca é acionado. O slot só é repassado
// quando informado (>= 0).
func pkcs11JarArgs(lib string, slot int) []string {
	if lib == "" {
		return nil
	}
	args := []string{"--pkcs11-lib", lib}
	if slot >= 0 {
		args = append(args, "--pkcs11-slot", strconv.Itoa(slot))
	}
	return args
}

// pkcs11ArgsFromFlags lê --pkcs11-lib/--pkcs11-slot de um comando e devolve os
// argumentos correspondentes para o JAR. Seguro quando as flags não existem no
// comando (ex.: validate não expõe PKCS#11): devolve nil.
func pkcs11ArgsFromFlags(cmd *cobra.Command) []string {
	if !cmd.Flags().Changed("pkcs11-lib") {
		return nil
	}
	lib, _ := cmd.Flags().GetString("pkcs11-lib")
	slot := pkcs11SlotUnset
	if cmd.Flags().Changed("pkcs11-slot") {
		slot, _ = cmd.Flags().GetInt("pkcs11-slot")
	}
	return pkcs11JarArgs(lib, slot)
}
