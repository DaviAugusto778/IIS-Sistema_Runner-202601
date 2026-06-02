package release

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestDownload_HappyPath(t *testing.T) {
	body := []byte("conteudo do artefato")
	expected := sha256Hex(body)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "sub", "artefato.jar")
	if err := Download(context.Background(), srv.Client(), srv.URL, dest, expected); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ler dest: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("conteudo difere do enviado pelo servidor")
	}
	if _, err := os.Stat(dest + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temporario nao foi removido: %v", err)
	}
}

func TestDownload_HashMismatchRemovesPartialFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("conteudo errado"))
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "artefato.jar")
	wrongHash := strings.Repeat("0", 64)
	err := Download(context.Background(), srv.Client(), srv.URL, dest, wrongHash)
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("err = %v, quer ErrChecksumMismatch", err)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Errorf("dest existe apesar da falha: %v", statErr)
	}
	if _, statErr := os.Stat(dest + ".tmp"); !os.IsNotExist(statErr) {
		t.Errorf("temporario existe apesar da falha: %v", statErr)
	}
}

func TestDownload_EmptyHashSkipsValidation(t *testing.T) {
	body := []byte("qualquer coisa")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "jre.tar.gz")
	if err := Download(context.Background(), srv.Client(), srv.URL, dest, ""); err != nil {
		t.Fatalf("erro inesperado quando sha esta vazio: %v", err)
	}
}

func TestDownload_HashIsCaseInsensitive(t *testing.T) {
	body := []byte("teste case")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "a.jar")
	upper := strings.ToUpper(sha256Hex(body))
	if err := Download(context.Background(), srv.Client(), srv.URL, dest, upper); err != nil {
		t.Errorf("sha em maiusculas devia ser aceito: %v", err)
	}
}

func TestDownload_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nao encontrado", http.StatusNotFound)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "x.jar")
	err := Download(context.Background(), srv.Client(), srv.URL, dest, "")
	if err == nil {
		t.Fatal("esperava erro para 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("err = %v, quer mencionar 404", err)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Errorf("dest nao deveria existir para erro 404")
	}
}

func TestDownload_ContextCanceledStopsTransfer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		// Envia o primeiro byte e depois bloqueia até o cliente desistir.
		_, _ = w.Write([]byte("a"))
		if flusher != nil {
			flusher.Flush()
		}
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	dest := filepath.Join(t.TempDir(), "slow.jar")
	err := Download(ctx, srv.Client(), srv.URL, dest, "")
	if err == nil {
		t.Fatal("esperava erro por contexto cancelado")
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Errorf("dest nao deveria existir apos cancelamento")
	}
}

func TestDownloadArtifact_VerifiesArtifactHash(t *testing.T) {
	body := []byte("payload")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	art := Artifact{
		URL:     srv.URL,
		Version: "1.0.0",
		Tag:     "v1.0.0",
		SHA256:  sha256Hex(body),
	}
	dest := filepath.Join(t.TempDir(), "art.jar")
	if err := DownloadArtifact(context.Background(), srv.Client(), art, dest); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
}

func TestDownloadArtifact_DetectsTampering(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("payload adulterado"))
	}))
	defer srv.Close()

	art := Artifact{
		URL:    srv.URL,
		SHA256: sha256Hex([]byte("payload original")),
	}
	dest := filepath.Join(t.TempDir(), "art.jar")
	err := DownloadArtifact(context.Background(), nil, art, dest)
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Errorf("err = %v, quer ErrChecksumMismatch", err)
	}
}
