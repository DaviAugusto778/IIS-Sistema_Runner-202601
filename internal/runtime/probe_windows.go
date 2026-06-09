//go:build windows

package runtime

import (
	"os"
	"syscall"
)

// probeAlive em Windows abre um handle limitado ao processo e consulta
// o exit code. STILL_ACTIVE (259) indica que o processo continua em
// execução.
func probeAlive(p *os.Process) bool {
	const processQueryLimitedInformation = 0x1000
	const stillActive = 259

	h, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(p.Pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)

	var code uint32
	if err := syscall.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	return code == stillActive
}
