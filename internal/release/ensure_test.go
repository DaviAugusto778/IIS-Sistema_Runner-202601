package release

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withFakeHome aponta os.UserHomeDir() para um diretorio temporario.
func withFakeHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	return tmp
}

func TestFetchManifestSuccess(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "release.json"))
	if err != nil {
		t.Fatalf("setup ReadFile: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	m, err := FetchManifest(context.Background(), nil, srv.URL)
	if err != nil {
		t.Fatalf("FetchManifest: %v", err)
	}
	if m.Simulador.Version != "0.1.7" {
		t.Errorf("Simulador.Version = %q, quer 0.1.7", m.Simulador.Version)
	}
}

func TestFetchManifestNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "no", http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := FetchManifest(context.Background(), nil, srv.URL); err == nil {
		t.Errorf("queria erro para HTTP 404, got nil")
	}
}

func TestFetchManifestInvalidSchema(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"not": "a manifest"}`))
	}))
	defer srv.Close()

	if _, err := FetchManifest(context.Background(), nil, srv.URL); err == nil {
		t.Errorf("queria erro de schema, got nil")
	}
}

func TestArtifactPathIsVersioned(t *testing.T) {
	home := withFakeHome(t)
	p, err := ArtifactPath("simulador", "0.1.7")
	if err != nil {
		t.Fatalf("ArtifactPath: %v", err)
	}
	want := filepath.Join(home, ".hubsaude", "simulador-0.1.7.jar")
	if p != want {
		t.Errorf("ArtifactPath = %q, quer %q", p, want)
	}
}

func TestVerifySuccess(t *testing.T) {
	home := withFakeHome(t)
	payload := []byte("conteudo do jar")
	p := filepath.Join(home, "ok.bin")
	if err := os.WriteFile(p, payload, 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := Verify(p, sha256Hex(payload)); err != nil {
		t.Errorf("Verify: %v", err)
	}
}

func TestVerifyMismatch(t *testing.T) {
	home := withFakeHome(t)
	p := filepath.Join(home, "x.bin")
	if err := os.WriteFile(p, []byte("hello"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	err := Verify(p, strings.Repeat("0", 64))
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Errorf("err = %v, quer ErrChecksumMismatch", err)
	}
}

func TestVerifyAbsentIsErrNotExist(t *testing.T) {
	withFakeHome(t)
	err := Verify(filepath.Join("nao", "existe.bin"), strings.Repeat("0", 64))
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("err = %v, quer satisfazer os.ErrNotExist", err)
	}
}

func TestEnsureCacheHitSkipsDownload(t *testing.T) {
	home := withFakeHome(t)
	payload := []byte("jar cacheado")
	sum := sha256Hex(payload)
	dest := filepath.Join(home, ".hubsaude", "simulador-0.1.7.jar")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatalf("setup mkdir: %v", err)
	}
	if err := os.WriteFile(dest, payload, 0o644); err != nil {
		t.Fatalf("setup write: %v", err)
	}

	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	art := Artifact{URL: srv.URL, Version: "0.1.7", Tag: "v0.1.7", SHA256: sum}
	if err := Ensure(context.Background(), nil, art, dest); err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if hit {
		t.Errorf("Ensure baixou apesar de cache valido")
	}
}

func TestEnsureRedownloadsOnMismatch(t *testing.T) {
	home := withFakeHome(t)
	novo := []byte("nova versao bytes")
	sum := sha256Hex(novo)
	dest := filepath.Join(home, ".hubsaude", "simulador-0.2.0.jar")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatalf("setup mkdir: %v", err)
	}
	if err := os.WriteFile(dest, []byte("conteudo velho"), 0o644); err != nil {
		t.Fatalf("setup write: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(novo)
	}))
	defer srv.Close()

	art := Artifact{URL: srv.URL, Version: "0.2.0", Tag: "v0.2.0", SHA256: sum}
	if err := Ensure(context.Background(), nil, art, dest); err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(novo) {
		t.Errorf("dest = %q, quer %q", got, novo)
	}
}

func TestEnsureDownloadsWhenAbsent(t *testing.T) {
	home := withFakeHome(t)
	payload := []byte("primeiro download")
	sum := sha256Hex(payload)
	dest := filepath.Join(home, ".hubsaude", "simulador-1.0.0.jar")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	art := Artifact{URL: srv.URL, Version: "1.0.0", Tag: "v1.0.0", SHA256: sum}
	if err := Ensure(context.Background(), nil, art, dest); err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("destino nao criado: %v", err)
	}
}

func TestEnsureSimuladorOrchestrates(t *testing.T) {
	home := withFakeHome(t)
	jarBytes := []byte("simulador jar simulado")
	jarSum := sha256Hex(jarBytes)

	mux := http.NewServeMux()
	mux.HandleFunc("/simulador.jar", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jarBytes)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	manifest := fmt.Sprintf(`{
        "validador": {"url": "%s/v.jar", "version": "0.1.8", "tag": "v0.1.8", "sha256": "%s"},
        "simulador": {"url": "%s/simulador.jar", "version": "0.3.0", "tag": "sim-v0.3.0", "sha256": "%s"},
        "jre": {"linux_x64": "https://example.com/jre"}
    }`, srv.URL, strings.Repeat("a", 64), srv.URL, jarSum)
	mux.HandleFunc("/release.json", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(manifest))
	})

	got, err := EnsureSimulador(context.Background(), nil, srv.URL+"/release.json")
	if err != nil {
		t.Fatalf("EnsureSimulador: %v", err)
	}
	want := filepath.Join(home, ".hubsaude", "simulador-0.3.0.jar")
	if got != want {
		t.Errorf("path = %q, quer %q", got, want)
	}
	if err := Verify(got, jarSum); err != nil {
		t.Errorf("Verify do jar baixado: %v", err)
	}
}

func TestEnsureSimuladorPropagatesManifestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "no", http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := EnsureSimulador(context.Background(), nil, srv.URL); err == nil {
		t.Errorf("queria erro de manifesto, got nil")
	}
}
