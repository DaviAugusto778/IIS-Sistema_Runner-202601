package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Mostra se o assinador está em execução no modo servidor",
	Long: `Consulta ~/.hubsaude/assinador.json e verifica o processo registrado
(PID vivo) confirmando com um health check HTTP em GET /health.

Exit codes:
  0  assinador em execução e respondendo
  1  assinador parado, ou registrado mas sem responder`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		code, err := runStatus(cmd.OutOrStdout())
		if err != nil {
			return err
		}
		if code != 0 {
			os.Exit(code)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

// runStatus imprime o estado da instância e devolve o exit code apropriado
// (0 em execução; 1 parado/sem resposta). Erros de I/O ao consultar o
// estado são retornados (tratados como exit 1 pelo Cobra).
func runStatus(out io.Writer) (int, error) {
	status, st, err := detectInstance(out)
	if err != nil {
		return 1, err
	}
	switch status {
	case statusRunning:
		_, _ = fmt.Fprintf(out, "Assinador em execucao: PID=%d porta=%d (respondendo em %s)\n",
			st.PID, st.Port, healthPath)
		return 0, nil
	case statusUnresponsive:
		_, _ = fmt.Fprintf(out, "Assinador registrado (PID=%d porta=%d) mas nao responde em %s\n",
			st.PID, st.Port, healthPath)
		return 1, nil
	default:
		_, _ = fmt.Fprintln(out, "Assinador parado (nenhuma instancia ativa)")
		return 1, nil
	}
}
