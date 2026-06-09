//go:build !windows

package runtime

import (
	"os"
	"syscall"
)

// probeAlive em Unix usa o "null signal" (sinal 0) que apenas verifica
// se o processo existe e se temos permissão para sinalizá-lo, sem
// entregar nenhum sinal de fato.
func probeAlive(p *os.Process) bool {
	return p.Signal(syscall.Signal(0)) == nil
}
