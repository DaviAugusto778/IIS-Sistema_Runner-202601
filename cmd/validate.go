package cmd

import (
	"fmt"
	"os"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/invoker"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Valida uma assinatura digital simulada",
	Long:  `Invoca o assinador.jar para verificar se uma assinatura digital é válida para o conteúdo fornecido.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		content, _ := cmd.Flags().GetString("content")
		signature, _ := cmd.Flags().GetString("signature")

		resp, err := invoker.Invoke("validate", "--content", content, "--signature", signature)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			os.Exit(1)
		}

		if resp.Valid {
			fmt.Println("Assinatura válida")
			fmt.Println("Mensagem:", resp.Message)
		} else {
			fmt.Fprintln(os.Stderr, "Assinatura inválida:", resp.Message)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	validateCmd.Flags().String("content", "", "Conteúdo original (obrigatório)")
	validateCmd.Flags().String("signature", "", "Assinatura a ser validada (obrigatório)")
	validateCmd.MarkFlagRequired("content")
	validateCmd.MarkFlagRequired("signature")
	rootCmd.AddCommand(validateCmd)
}
