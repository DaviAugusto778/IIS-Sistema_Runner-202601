package invoker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// httpClient é a indireção usada pelo InvokeHTTP; substituível em testes.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// InvokeHTTP envia a operação op ("sign" ou "validate") para o assinador em
// modo servidor via POST http://localhost:<port>/<op> com o payload em JSON,
// e devolve um Result no mesmo formato do modo local (US-01.6).
//
// Mapeamento de exit code (contrato do ADR-0003):
//   - corpo com "valid":true  → ExitCode 0;
//   - corpo com "valid":false → ExitCode 1 (falha de domínio, inclui 400 de
//     validação);
//   - resposta sem JSON reconhecível → ExitCode 1.
//
// Falhas de transporte (conexão recusada, timeout) retornam erro — o caller
// trata como falha de execução (exit 2 ou fallback para o modo local).
func InvokeHTTP(port int, op string, payload map[string]string) (*Result, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("serializar payload: %w", err)
	}

	url := fmt.Sprintf("http://localhost:%d/%s", port, op)
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ler resposta de %s: %w", url, err)
	}

	return &Result{Stdout: respBody, ExitCode: exitCodeFor(respBody)}, nil
}

// exitCodeFor deriva o exit code do campo "valid" do corpo JSON. Ausência ou
// valor falso resulta em 1, preservando a semântica do modo local.
func exitCodeFor(respBody []byte) int {
	var parsed struct {
		Valid bool `json:"valid"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil || !parsed.Valid {
		return 1
	}
	return 0
}
