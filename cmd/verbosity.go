package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Verbosidade global (criterio I): --quiet silencia o progresso/diagnóstico,
// mantendo o resultado e os erros; --verbose adiciona detalhe. São mutuamente
// exclusivos. O contrato de saída (JSON do resultado em stdout) é preservado em
// qualquer nível — apenas o ruído de diagnóstico é afetado.
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

// progressWriter é o destino das mensagens de progresso/diagnóstico (stderr),
// ou io.Discard sob --quiet.
func progressWriter() io.Writer {
	if optQuiet {
		return io.Discard
	}
	return os.Stderr
}

// cmdOut é o destino das mensagens humanas de um comando (ex.: start/stop),
// silenciado sob --quiet — o desfecho permanece legível pelo exit code.
func cmdOut(cmd *cobra.Command) io.Writer {
	if optQuiet {
		return io.Discard
	}
	return cmd.OutOrStdout()
}

// tracef emite uma linha de detalhe em stderr apenas sob --verbose.
func tracef(format string, a ...any) {
	if optVerbose && !optQuiet {
		fmt.Fprintf(os.Stderr, "verbose: "+format+"\n", a...)
	}
}
