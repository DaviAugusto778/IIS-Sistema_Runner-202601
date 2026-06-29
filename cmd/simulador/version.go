package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version é a versão semântica do CLI, injetada via -ldflags no CI.
	Version = "dev"
	// Commit é o SHA curto do HEAD, injetado via -ldflags no CI.
	Commit = "none"
	// Date é o timestamp UTC do build em RFC3339, injetado via -ldflags no CI.
	Date = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão do CLI simulador",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("simulador %s (commit %s, build %s)\n", Version, Commit, Date)
	},
}

func init() {
	// Habilita também a flag --version (criterio I), com saída idêntica à do
	// subcomando version: "simulador <versao> (commit <sha>, build <data>)".
	rootCmd.Version = fmt.Sprintf("%s (commit %s, build %s)", Version, Commit, Date)
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	rootCmd.AddCommand(versionCmd)
}
