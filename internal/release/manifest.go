// Package release lê e interpreta o release.json descrito na
// especificacao@4d7d40f. O manifesto declara, em um único lugar,
// onde buscar os artefatos (validador, simulador) e os endpoints da
// JRE Adoptium para cada combinação OS/arquitetura suportada.
//
// O parser apenas valida a estrutura. A verificação de integridade
// (SHA256) e o download dos artefatos são responsabilidade de outros
// componentes deste pacote (a serem adicionados em iterações futuras).
package release

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
)

// Manifest é a representação tipada do release.json.
type Manifest struct {
	Validador Artifact     `json:"validador"`
	Simulador Artifact     `json:"simulador"`
	JRE       JREEndpoints `json:"jre"`
}

// Artifact descreve um artefato distribuível (.jar) hospedado em uma
// release do GitHub.
type Artifact struct {
	URL     string `json:"url"`
	Version string `json:"version"`
	Tag     string `json:"tag"`
	SHA256  string `json:"sha256"`
}

// JREEndpoints lista as URLs da JRE Adoptium para cada par OS/arch.
type JREEndpoints struct {
	WindowsX64   string `json:"windows_x64"`
	WindowsARM64 string `json:"windows_arm64"`
	LinuxX64     string `json:"linux_x64"`
	LinuxARM64   string `json:"linux_arm64"`
	MacX64       string `json:"mac_x64"`
	MacARM64     string `json:"mac_arm64"`
}

// ErrUnsupportedPlatform é devolvido por ForPlatform quando o par
// goos/goarch informado não tem endpoint declarado no manifesto.
var ErrUnsupportedPlatform = errors.New("plataforma sem JRE declarada no manifesto")

var sha256Pattern = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

// Parse lê e valida o manifesto a partir de um io.Reader.
func Parse(r io.Reader) (*Manifest, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var m Manifest
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("release: parse json: %w", err)
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// ParseFile lê e valida o manifesto a partir de um caminho no disco.
func ParseFile(path string) (*Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("release: abrir %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	return Parse(f)
}

// ForPlatform devolve a URL da JRE para o par goos/goarch indicado
// (formato padrão do runtime: "windows"/"linux"/"darwin" e
// "amd64"/"arm64"). Retorna ErrUnsupportedPlatform quando o par não
// está mapeado no manifesto.
func (j JREEndpoints) ForPlatform(goos, goarch string) (string, error) {
	key := goos + "/" + goarch
	var url string
	switch key {
	case "windows/amd64":
		url = j.WindowsX64
	case "windows/arm64":
		url = j.WindowsARM64
	case "linux/amd64":
		url = j.LinuxX64
	case "linux/arm64":
		url = j.LinuxARM64
	case "darwin/amd64":
		url = j.MacX64
	case "darwin/arm64":
		url = j.MacARM64
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedPlatform, key)
	}
	if url == "" {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedPlatform, key)
	}
	return url, nil
}

func (m *Manifest) validate() error {
	if err := m.Validador.validate("validador"); err != nil {
		return err
	}
	if err := m.Simulador.validate("simulador"); err != nil {
		return err
	}
	// JRE: ao menos uma plataforma precisa estar declarada — não
	// exigimos as seis, pois novas plataformas podem ser adicionadas
	// gradualmente no upstream.
	if m.JRE == (JREEndpoints{}) {
		return errors.New("release: bloco 'jre' vazio")
	}
	return nil
}

func (a Artifact) validate(name string) error {
	if a.URL == "" {
		return fmt.Errorf("release: %s.url vazio", name)
	}
	if a.Version == "" {
		return fmt.Errorf("release: %s.version vazio", name)
	}
	if a.Tag == "" {
		return fmt.Errorf("release: %s.tag vazio", name)
	}
	if !sha256Pattern.MatchString(a.SHA256) {
		return fmt.Errorf("release: %s.sha256 nao e hex de 64 chars: %q", name, a.SHA256)
	}
	return nil
}
