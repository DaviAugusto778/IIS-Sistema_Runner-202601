package jdk

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setGOARCH(t *testing.T, goarch string) {
	t.Helper()
	original := runtimeGOARCH
	runtimeGOARCH = goarch
	t.Cleanup(func() { runtimeGOARCH = original })
}

// adoptiumAssetsBaseForTest aponta a base da API para um servidor fake,
// restaurando o valor original ao fim do teste.
func adoptiumAssetsBaseForTest(t *testing.T, base string) string {
	t.Helper()
	original := adoptiumAssetsBase
	adoptiumAssetsBase = base
	t.Cleanup(func() { adoptiumAssetsBase = original })
	return original
}

func TestOSToken(t *testing.T) {
	cases := map[string]string{"windows": "windows", "linux": "linux", "darwin": "mac"}
	for goos, want := range cases {
		got, err := osToken(goos)
		if err != nil || got != want {
			t.Errorf("osToken(%q) = %q, %v; quer %q", goos, got, err, want)
		}
	}
	if _, err := osToken("plan9"); err == nil {
		t.Error("esperava erro para SO nao suportado")
	}
}

func TestArchToken(t *testing.T) {
	cases := map[string]string{"amd64": "x64", "arm64": "aarch64"}
	for goarch, want := range cases {
		got, err := archToken(goarch)
		if err != nil || got != want {
			t.Errorf("archToken(%q) = %q, %v; quer %q", goarch, got, err, want)
		}
	}
	if _, err := archToken("riscv64"); err == nil {
		t.Error("esperava erro para arquitetura nao suportada")
	}
}

func TestAssetsURL(t *testing.T) {
	setGOOS(t, "darwin")
	setGOARCH(t, "arm64")
	u, err := assetsURL()
	if err != nil {
		t.Fatalf("assetsURL: %v", err)
	}
	if !strings.Contains(u, "os=mac") || !strings.Contains(u, "architecture=aarch64") {
		t.Errorf("URL = %q, quer conter os=mac e architecture=aarch64", u)
	}
	if !strings.Contains(u, "image_type=jdk") {
		t.Errorf("URL = %q, quer conter image_type=jdk", u)
	}
}

func TestResolveAsset(t *testing.T) {
	setGOOS(t, "linux")
	setGOARCH(t, "amd64")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("os"); got != "linux" {
			t.Errorf("query os = %q, quer linux", got)
		}
		fmt.Fprint(w, `[{"binary":{"package":{"link":"https://x/jdk.tar.gz","checksum":"abc123","name":"jdk.tar.gz"}}}]`)
	}))
	defer srv.Close()

	adoptiumAssetsBaseForTest(t, srv.URL)

	a, err := resolveAsset(context.Background(), srv.Client())
	if err != nil {
		t.Fatalf("resolveAsset: %v", err)
	}
	if a.URL != "https://x/jdk.tar.gz" || a.SHA256 != "abc123" || a.Filename != "jdk.tar.gz" {
		t.Errorf("asset = %+v", a)
	}
}

func TestResolveAsset_Vazio(t *testing.T) {
	setGOOS(t, "linux")
	setGOARCH(t, "amd64")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[]`)
	}))
	defer srv.Close()
	adoptiumAssetsBaseForTest(t, srv.URL)

	if _, err := resolveAsset(context.Background(), srv.Client()); err == nil {
		t.Error("esperava erro quando Adoptium nao retorna pacote")
	}
}

func TestArchiveExt(t *testing.T) {
	if got := archiveExt(asset{Filename: "jdk.zip"}); got != ".zip" {
		t.Errorf("ext zip = %q", got)
	}
	if got := archiveExt(asset{Filename: "jdk.tar.gz"}); got != ".tar.gz" {
		t.Errorf("ext tar.gz = %q", got)
	}
	if got := archiveExt(asset{URL: "https://x/jdk.zip"}); got != ".zip" {
		t.Errorf("ext via URL = %q", got)
	}
}

func TestSafeJoin_RejeitaTraversal(t *testing.T) {
	dest := t.TempDir()
	if _, err := safeJoin(dest, "../../etc/passwd"); err == nil {
		t.Error("esperava rejeicao de caminho com traversal")
	}
	if _, err := safeJoin(dest, "bin/java"); err != nil {
		t.Errorf("caminho valido rejeitado: %v", err)
	}
}

// TestExtractTarGz_EFindJavaHome cobre extração e localização do home num
// layout típico (raiz versionada + bin/java).
func TestExtractTarGz_EFindJavaHome(t *testing.T) {
	setGOOS(t, "linux") // bin/java
	archive := buildTarGz(t, map[string]string{
		"jdk-21.0.5+11/bin/java":   "binario",
		"jdk-21.0.5+11/lib/rt.txt": "lib",
		"jdk-21.0.5+11/release":    "JAVA_VERSION=21",
	})
	dest := t.TempDir()
	if err := extractArchive(archive, dest); err != nil {
		t.Fatalf("extractArchive: %v", err)
	}

	extracted := filepath.Join(dest, "jdk-21.0.5+11", "bin", "java")
	if _, err := os.Stat(extracted); err != nil {
		t.Fatalf("bin/java nao extraido: %v", err)
	}

	home, err := findJavaHome(dest)
	if err != nil {
		t.Fatalf("findJavaHome: %v", err)
	}
	if home != filepath.Join(dest, "jdk-21.0.5+11") {
		t.Errorf("home = %q, quer raiz versionada", home)
	}
}

// TestExtractZip_EFindJavaHome cobre o caminho Windows (.zip + java.exe).
func TestExtractZip_EFindJavaHome(t *testing.T) {
	setGOOS(t, "windows") // java.exe
	archive := buildZip(t, map[string]string{
		"jdk-21.0.5+11/bin/java.exe": "binario",
		"jdk-21.0.5+11/release":      "JAVA_VERSION=21",
	})
	dest := t.TempDir()
	if err := extractArchive(archive, dest); err != nil {
		t.Fatalf("extractArchive: %v", err)
	}

	home, err := findJavaHome(dest)
	if err != nil {
		t.Fatalf("findJavaHome: %v", err)
	}
	if home != filepath.Join(dest, "jdk-21.0.5+11") {
		t.Errorf("home = %q, quer raiz versionada", home)
	}
}

func TestFindJavaHome_Ausente(t *testing.T) {
	setGOOS(t, "linux")
	dest := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dest, "vazio"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := findJavaHome(dest); err == nil {
		t.Error("esperava erro quando bin/java nao existe")
	}
}

// --- helpers de construção de arquivos ---

func buildTarGz(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "jdk.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func buildZip(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "jdk.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
