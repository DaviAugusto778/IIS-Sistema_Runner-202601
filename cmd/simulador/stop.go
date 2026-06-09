package main

import (
	"errors"

	"github.com/spf13/cobra"
)

var errNotImplemented = errors.New("comando ainda nao implementado nesta iteracao")

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Encerra a instância em execução (via POST /shutdown)",
	RunE: func(_ *cobra.Command, _ []string) error {
		// Implementação completa virá em 5c-B (POST /shutdown + cleanup state).
		return errNotImplemented
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
