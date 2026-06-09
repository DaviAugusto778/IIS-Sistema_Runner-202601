package main

import (
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Baixa (se preciso) e inicia o Simulador em background",
	RunE: func(_ *cobra.Command, _ []string) error {
		// Implementação completa virá em 5c-C:
		//   1) ler release.json (internal/release)
		//   2) garantir simulador.jar local (internal/release.DownloadArtifact)
		//   3) verificar porta 8443 livre
		//   4) spawnar java -jar simulador.jar
		//   5) gravar runtime.State (PID, porta)
		//   6) polling em GET /api/info até "ready"
		return errNotImplemented
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
