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
	rootCmd.AddCommand(versionCmd)
}
