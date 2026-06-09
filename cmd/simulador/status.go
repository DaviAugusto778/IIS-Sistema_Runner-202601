package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/runtime"
	"github.com/spf13/cobra"
)

const stateName = "simulador"

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Mostra se o Simulador está em execução",
	Long: `Lê o estado persistido em ~/.hubsaude/simulador.json e verifica
se o processo correspondente ainda está vivo no sistema operacional.

Exit codes:
  0  Simulador em execução
  1  Simulador parado (sem state file ou processo morto)
  2  Erro ao consultar o estado`,
	RunE: func(_ *cobra.Command, _ []string) error {
		st, err := runtime.Load(stateName)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("Simulador parado (nenhum state file)")
			os.Exit(1)
		}
		if err != nil {
			return err
		}
		if !runtime.IsAlive(st.PID) {
			fmt.Printf("Simulador parado (PID %d nao esta mais vivo; state orfao)\n", st.PID)
			os.Exit(1)
		}
		fmt.Printf("Simulador em execucao: PID=%d porta=%d\n", st.PID, st.Port)
		// Nota: o health check via GET /api/info sera adicionado em 5c-B.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
