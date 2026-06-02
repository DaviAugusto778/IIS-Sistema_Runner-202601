package release

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ErrChecksumMismatch sinaliza que o conteúdo baixado não corresponde
// ao SHA256 declarado pelo chamador.
var ErrChecksumMismatch = errors.New("release: SHA256 do arquivo nao confere com o esperado")

// DownloadArtifact baixa art.URL para dest verificando que art.SHA256
// confere. Wrapper conveniente sobre Download.
func DownloadArtifact(ctx context.Context, client *http.Client, art Artifact, dest string) error {
	return Download(ctx, client, art.URL, dest, art.SHA256)
}

// Download faz GET em url e escreve o corpo em dest. Quando
// expectedSHA256 é não-vazio, valida o hash dos bytes baixados.
//
// A escrita é atómica: o conteúdo vai primeiro para dest+".tmp" e só é
// renomeado para dest se o hash bater e o fechamento do arquivo for
// limpo. Em qualquer falha, o temporário é removido — dest nunca
// existirá em estado parcial.
//
// Se client for nil, http.DefaultClient é usado. O cancelamento e o
// timeout são controlados pelo ctx — a função não impõe timeout
// próprio.
func Download(ctx context.Context, client *http.Client, url, dest, expectedSHA256 string) error {
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("release: criar request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("release: GET %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("release: GET %s: status %d", url, resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("release: criar diretorio destino: %w", err)
	}

	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("release: criar arquivo temporario: %w", err)
	}

	hasher := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(f, hasher), resp.Body)
	closeErr := f.Close()

	if copyErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("release: escrever %s: %w", tmp, copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("release: fechar %s: %w", tmp, closeErr)
	}

	if expectedSHA256 != "" {
		got := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(got, expectedSHA256) {
			_ = os.Remove(tmp)
			return fmt.Errorf("%w: esperado %s, obtido %s", ErrChecksumMismatch, expectedSHA256, got)
		}
	}

	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("release: renomear %s para %s: %w", tmp, dest, err)
	}
	return nil
}
