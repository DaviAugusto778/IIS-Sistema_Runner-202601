//go:build !windows

package server

import (
	"os/exec"
	"syscall"
)

// Detach coloca o filho em um novo grupo de processos, evitando que o
// SIGHUP enviado ao CLI ao terminar se propague ao JVM.
func Detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
