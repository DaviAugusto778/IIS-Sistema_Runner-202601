package cmd

import (
	"fmt"
	"os"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/invoker"
	"github.com/spf13/cobra"
)

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Cria uma assinatura digital simulada",
	Long: `Invoca o assinador.jar para criar uma assinatura digital simulada.

A validação dos parâmetros é responsabilidade exclusiva do assinador.jar
(autoridade única, conforme criterio E3). Esta CLI apenas repassa os
argumentos, o stdout, o stderr e o exit code do JAR.

Exemplos:
  assinatura sign --content "ola mundo"
  assinatura sign --content "$(cat documento.txt)" --token abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		invokeArgs := []string{"sign"}
		if cmd.Flags().Changed("content") {
			content, _ := cmd.Flags().GetString("content")
			invokeArgs = append(invokeArgs, "--content", content)
		}
		if cmd.Flags().Changed("token") {
			token, _ := cmd.Flags().GetString("token")
			invokeArgs = append(invokeArgs, "--token", token)
		}

		return runAndPassthrough(invokeArgs...)
	},
}

func init() {
	signCmd.Flags().String("content", "", "Conteúdo a ser assinado")
	signCmd.Flags().String("token", "", "Token de autenticação")
	rootCmd.AddCommand(signCmd)
}

// runAndPassthrough invoca o assinador.jar, repassa stdout/stderr para
// o processo atual e termina com o mesmo exit code do JAR. Falhas de
// execução (jar ausente, java fora do PATH) saem com código 2 e
// mensagem em stderr.
func runAndPassthrough(args ...string) error {
	result, err := invoker.Invoke(args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(2)
	}
	if len(result.Stderr) > 0 {
		_, _ = os.Stderr.Write(result.Stderr)
	}
	if len(result.Stdout) > 0 {
		_, _ = os.Stdout.Write(result.Stdout)
	}
	os.Exit(result.ExitCode)
	return nil
}
