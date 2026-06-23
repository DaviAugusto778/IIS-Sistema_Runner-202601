package jdk

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/DaviAugusto778/IIS-Sistema_Runner-202601/internal/release"
)

// adoptiumAssetsBase é o endpoint da API Eclipse Adoptium que devolve o
// binário GA mais recente do JDK para uma plataforma, já com o checksum
// SHA256 do pacote — usado para verificar a integridade do download.
// É var (não const) para que os testes apontem para um servidor fake.
var adoptiumAssetsBase = "https://api.adoptium.net/v3/assets/latest/21/hotspot"

// asset descreve o pacote selecionado para download.
type asset struct {
	URL      string
	SHA256   string
	Filename string
}

// resposta parcial da API Adoptium (apenas os campos que consumimos).
type adoptiumResponse struct {
	Binary struct {
		Package struct {
			Link     string `json:"link"`
			Checksum string `json:"checksum"`
			Name     string `json:"name"`
		} `json:"package"`
	} `json:"binary"`
}

// osToken mapeia GOOS para o token de sistema operacional da API Adoptium.
func osToken(goos string) (string, error) {
	switch goos {
	case "windows":
		return "windows", nil
	case "linux":
		return "linux", nil
	case "darwin":
		return "mac", nil
	default:
		return "", fmt.Errorf("jdk: sistema operacional sem distribuicao suportada: %s", goos)
	}
}

// archToken mapeia GOARCH para o token de arquitetura da API Adoptium.
func archToken(goarch string) (string, error) {
	switch goarch {
	case "amd64":
		return "x64", nil
	case "arm64":
		return "aarch64", nil
	default:
		return "", fmt.Errorf("jdk: arquitetura sem distribuicao suportada: %s", goarch)
	}
}

// assetsURL monta a URL de consulta da API para a plataforma atual.
func assetsURL() (string, error) {
	osTok, err := osToken(runtimeGOOS)
	if err != nil {
		return "", err
	}
	archTok, err := archToken(runtimeGOARCH)
	if err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("architecture", archTok)
	q.Set("image_type", "jdk")
	q.Set("os", osTok)
	q.Set("vendor", "eclipse")
	return adoptiumAssetsBase + "?" + q.Encode(), nil
}

// resolveAsset consulta a API Adoptium e devolve o pacote do JDK 21 para a
// plataforma corrente, com URL, checksum SHA256 e nome do arquivo.
func resolveAsset(ctx context.Context, client *http.Client) (asset, error) {
	if client == nil {
		client = http.DefaultClient
	}
	u, err := assetsURL()
	if err != nil {
		return asset{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return asset{}, fmt.Errorf("jdk: criar request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return asset{}, fmt.Errorf("jdk: consultar Adoptium: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return asset{}, fmt.Errorf("jdk: Adoptium respondeu status %d", resp.StatusCode)
	}

	var assets []adoptiumResponse
	if err := json.NewDecoder(resp.Body).Decode(&assets); err != nil {
		return asset{}, fmt.Errorf("jdk: decodificar resposta Adoptium: %w", err)
	}
	if len(assets) == 0 || assets[0].Binary.Package.Link == "" {
		return asset{}, fmt.Errorf("jdk: Adoptium nao retornou pacote para %s/%s", runtimeGOOS, runtimeGOARCH)
	}
	pkg := assets[0].Binary.Package
	return asset{URL: pkg.Link, SHA256: pkg.Checksum, Filename: pkg.Name}, nil
}

// Ensure devolve o caminho de um launcher de JDK 21 pronto para uso,
// provisionando-o automaticamente quando ausente.
//
// É a operação ponta-a-ponta de US-04.1: detecta um JDK existente (no-op
// quando há) ou baixa e extrai a distribuição adequada em ~/.hubsaude/jdk/.
func Ensure(ctx context.Context, client *http.Client) (string, error) {
	if p, err := Detect(); err == nil {
		return p, nil
	}
	return Provision(ctx, client)
}

// Provision baixa, verifica e extrai o JDK 21 para ~/.hubsaude/jdk/,
// devolvendo o caminho do launcher. Sobrescreve um diretório gerenciado
// pré-existente (ex.: instalação corrompida ou incompleta).
func Provision(ctx context.Context, client *http.Client) (string, error) {
	a, err := resolveAsset(ctx, client)
	if err != nil {
		return "", err
	}

	cacheDir, err := release.CacheDir()
	if err != nil {
		return "", err
	}

	archive := filepath.Join(cacheDir, "jdk-download"+archiveExt(a))
	if err := release.Download(ctx, client, a.URL, archive, a.SHA256); err != nil {
		return "", fmt.Errorf("jdk: baixar distribuicao: %w", err)
	}
	defer func() { _ = os.Remove(archive) }()

	staging := filepath.Join(cacheDir, "jdk.staging")
	if err := os.RemoveAll(staging); err != nil {
		return "", fmt.Errorf("jdk: limpar staging: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	if err := extractArchive(archive, staging); err != nil {
		return "", err
	}

	home, err := findJavaHome(staging)
	if err != nil {
		return "", err
	}

	managed, err := ManagedDir()
	if err != nil {
		return "", err
	}
	if err := os.RemoveAll(managed); err != nil {
		return "", fmt.Errorf("jdk: remover JDK gerenciado anterior: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(managed), 0o755); err != nil {
		return "", fmt.Errorf("jdk: criar diretorio gerenciado: %w", err)
	}
	if err := os.Rename(home, managed); err != nil {
		return "", fmt.Errorf("jdk: instalar JDK em %s: %w", managed, err)
	}

	java, err := ManagedJava()
	if err != nil {
		return "", err
	}
	if !meetsRequirement(java) {
		return "", fmt.Errorf("jdk: JDK provisionado em %s nao satisfaz versao %d", managed, RequiredMajor)
	}
	return java, nil
}

// archiveExt deduz a extensão do arquivo baixado a partir do nome/URL do
// pacote: .zip no Windows, .tar.gz nas demais plataformas.
func archiveExt(a asset) string {
	name := a.Filename
	if name == "" {
		name = a.URL
	}
	if strings.HasSuffix(strings.ToLower(name), ".zip") {
		return ".zip"
	}
	return ".tar.gz"
}

// extractArchive extrai archive para dest, escolhendo o formato pelo sufixo.
func extractArchive(archive, dest string) error {
	if strings.HasSuffix(strings.ToLower(archive), ".zip") {
		return extractZip(archive, dest)
	}
	return extractTarGz(archive, dest)
}

// extractTarGz extrai um .tar.gz preservando os bits de permissão (para que
// bin/java permaneça executável em sistemas Unix).
func extractTarGz(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return fmt.Errorf("jdk: abrir %s: %w", archive, err)
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("jdk: gzip %s: %w", archive, err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("jdk: ler tar: %w", err)
		}
		target, err := safeJoin(dest, hdr.Name)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("jdk: criar diretorio %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := writeFile(target, tr, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("jdk: criar diretorio %s: %w", filepath.Dir(target), err)
			}
			_ = os.Remove(target)
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return fmt.Errorf("jdk: criar symlink %s: %w", target, err)
			}
		}
	}
}

// extractZip extrai um .zip preservando os bits de permissão.
func extractZip(archive, dest string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return fmt.Errorf("jdk: abrir zip %s: %w", archive, err)
	}
	defer func() { _ = r.Close() }()

	for _, entry := range r.File {
		target, err := safeJoin(dest, entry.Name)
		if err != nil {
			return err
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("jdk: criar diretorio %s: %w", target, err)
			}
			continue
		}
		rc, err := entry.Open()
		if err != nil {
			return fmt.Errorf("jdk: abrir entrada %s: %w", entry.Name, err)
		}
		err = writeFile(target, rc, entry.Mode())
		_ = rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// writeFile grava o conteúdo de r em path com o modo indicado, criando os
// diretórios intermediários.
func writeFile(path string, r io.Reader, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("jdk: criar diretorio %s: %w", filepath.Dir(path), err)
	}
	if mode == 0 {
		mode = 0o644
	}
	out, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
	if err != nil {
		return fmt.Errorf("jdk: criar arquivo %s: %w", path, err)
	}
	_, copyErr := io.Copy(out, r)
	closeErr := out.Close()
	if copyErr != nil {
		return fmt.Errorf("jdk: escrever %s: %w", path, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("jdk: fechar %s: %w", path, closeErr)
	}
	return nil
}

// safeJoin junta dest e name rejeitando caminhos que escapem de dest
// (proteção contra zip/tar slip).
func safeJoin(dest, name string) (string, error) {
	clean := filepath.Clean(filepath.Join(dest, name))
	prefix := filepath.Clean(dest) + string(os.PathSeparator)
	if clean != filepath.Clean(dest) && !strings.HasPrefix(clean, prefix) {
		return "", fmt.Errorf("jdk: caminho invalido no arquivo: %s", name)
	}
	return clean, nil
}

// findJavaHome localiza, dentro de root, o diretório "home" do JDK — aquele
// que contém bin/java[.exe]. Necessário porque as distribuições trazem um
// diretório raiz versionado (e, no macOS, Contents/Home).
func findJavaHome(root string) (string, error) {
	exe := javaExeName()
	var home string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != exe {
			return nil
		}
		if filepath.Base(filepath.Dir(path)) == "bin" {
			home = filepath.Dir(filepath.Dir(path))
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("jdk: localizar home do JDK extraido: %w", err)
	}
	if home == "" {
		return "", fmt.Errorf("jdk: bin/%s nao encontrado na distribuicao extraida", exe)
	}
	return home, nil
}
