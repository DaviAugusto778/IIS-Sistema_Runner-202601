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

// DefaultManifestURL é a URL canônica do release.json mantido pela
// disciplina (kyriosdata/runner). Pode ser sobrescrito pelo CLI via
// flag --source para apontar para forks ou releases experimentais.
const DefaultManifestURL = "https://raw.githubusercontent.com/kyriosdata/runner/main/release.json"

// FetchManifest faz GET em url e parseia o release.json apontado.
// Erros HTTP (status != 2xx) ou de schema sao propagados.
func FetchManifest(ctx context.Context, client *http.Client, url string) (*Manifest, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("release: criar request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("release: GET %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("release: GET %s: status %d", url, resp.StatusCode)
	}
	return Parse(resp.Body)
}

// CacheDir devolve ~/.hubsaude/ criando-o se necessário. É o diretório
// canônico para artefatos baixados e estado dos CLIs.
func CacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("release: home do usuario: %w", err)
	}
	dir := filepath.Join(home, ".hubsaude")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("release: criar %s: %w", dir, err)
	}
	return dir, nil
}

// ArtifactPath devolve o caminho local de cache para o artefato
// <name> na versão <version> (ex.: ~/.hubsaude/simulador-0.1.7.jar).
// O nome versionado permite manter múltiplas versões em cache e detectar
// atualizações sem comparar bytes.
func ArtifactPath(name, version string) (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%s.jar", name, version)), nil
}

// Verify confere se o SHA256 do arquivo em path bate com expected (hex).
// Retorna nil em sucesso; ErrChecksumMismatch quando o conteudo difere;
// um erro que satisfaz errors.Is(err, os.ErrNotExist) quando o arquivo
// nao existe; demais erros de IO sao propagados.
func Verify(path, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("release: ler %s: %w", path, err)
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, expected) {
		return fmt.Errorf("%w: esperado %s, obtido %s", ErrChecksumMismatch, expected, got)
	}
	return nil
}

// Ensure garante que art esteja em dest com checksum válido. Se dest já
// existe e o SHA256 confere, é no-op (cache hit). Caso contrário, baixa
// via DownloadArtifact (que escreve atomicamente e revalida o hash).
func Ensure(ctx context.Context, client *http.Client, art Artifact, dest string) error {
	err := Verify(dest, art.SHA256)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, os.ErrNotExist), errors.Is(err, ErrChecksumMismatch):
		// fall through: precisa baixar
	default:
		return err
	}
	return DownloadArtifact(ctx, client, art, dest)
}

// EnsureSimulador é a operação ponta-a-ponta de US-03.4: busca o
// release.json em manifestURL, calcula o caminho de cache para o
// simulador.jar na versão declarada, baixa se ausente ou desatualizado,
// e devolve o caminho local pronto para `java -jar`.
//
// Se manifestURL for vazio, usa DefaultManifestURL (repo da disciplina).
func EnsureSimulador(ctx context.Context, client *http.Client, manifestURL string) (string, error) {
	if manifestURL == "" {
		manifestURL = DefaultManifestURL
	}
	m, err := FetchManifest(ctx, client, manifestURL)
	if err != nil {
		return "", err
	}
	dest, err := ArtifactPath("simulador", m.Simulador.Version)
	if err != nil {
		return "", err
	}
	if err := Ensure(ctx, client, m.Simulador, dest); err != nil {
		return "", err
	}
	return dest, nil
}
