package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "simulador",
	Short: "CLI para gerenciar o ciclo de vida do Simulador HubSaúde",
	Long: `O simulador é um JAR Java distribuído pela disciplina via release.json.
Este CLI baixa o JAR sob demanda, inicia o processo em background na
porta 8443, e oferece comandos para inspecionar (status) e encerrar
(stop) a instância em execução.

O estado do processo (PID, porta) é persistido em ~/.hubsaude/ entre
invocações.

Exemplos:
  simulador start         # baixa o JAR (se preciso) e inicia
  simulador status        # mostra se está em execução
  simulador stop          # encerra a instância em execução`,
}
