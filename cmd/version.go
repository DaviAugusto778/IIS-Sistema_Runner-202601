package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
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
	rootCmd.AddCommand(versionCmd)
}
