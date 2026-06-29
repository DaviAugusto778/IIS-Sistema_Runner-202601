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
	"time"
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

// loadAttempts limita os retries de Load diante de condições transitórias de
// concorrência (ver Load). Com o backoff curto, cobre dezenas de ms — folga de
// sobra para um Save (que dura microssegundos) terminar.
const loadAttempts = 20

// Save persiste s no arquivo de estado de name.
//
// Mantém uma escrita simples (os.WriteFile): no Windows, substituir o arquivo
// por rename enquanto um leitor concorrente o mantém aberto falha com sharing
// violation. A integridade sob concorrência ("race no start", criterio G) é
// garantida do lado da leitura: um Save concorrente pode deixar o arquivo
// momentaneamente truncado, e Load relê até obter um JSON íntegro.
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

// Load lê o estado persistido. Quando não há arquivo, devolve um erro que
// satisfaz errors.Is(err, os.ErrNotExist) — o caller decide como tratar
// (ex.: "simulador nunca foi iniciado").
//
// Tolera condições transitórias de um Save concorrente (cenário "race no
// start"): uma abertura recusada (sharing violation no Windows) ou um JSON
// truncado fazem Load reler após um backoff curto, em vez de propagar um erro
// espúrio. ErrNotExist nunca é retried — não é transitório. O erro de parse só
// é devolvido se persistir por todas as tentativas (arquivo de fato corrompido).
func Load(name string) (*State, error) {
	p, err := StatePath(name)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for i := 0; i < loadAttempts; i++ {
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			if errors.Is(rerr, os.ErrNotExist) {
				return nil, rerr
			}
			lastErr = rerr // transitório: ex. sharing violation durante um Save
		} else {
			var s State
			if perr := json.Unmarshal(data, &s); perr == nil {
				return &s, nil
			} else {
				lastErr = fmt.Errorf("runtime: parse %s: %w", p, perr)
			}
		}
		delay := time.Duration(i+1) * time.Millisecond
		if delay > 5*time.Millisecond {
			delay = 5 * time.Millisecond
		}
		time.Sleep(delay)
	}
	return nil, lastErr
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
