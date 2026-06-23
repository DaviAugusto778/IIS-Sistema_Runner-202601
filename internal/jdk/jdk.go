// Package jdk detecta e provisiona automaticamente um JDK 21 para que o
// usuário não precise instalar o Java manualmente (US-04.1).
//
// A resolução segue a ordem:
//  1. java disponível no PATH, desde que seja versão >= 21;
//  2. JDK gerenciado em ~/.hubsaude/jdk/ provisionado por execução anterior;
//  3. download automático da distribuição Eclipse Temurin adequada à
//     plataforma, extraído em ~/.hubsaude/jdk/ para reuso.
//
// O download nunca é repetido enquanto um JDK válido estiver disponível
// (PATH ou gerenciado), atendendo ao critério de cache da história.
package jdk

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// RequiredMajor é a versão major mínima do JDK exigida pelo assinador.jar
// (Java 21, conforme a stack do projeto).
const RequiredMajor = 21

// ErrNotFound indica que nenhum JDK 21 foi localizado no PATH nem no
// diretório gerenciado. É o gatilho para o provisionamento automático.
var ErrNotFound = errors.New("jdk: nenhum JDK 21 disponivel no PATH ou em ~/.hubsaude/jdk")

// Indireções para teste: permitem simular plataforma, busca no PATH e a
// consulta de versão sem depender do ambiente real.
var (
	runtimeGOOS   = runtime.GOOS
	runtimeGOARCH = runtime.GOARCH
	lookPath      = exec.LookPath

	// runVersion executa `<javaPath> -version` e devolve a saída combinada.
	// O launcher do Java escreve a versão em stderr, por isso usamos
	// CombinedOutput.
	runVersion = func(javaPath string) (string, error) {
		out, err := exec.Command(javaPath, "-version").CombinedOutput()
		return string(out), err
	}
)

// javaExeName devolve o nome do executável do launcher conforme a
// plataforma alvo ("java.exe" no Windows, "java" caso contrário).
func javaExeName() string {
	if runtimeGOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

// ManagedDir devolve ~/.hubsaude/jdk — a raiz do JDK gerenciado, onde o
// provisionamento extrai a distribuição para reuso.
func ManagedDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".hubsaude", "jdk"), nil
}

// ManagedJava devolve o caminho esperado do launcher dentro do JDK
// gerenciado (~/.hubsaude/jdk/bin/java[.exe]).
func ManagedJava() (string, error) {
	dir, err := ManagedDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "bin", javaExeName()), nil
}

// Detect localiza um launcher de JDK 21+ utilizável sem baixar nada.
//
// Tenta primeiro o PATH (reutiliza um Java já instalado pelo usuário) e,
// em seguida, o JDK gerenciado provisionado anteriormente. Devolve
// ErrNotFound quando nenhum candidato satisfaz a versão exigida.
func Detect() (string, error) {
	if p, err := lookPath("java"); err == nil && meetsRequirement(p) {
		return p, nil
	}
	if p, err := ManagedJava(); err == nil {
		if _, statErr := os.Stat(p); statErr == nil && meetsRequirement(p) {
			return p, nil
		}
	}
	return "", ErrNotFound
}

// meetsRequirement reporta se o launcher em javaPath é executável e tem
// versão major >= RequiredMajor.
func meetsRequirement(javaPath string) bool {
	out, err := runVersion(javaPath)
	if err != nil {
		return false
	}
	major, err := parseMajor(out)
	if err != nil {
		return false
	}
	return major >= RequiredMajor
}

// parseMajor extrai a versão major da saída de `java -version`.
//
// Suporta os formatos moderno (>= 9), ex.: `openjdk version "21.0.5"` →
// 21, e o legado (8), ex.: `java version "1.8.0_392"` → 8.
func parseMajor(versionOutput string) (int, error) {
	openQuote := strings.IndexByte(versionOutput, '"')
	if openQuote < 0 {
		return 0, errors.New("jdk: versao nao encontrada na saida de java -version")
	}
	rest := versionOutput[openQuote+1:]
	closeQuote := strings.IndexByte(rest, '"')
	if closeQuote < 0 {
		return 0, errors.New("jdk: versao malformada na saida de java -version")
	}
	version := rest[:closeQuote]

	parts := strings.Split(version, ".")
	first, err := leadingInt(parts[0])
	if err != nil {
		return 0, err
	}
	// Esquema legado "1.x": a major real é o segundo componente.
	if first == 1 && len(parts) > 1 {
		return leadingInt(parts[1])
	}
	return first, nil
}

// leadingInt converte o prefixo numérico de s em inteiro (ex.: "21+35" → 21).
func leadingInt(s string) (int, error) {
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, errors.New("jdk: componente de versao sem digitos: " + s)
	}
	return strconv.Atoi(s[:end])
}
