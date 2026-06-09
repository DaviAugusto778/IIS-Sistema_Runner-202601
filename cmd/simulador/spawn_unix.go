//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// detachProcess coloca o filho em um novo grupo de processos, evitando
// que o SIGHUP enviado ao CLI ao terminar se propague ao JVM.
func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
