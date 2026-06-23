package cmd

import (
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Valida uma assinatura digital simulada",
	Long: `Invoca o assinador.jar para verificar se uma assinatura digital
corresponde ao conteúdo fornecido.

Por padrão usa o modo servidor (ADR-0002): reutiliza uma instância ativa ou
inicia uma automaticamente. Use --local para forçar a invocação direta
'java -jar'.

A validação dos parâmetros é responsabilidade exclusiva do assinador.jar
(autoridade única, conforme criterio E3).

Exemplos:
  assinatura validate --content "ola mundo" --signature "MOCKED_SIGNATURE_BASE64_=="
  assinatura validate --content "ola" --signature "..." --local`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		payload := map[string]string{}
		if cmd.Flags().Changed("content") {
			v, _ := cmd.Flags().GetString("content")
			payload["content"] = v
		}
		if cmd.Flags().Changed("signature") {
			v, _ := cmd.Flags().GetString("signature")
			payload["signature"] = v
		}
		return runOperation(cmd, "validate", payload)
	},
}

func init() {
	validateCmd.Flags().String("content", "", "Conteúdo original")
	validateCmd.Flags().String("signature", "", "Assinatura a ser validada")
	validateCmd.Flags().Bool("local", false, "Força invocação direta 'java -jar' (sem modo servidor)")
	rootCmd.AddCommand(validateCmd)
}
