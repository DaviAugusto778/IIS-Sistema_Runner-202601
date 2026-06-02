package release

import (
	"errors"
	"strings"
	"testing"
)

const validJSON = `{
  "validador": {
    "url": "https://example.com/v.jar",
    "version": "1.0.0",
    "tag": "v1.0.0",
    "sha256": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "simulador": {
    "url": "https://example.com/s.jar",
    "version": "0.1.0",
    "tag": "sim-v0.1.0",
    "sha256": "1111111111111111111111111111111111111111111111111111111111111111"
  },
  "jre": {
    "windows_x64": "https://j/win-x64",
    "windows_arm64": "https://j/win-arm64",
    "linux_x64": "https://j/lin-x64",
    "linux_arm64": "https://j/lin-arm64",
    "mac_x64": "https://j/mac-x64",
    "mac_arm64": "https://j/mac-arm64"
  }
}`

func TestParse_Fixture(t *testing.T) {
	m, err := ParseFile("testdata/release.json")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if m.Validador.Version != "0.1.8" {
		t.Errorf("Validador.Version = %q, quer 0.1.8", m.Validador.Version)
	}
	if m.Simulador.Tag != "hubsaude-simulador-v0.1.7" {
		t.Errorf("Simulador.Tag = %q, quer hubsaude-simulador-v0.1.7", m.Simulador.Tag)
	}
	if !strings.HasPrefix(m.JRE.LinuxX64, "https://api.adoptium.net/") {
		t.Errorf("JRE.LinuxX64 = %q, quer prefixo Adoptium", m.JRE.LinuxX64)
	}
}

func TestParse_Inline(t *testing.T) {
	m, err := Parse(strings.NewReader(validJSON))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if m.Validador.URL != "https://example.com/v.jar" {
		t.Errorf("Validador.URL = %q", m.Validador.URL)
	}
}

func TestForPlatform_Supported(t *testing.T) {
	m, err := Parse(strings.NewReader(validJSON))
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	cases := []struct {
		goos, goarch, want string
	}{
		{"windows", "amd64", "https://j/win-x64"},
		{"windows", "arm64", "https://j/win-arm64"},
		{"linux", "amd64", "https://j/lin-x64"},
		{"linux", "arm64", "https://j/lin-arm64"},
		{"darwin", "amd64", "https://j/mac-x64"},
		{"darwin", "arm64", "https://j/mac-arm64"},
	}
	for _, tc := range cases {
		got, err := m.JRE.ForPlatform(tc.goos, tc.goarch)
		if err != nil {
			t.Errorf("%s/%s: erro inesperado: %v", tc.goos, tc.goarch, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s/%s: got %q, want %q", tc.goos, tc.goarch, got, tc.want)
		}
	}
}

func TestForPlatform_Unsupported(t *testing.T) {
	m, err := Parse(strings.NewReader(validJSON))
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, err = m.JRE.ForPlatform("freebsd", "amd64")
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Errorf("err = %v, quer ErrUnsupportedPlatform", err)
	}
}

func TestForPlatform_EmptyURLIsUnsupported(t *testing.T) {
	// Mesmo um par mapeado deve falhar se a URL estiver vazia: protege
	// contra manifestos que ainda não declararam aquela plataforma.
	jre := JREEndpoints{LinuxX64: ""}
	_, err := jre.ForPlatform("linux", "amd64")
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Errorf("err = %v, quer ErrUnsupportedPlatform", err)
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := Parse(strings.NewReader("not json"))
	if err == nil {
		t.Fatal("esperava erro de parse")
	}
}

func TestParse_UnknownFieldRejected(t *testing.T) {
	body := `{"validador":{},"simulador":{},"jre":{},"extra":"x"}`
	_, err := Parse(strings.NewReader(body))
	if err == nil {
		t.Fatal("esperava erro com campo desconhecido")
	}
}

func TestParse_MissingSHA256(t *testing.T) {
	body := strings.Replace(validJSON,
		`"sha256": "0000000000000000000000000000000000000000000000000000000000000000"`,
		`"sha256": ""`, 1)
	_, err := Parse(strings.NewReader(body))
	if err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Errorf("err = %v, quer mensagem mencionando sha256", err)
	}
}

func TestParse_InvalidSHA256(t *testing.T) {
	body := strings.Replace(validJSON,
		`"sha256": "0000000000000000000000000000000000000000000000000000000000000000"`,
		`"sha256": "xyz"`, 1)
	_, err := Parse(strings.NewReader(body))
	if err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Errorf("err = %v, quer mensagem mencionando sha256", err)
	}
}

func TestParse_EmptyJRE(t *testing.T) {
	body := strings.Replace(validJSON,
		`"jre": {
    "windows_x64": "https://j/win-x64",
    "windows_arm64": "https://j/win-arm64",
    "linux_x64": "https://j/lin-x64",
    "linux_arm64": "https://j/lin-arm64",
    "mac_x64": "https://j/mac-x64",
    "mac_arm64": "https://j/mac-arm64"
  }`,
		`"jre": {}`, 1)
	_, err := Parse(strings.NewReader(body))
	if err == nil || !strings.Contains(err.Error(), "jre") {
		t.Errorf("err = %v, quer mencionar jre", err)
	}
}
