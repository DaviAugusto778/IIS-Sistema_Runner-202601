package invoker

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// Invoke executa o assinador.jar com os argumentos informados.
//
// Retorna um *Result preenchido sempre que o java foi de fato executado
// (mesmo com exit code != 0). Devolve erro apenas quando algo impede a
// execução: jar ausente, java fora do PATH, falha de I/O do processo.
func Invoke(args ...string) (*Result, error) {
	jar, err := jarPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(jar); err != nil {
		return nil, fmt.Errorf("assinador.jar nao encontrado em %q: %w", jar, err)
	}

	cmdArgs := append([]string{"-jar", jar}, args...)
	var stdout, stderr bytes.Buffer
	cmd := execCommand("java", cmdArgs...)
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

// jarPath resolve o caminho do assinador.jar.
//
// Ordem de busca:
//  1. variável de ambiente ASSINADOR_JAR (override explícito);
//  2. diretório do próprio executável (atende E1: independente do CWD).
func jarPath() (string, error) {
	if p := os.Getenv("ASSINADOR_JAR"); p != "" {
		return p, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("nao foi possivel localizar o diretorio do executavel: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), "assinador.jar"), nil
}
