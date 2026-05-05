package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd representa o comando 'version'
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão atual do CLI",
	Long:  `Exibe a versão atual do executável do CLI assinatura.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Em um cenário real, essa versão pode ser injetada via LDFLAGS durante o build do CI/CD
		fmt.Println("assinatura CLI versão v0.1.0-dev")
	},
}

func init() {
	// Adiciona o comando 'version' como um subcomando do 'rootCmd'
	rootCmd.AddCommand(versionCmd)
}