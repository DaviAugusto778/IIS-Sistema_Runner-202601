package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd representa o comando base quando chamado sem nenhum subcomando
var rootCmd = &cobra.Command{
	Use:   "assinatura",
	Short: "CLI para gerenciamento de assinaturas digitais via HubSaúde",
	Long: `O Sistema Runner (CLI assinatura) facilita o acesso à funcionalidade 
de execução de aplicações Java para criação e validação de assinaturas digitais,
abstraindo a complexidade de configuração do ambiente.`,
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adiciona todos os comandos filhos ao comando raiz e seta as flags apropriadas.
// É chamado pelo main.main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Aqui você pode definir flags globais que valerão para todos os subcomandos.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "arquivo de configuração")
}
