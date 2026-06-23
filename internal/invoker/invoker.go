// Package invoker executa o assinador.jar como subprocesso, propagando
// stdout, stderr e exit code conforme o contrato CLI<->JAR descrito em
// docs/adr/0003-contrato-cli-jar.md.
package invoker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/jdk"
)

// Result captura a saída de uma invocação ao assinador.jar.
// Stdout transporta o resultado (JSON), Stderr o diagnóstico, ExitCode o
// status devolvido pelo JAR. A separação atende ao critério E1 da
// especificação (docs/criterios.md@4d7d40f).
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// execCommand é uma indireção que permite substituir o launcher do java em
// testes (ver invoker_test.go). Em produção, equivale a exec.Command.
var execCommand = exec.Command

// resolveJavaPath localiza (e, se necessário, provisiona) o launcher do
// JDK 21 a ser usado. Indireção para testes; em produção delega a
// jdk.Ensure, que detecta um JDK existente ou baixa a distribuição
// adequada para ~/.hubsaude/jdk/ (US-04.1).
var resolveJavaPath = func() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	return jdk.Ensure(ctx, nil)
}

// Invoke executa o assinador.jar com os argumentos informados.
//
// Retorna um *Result preenchido sempre que o java foi de fato executado
// (mesmo com exit code != 0). Devolve erro apenas quando algo impede a
// execução: jar ausente, java fora do PATH, falha de I/O do processo.
func Invoke(args ...string) (*Result, error) {
	jar, err := JarPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(jar); err != nil {
		return nil, fmt.Errorf("assinador.jar nao encontrado em %q: %w", jar, err)
	}

	javaBin, err := resolveJavaPath()
	if err != nil {
		return nil, fmt.Errorf("nao foi possivel localizar ou provisionar o JDK 21: %w", err)
	}

	cmdArgs := append([]string{"-jar", jar}, args...)
	var stdout, stderr bytes.Buffer
	cmd := execCommand(javaBin, cmdArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	result := &Result{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: cmd.ProcessState.ExitCode(),
	}

	var exitErr *exec.ExitError
	if runErr != nil && !errors.As(runErr, &exitErr) {
		return result, fmt.Errorf("falha ao invocar java: %w", runErr)
	}
	return result, nil
}

// JarPath resolve o caminho do assinador.jar.
//
// Ordem de busca:
//  1. variável de ambiente ASSINADOR_JAR (override explícito);
//  2. diretório do próprio executável (atende E1: independente do CWD).
func JarPath() (string, error) {
	if p := os.Getenv("ASSINADOR_JAR"); p != "" {
		return p, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("nao foi possivel localizar o diretorio do executavel: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), "assinador.jar"), nil
}
