package cmd

import (
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Valida uma assinatura digital simulada",
	Long: `Invoca o assinador.jar para verificar se uma assinatura digital
corresponde ao conteúdo fornecido.

A validação dos parâmetros é responsabilidade exclusiva do assinador.jar
(autoridade única, conforme criterio E3). Esta CLI apenas repassa os
argumentos, o stdout, o stderr e o exit code do JAR.

Exemplos:
  assinatura validate --content "ola mundo" --signature "MOCKED_SIGNATURE_BASE64_=="`,
	RunE: func(cmd *cobra.Command, args []string) error {
		invokeArgs := []string{"validate"}
		if cmd.Flags().Changed("content") {
			content, _ := cmd.Flags().GetString("content")
			invokeArgs = append(invokeArgs, "--content", content)
		}
		if cmd.Flags().Changed("signature") {
			signature, _ := cmd.Flags().GetString("signature")
			invokeArgs = append(invokeArgs, "--signature", signature)
		}

		return runAndPassthrough(invokeArgs...)
	},
}

func init() {
	validateCmd.Flags().String("content", "", "Conteúdo original")
	validateCmd.Flags().String("signature", "", "Assinatura a ser validada")
	rootCmd.AddCommand(validateCmd)
}
