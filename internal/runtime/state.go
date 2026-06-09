// Package runtime gerencia o estado persistente de processos em
// background lançados pelos CLIs Runner (simulador, assinatura no
// modo servidor).
//
// O estado vive em ~/.hubsaude/<name>.json e captura o mínimo
// necessário para que invocações subsequentes do CLI consigam
// inspecionar, controlar ou liberar o processo: PID e porta.
//
// O pacote oferece um teste de vivacidade cross-platform (IsAlive)
// para distinguir um arquivo de estado órfão (processo já morto) de
// um processo ainda em execução.
package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// State captura o estado mínimo de um processo em background.
type State struct {
	PID  int `json:"pid"`
	Port int `json:"port"`
}

// HomeDir devolve o caminho ~/.hubsaude, criando-o se necessário.
func HomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("runtime: home do usuario: %w", err)
	}
	dir := filepath.Join(home, ".hubsaude")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("runtime: criar %s: %w", dir, err)
	}
	return dir, nil
}

// StatePath devolve o caminho do arquivo de estado de name
// (ex.: "simulador" → ~/.hubsaude/simulador.json).
func StatePath(name string) (string, error) {
	dir, err := HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

// Save persiste s no arquivo de estado de name.
func Save(name string, s State) error {
	p, err := StatePath(name)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("runtime: serializar estado: %w", err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		return fmt.Errorf("runtime: escrever %s: %w", p, err)
	}
	return nil
}

// Load lê o estado persistido. Quando não há arquivo, devolve um erro
// que satisfaz errors.Is(err, os.ErrNotExist) — o caller decide como
// tratar (ex.: "simulador nunca foi iniciado").
func Load(name string) (*State, error) {
	p, err := StatePath(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("runtime: parse %s: %w", p, err)
	}
	return &s, nil
}

// Delete remove o arquivo de estado. Não é erro se ele já não existia.
func Delete(name string) error {
	p, err := StatePath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("runtime: remover %s: %w", p, err)
	}
	return nil
}

// IsAlive verifica se o processo com o PID indicado ainda está em
// execução. A implementação é cross-platform (probe_unix.go,
// probe_windows.go). PIDs <= 0 sempre retornam false.
func IsAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return probeAlive(proc)
}
