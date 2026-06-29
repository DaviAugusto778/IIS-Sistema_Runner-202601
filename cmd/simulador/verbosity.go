package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// Verbosidade global (criterio I): --quiet silencia o progresso/diagnóstico,
// mantendo o resultado e os erros; --verbose adiciona detalhe. Mutuamente
// exclusivos.
var (
	optVerbose bool
	optQuiet   bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&optVerbose, "verbose", "v", false,
		"saída detalhada: mostra passos e comandos executados")
	rootCmd.PersistentFlags().BoolVarP(&optQuiet, "quiet", "q", false,
		"silencia o progresso; mantém apenas o resultado e os erros")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		if optVerbose && optQuiet {
			return fmt.Errorf("--verbose e --quiet sao mutuamente exclusivos")
		}
		return nil
	}
}

// cmdOut é o destino das mensagens humanas de um comando, silenciado sob
// --quiet — o desfecho permanece legível pelo exit code.
func cmdOut(cmd *cobra.Command) io.Writer {
	if optQuiet {
		return io.Discard
	}
	return cmd.OutOrStdout()
}
