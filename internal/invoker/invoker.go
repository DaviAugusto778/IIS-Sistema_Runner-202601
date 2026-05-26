package invoker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Response struct {
	Signature string `json:"signature"`
	Valid      bool   `json:"valid"`
	Message    string `json:"message"`
}

func Invoke(args ...string) (*Response, error) {
	jar, err := jarPath()
	if err != nil {
		return nil, err
	}

	cmdArgs := append([]string{"-jar", jar}, args...)
	out, execErr := exec.Command("java", cmdArgs...).Output()

	if execErr != nil {
		if exitErr, ok := execErr.(*exec.ExitError); ok {
			var resp Response
			if len(out) > 0 && json.Unmarshal(out, &resp) == nil {
				return &resp, nil
			}
			if json.Unmarshal(exitErr.Stderr, &resp) == nil {
				return &resp, nil
			}
			return nil, fmt.Errorf("assinador.jar falhou: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("erro ao executar java: %w", execErr)
	}

	var resp Response
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("resposta inválida do assinador.jar: %w", err)
	}
	return &resp, nil
}

func jarPath() (string, error) {
	if p := os.Getenv("ASSINADOR_JAR"); p != "" {
		return p, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("não foi possível localizar assinador.jar: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), "assinador.jar"), nil
}
