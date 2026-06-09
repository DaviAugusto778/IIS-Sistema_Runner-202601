// Command simulador é o CLI Runner para gerenciar o ciclo de vida do
// Simulador HubSaúde (start/stop/status).
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
