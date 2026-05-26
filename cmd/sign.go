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
	Long:  `Invoca o assinador.jar para criar uma assinatura digital simulada a partir do conteúdo fornecido.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		content, _ := cmd.Flags().GetString("content")
		token, _ := cmd.Flags().GetString("token")

		invokeArgs := []string{"sign", "--content", content}
		if token != "" {
			invokeArgs = append(invokeArgs, "--token", token)
		}

		resp, err := invoker.Invoke(invokeArgs...)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			os.Exit(1)
		}

		if resp.Valid {
			fmt.Println("Assinatura criada com sucesso")
			fmt.Println("Assinatura:", resp.Signature)
		} else {
			fmt.Fprintln(os.Stderr, "Erro:", resp.Message)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	signCmd.Flags().String("content", "", "Conteúdo a ser assinado (obrigatório)")
	signCmd.Flags().String("token", "", "Token de autenticação (opcional)")
	signCmd.MarkFlagRequired("content")
	rootCmd.AddCommand(signCmd)
}
