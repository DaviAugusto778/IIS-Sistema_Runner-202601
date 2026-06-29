package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version é a versão semântica do CLI, injetada via -ldflags no build do CI.
	Version = "dev"
	// Commit é o SHA curto do HEAD, injetado via -ldflags no build do CI.
	Commit = "none"
	// Date é o timestamp UTC do build em RFC3339, injetado via -ldflags no CI.
	Date = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão atual do CLI",
	Long:  `Exibe a versão, commit e data de build do CLI assinatura. Os valores são injetados via -ldflags no build do CI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("assinatura %s (commit %s, build %s)\n", Version, Commit, Date)
	},
}

func init() {
	// Habilita também a flag --version (criterio I), com saída idêntica à do
	// subcomando version: "assinatura <versao> (commit <sha>, build <data>)".
	rootCmd.Version = fmt.Sprintf("%s (commit %s, build %s)", Version, Commit, Date)
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	rootCmd.AddCommand(versionCmd)
}
