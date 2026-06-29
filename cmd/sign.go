package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

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

Para assinar interagindo com um dispositivo criptográfico (token/smart card)
via PKCS#11 (US-02.5), use --pkcs11-lib (o PIN vai em --token). A interação
com o dispositivo é por invocação, então --pkcs11-lib implica modo local; para
o modo servidor com dispositivo, configure-o em 'assinatura start --pkcs11-lib'.

Exemplos:
  assinatura sign --content "ola mundo"
  assinatura sign --content "$(cat documento.txt)" --token abc123
  assinatura sign --content "ola" --local
  assinatura sign --content "ola" --token 1234 --pkcs11-lib /usr/lib/softhsm/libsofthsm2.so`,
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
	signCmd.Flags().String("token", "", "Token de autenticação (PIN do dispositivo PKCS#11, quando aplicável)")
	signCmd.Flags().Bool("local", false, "Força invocação direta 'java -jar' (sem modo servidor)")
	signCmd.Flags().String("pkcs11-lib", "", "Caminho da biblioteca PKCS#11 do dispositivo (ativa o modo dispositivo; implica --local)")
	signCmd.Flags().Int("pkcs11-slot", pkcs11SlotUnset, "Slot do dispositivo PKCS#11 (padrão: primeiro slot)")
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
	pkcs11 := pkcs11ArgsFromFlags(cmd)

	// O dispositivo PKCS#11 é configurado por invocação (modo local) ou no
	// servidor (via `start --pkcs11-lib`). Aplicá-lo a um servidor genérico já
	// no ar produziria uma assinatura sem dispositivo — por isso --pkcs11-lib
	// força o modo local nesta chamada.
	if len(pkcs11) > 0 && !local {
		fmt.Fprintln(progressWriter(), "aviso: --pkcs11-lib configura o dispositivo por invocacao; usando modo local "+
			"(para modo servidor com dispositivo, use `assinatura start --pkcs11-lib`)")
		local = true
	}
	if local {
		tracef("operacao=%s modo=local", op)
		return runLocal(op, payload, pkcs11)
	}

	// Progresso (provisão de JDK, startup do servidor) vai para stderr para
	// manter o stdout limpo com apenas o JSON do resultado (ADR-0003);
	// silenciável com --quiet.
	port, err := ensureServer(progressWriter())
	if err != nil {
		fmt.Fprintf(progressWriter(), "aviso: modo servidor indisponivel (%v); usando modo local\n", err)
		return runLocal(op, payload, pkcs11)
	}
	tracef("operacao=%s modo=servidor porta=%d", op, port)

	result, err := invoker.InvokeHTTP(port, op, payload)
	if err != nil {
		fmt.Fprintf(progressWriter(), "aviso: falha na chamada HTTP (%v); usando modo local\n", err)
		return runLocal(op, payload, pkcs11)
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
func runLocal(op string, payload map[string]string, extra []string) error {
	args := localArgs(op, payload, extra)
	tracef("invocando: java -jar assinador.jar %s", strings.Join(args, " "))
	result, err := invoker.Invoke(args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(2)
	}
	emitResult(result)
	return nil
}

// localArgs converte o payload em argumentos de linha de comando para o JAR,
// em ordem estável. Repassa apenas as chaves fornecidas pelo usuário, seguidas
// dos argumentos extras (ex.: --pkcs11-lib/--pkcs11-slot da US-02.5).
func localArgs(op string, payload map[string]string, extra []string) []string {
	args := []string{op}
	for _, k := range []string{"content", "token", "signature"} {
		if v, ok := payload[k]; ok {
			args = append(args, "--"+k, v)
		}
	}
	return append(args, extra...)
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
