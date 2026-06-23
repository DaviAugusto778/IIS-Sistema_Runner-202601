package jdk

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// setGOOS fixa a plataforma simulada durante o teste.
func setGOOS(t *testing.T, goos string) {
	t.Helper()
	original := runtimeGOOS
	runtimeGOOS = goos
	t.Cleanup(func() { runtimeGOOS = original })
}

// stubLookPath substitui a busca de "java" no PATH.
func stubLookPath(t *testing.T, fn func(string) (string, error)) {
	t.Helper()
	original := lookPath
	lookPath = fn
	t.Cleanup(func() { lookPath = original })
}

// stubRunVersion substitui a consulta de versão do launcher.
func stubRunVersion(t *testing.T, fn func(string) (string, error)) {
	t.Helper()
	original := runVersion
	runVersion = fn
	t.Cleanup(func() { runVersion = original })
}

// isolateHome aponta o home do usuário para um diretório temporário vazio,
// garantindo que nenhum JDK gerenciado real interfira no teste. Devolve o
// caminho do home isolado.
func isolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)        // Unix
	t.Setenv("USERPROFILE", home) // Windows
	return home
}

// createManagedJava cria um launcher gerenciado fictício em
// ~/.hubsaude/jdk/bin/java[.exe] e devolve seu caminho.
func createManagedJava(t *testing.T) string {
	t.Helper()
	java, err := ManagedJava()
	if err != nil {
		t.Fatalf("ManagedJava: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(java), 0o755); err != nil {
		t.Fatalf("criar bin gerenciado: %v", err)
	}
	if err := os.WriteFile(java, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("criar java gerenciado: %v", err)
	}
	return java
}

const java21Output = `openjdk version "21.0.5" 2024-10-15 LTS
OpenJDK Runtime Environment Temurin-21.0.5+11`

func TestParseMajor(t *testing.T) {
	cases := []struct {
		name string
		out  string
		want int
	}{
		{"temurin 21", `openjdk version "21.0.5" 2024-10-15`, 21},
		{"oracle 21 lts", `java version "21.0.11" 2026-04-21 LTS`, 21},
		{"sem patch", `openjdk version "21" 2023-09-19`, 21},
		{"build com mais", `openjdk version "17.0.9+11"`, 17},
		{"legado 8", `java version "1.8.0_392"`, 8},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseMajor(c.out)
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if got != c.want {
				t.Errorf("parseMajor = %d, quer %d", got, c.want)
			}
		})
	}
}

func TestParseMajor_Invalido(t *testing.T) {
	if _, err := parseMajor("sem aspas aqui"); err == nil {
		t.Error("esperava erro para saida sem versao")
	}
}

func TestJavaExeName(t *testing.T) {
	setGOOS(t, "windows")
	if got := javaExeName(); got != "java.exe" {
		t.Errorf("windows: javaExeName = %q, quer java.exe", got)
	}
	setGOOS(t, "linux")
	if got := javaExeName(); got != "java" {
		t.Errorf("linux: javaExeName = %q, quer java", got)
	}
}

// TestDetect_PATHPresente: java 21 no PATH é detectado nas três plataformas.
func TestDetect_PATHPresente(t *testing.T) {
	for _, goos := range []string{"windows", "linux", "darwin"} {
		t.Run(goos, func(t *testing.T) {
			setGOOS(t, goos)
			isolateHome(t)
			stubLookPath(t, func(string) (string, error) { return "/sys/bin/java", nil })
			stubRunVersion(t, func(string) (string, error) { return java21Output, nil })

			got, err := Detect()
			if err != nil {
				t.Fatalf("Detect erro: %v", err)
			}
			if got != "/sys/bin/java" {
				t.Errorf("Detect = %q, quer launcher do PATH", got)
			}
		})
	}
}

// TestDetect_Ausente: sem java no PATH e sem JDK gerenciado → ErrNotFound,
// nas três plataformas.
func TestDetect_Ausente(t *testing.T) {
	for _, goos := range []string{"windows", "linux", "darwin"} {
		t.Run(goos, func(t *testing.T) {
			setGOOS(t, goos)
			isolateHome(t)
			stubLookPath(t, func(string) (string, error) { return "", errors.New("not found") })
			stubRunVersion(t, func(string) (string, error) { return "", errors.New("nao deveria ser chamado") })

			_, err := Detect()
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("Detect erro = %v, quer ErrNotFound", err)
			}
		})
	}
}

// TestDetect_GerenciadoPresente: sem java no PATH, mas com JDK gerenciado
// provisionado, ele é reutilizado (sem download) nas três plataformas.
func TestDetect_GerenciadoPresente(t *testing.T) {
	for _, goos := range []string{"windows", "linux", "darwin"} {
		t.Run(goos, func(t *testing.T) {
			setGOOS(t, goos)
			isolateHome(t)
			stubLookPath(t, func(string) (string, error) { return "", errors.New("not found") })
			stubRunVersion(t, func(string) (string, error) { return java21Output, nil })
			managed := createManagedJava(t)

			got, err := Detect()
			if err != nil {
				t.Fatalf("Detect erro: %v", err)
			}
			if got != managed {
				t.Errorf("Detect = %q, quer JDK gerenciado %q", got, managed)
			}
		})
	}
}

// TestDetect_VersaoInsuficiente: java 17 no PATH e sem gerenciado → ErrNotFound.
func TestDetect_VersaoInsuficiente(t *testing.T) {
	setGOOS(t, "linux")
	isolateHome(t)
	stubLookPath(t, func(string) (string, error) { return "/sys/bin/java", nil })
	stubRunVersion(t, func(string) (string, error) { return `openjdk version "17.0.9"`, nil })

	_, err := Detect()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Detect erro = %v, quer ErrNotFound (versao < 21)", err)
	}
}

// TestDetect_PreferePATHaoGerenciado: havendo java 21 no PATH, ele é usado
// mesmo com gerenciado presente (evita download e prioriza instalação do
// usuário).
func TestDetect_PreferePATHaoGerenciado(t *testing.T) {
	setGOOS(t, "linux")
	isolateHome(t)
	createManagedJava(t)
	stubLookPath(t, func(string) (string, error) { return "/sys/bin/java", nil })
	stubRunVersion(t, func(string) (string, error) { return java21Output, nil })

	got, err := Detect()
	if err != nil {
		t.Fatalf("Detect erro: %v", err)
	}
	if got != "/sys/bin/java" {
		t.Errorf("Detect = %q, quer launcher do PATH", got)
	}
}
