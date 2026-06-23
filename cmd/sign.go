package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/invoker"
	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/runtime"
	"github.com/spf13/cobra"
)

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Cria uma assinatura digital simulada",
	Long: `Invoca o assinador.jar para criar uma assinatura digital simulada.

Por padrão usa o modo servidor (ADR-0002): reutiliza uma instância ativa ou
inicia uma automaticamente, eliminando o cold start da JVM a partir da
segunda chamada. Use --local para forçar a invocação direta 'java -jar'.

A validação dos parâmetros é responsabilidade exclusiva do assinador.jar
(autoridade única, conforme criterio E3).

Exemplos:
  assinatura sign --content "ola mundo"
  assinatura sign --content "$(cat documento.txt)" --token abc123
  assinatura sign --content "ola" --local`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		payload := map[string]string{}
		if cmd.Flags().Changed("content") {
			v, _ := cmd.Flags().GetString("content")
			payload["content"] = v
		}
		if cmd.Flags().Changed("token") {
			v, _ := cmd.Flags().GetString("token")
			payload["token"] = v
		}
		return runOperation(cmd, "sign", payload)
	},
}

func init() {
	signCmd.Flags().String("content", "", "Conteúdo a ser assinado")
	signCmd.Flags().String("token", "", "Token de autenticação")
	signCmd.Flags().Bool("local", false, "Força invocação direta 'java -jar' (sem modo servidor)")
	rootCmd.AddCommand(signCmd)
}

// runOperation roteia a operação op ("sign"/"validate") entre o modo local
// (java -jar one-shot) e o modo servidor (POST HTTP), conforme a flag --local.
//
// No modo servidor, garante uma instância no ar (reutiliza ou auto-start,
// ADR-0002) e envia a requisição via HTTP. Se o servidor não puder subir ou a
// chamada HTTP falhar, degrada para o modo local com aviso em stderr — a
// operação sempre produz um resultado.
func runOperation(cmd *cobra.Command, op string, payload map[string]string) error {
	local, _ := cmd.Flags().GetBool("local")
	if local {
		return runLocal(op, payload)
	}

	// Progresso (provisão de JDK, startup do servidor) vai para stderr para
	// manter o stdout limpo com apenas o JSON do resultado (ADR-0003).
	port, err := ensureServer(os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aviso: modo servidor indisponivel (%v); usando modo local\n", err)
		return runLocal(op, payload)
	}

	result, err := invoker.InvokeHTTP(port, op, payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aviso: falha na chamada HTTP (%v); usando modo local\n", err)
		return runLocal(op, payload)
	}
	emitResult(result)
	return nil
}

// ensureServer garante uma instância do assinador no ar e devolve a porta.
// Reutiliza instância ativa ou sobe uma nova (via runStart); mensagens de
// progresso vão para progress.
func ensureServer(progress io.Writer) (int, error) {
	// Auto-start sem timeout de inatividade: a duração do servidor levantado
	// implicitamente é responsabilidade do usuário (ele pode usar `start
	// --timeout` para um servidor autoexpirável).
	if err := runStart(progress, defaultServerPort, readinessTimeout, 0); err != nil {
		return 0, err
	}
	st, err := runtime.Load(stateName)
	if err != nil {
		return 0, err
	}
	return st.Port, nil
}

// runLocal invoca o assinador.jar diretamente (java -jar), repassa
// stdout/stderr e termina com o exit code do JAR. Falhas de execução (jar
// ausente, java fora do PATH) saem com código 2.
func runLocal(op string, payload map[string]string) error {
	result, err := invoker.Invoke(localArgs(op, payload)...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(2)
	}
	emitResult(result)
	return nil
}

// localArgs converte o payload em argumentos de linha de comando para o JAR,
// em ordem estável. Repassa apenas as chaves fornecidas pelo usuário.
func localArgs(op string, payload map[string]string) []string {
	args := []string{op}
	for _, k := range []string{"content", "token", "signature"} {
		if v, ok := payload[k]; ok {
			args = append(args, "--"+k, v)
		}
	}
	return args
}

// emitResult repassa o resultado (stdout/stderr) e encerra o processo com o
// exit code correspondente.
func emitResult(result *invoker.Result) {
	if len(result.Stderr) > 0 {
		_, _ = os.Stderr.Write(result.Stderr)
	}
	if len(result.Stdout) > 0 {
		_, _ = os.Stdout.Write(result.Stdout)
	}
	os.Exit(result.ExitCode)
}
