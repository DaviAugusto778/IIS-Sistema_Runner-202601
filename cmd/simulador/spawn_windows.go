//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// detachedProcess corresponde ao flag DETACHED_PROCESS da WinAPI
// (CreateProcess). Nao esta exportado em syscall, daí a constante local.
const detachedProcess = 0x00000008

// detachProcess separa o filho do console do CLI e cria um novo grupo
// de processos, permitindo que o CLI termine sem matar o JVM.
func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcess,
		HideWindow:    true,
	}
}
