package invoker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fakeExecCommand devolve um exec.Cmd que reentra no próprio binário de
// teste invocando TestHelperProcess. O helper consulta as variáveis de
// ambiente HELPER_STDOUT, HELPER_STDERR e HELPER_EXIT para simular o
// comportamento do java -jar assinador.jar.
func fakeExecCommand(name string, args ...string) *exec.Cmd {
	cs := append([]string{"-test.run=TestHelperProcess", "--", name}, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if out := os.Getenv("HELPER_STDOUT"); out != "" {
		fmt.Fprint(os.Stdout, out)
	}
	if errOut := os.Getenv("HELPER_STDERR"); errOut != "" {
		fmt.Fprint(os.Stderr, errOut)
	}
	exitCode := 0
	if v := os.Getenv("HELPER_EXIT"); v != "" {
		fmt.Sscanf(v, "%d", &exitCode)
	}
	os.Exit(exitCode)
}

func withFakeExec(t *testing.T) {
	t.Helper()
	original := execCommand
	execCommand = fakeExecCommand
	t.Cleanup(func() { execCommand = original })
}

func withFakeJar(t *testing.T) string {
	t.Helper()
	jar := filepath.Join(t.TempDir(), "assinador.jar")
	if err := os.WriteFile(jar, []byte("fake"), 0o644); err != nil {
		t.Fatalf("setup fake jar: %v", err)
	}
	t.Setenv("ASSINADOR_JAR", jar)
	return jar
}

func TestInvoke_SuccessPropagatesStdout(t *testing.T) {
	withFakeExec(t)
	withFakeJar(t)
	t.Setenv("HELPER_STDOUT", `{"valid":true,"signature":"X","message":"ok"}`)
	t.Setenv("HELPER_EXIT", "0")

	result, err := Invoke("sign", "--content", "olá mundo")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got := string(result.Stdout); !strings.Contains(got, `"valid":true`) {
		t.Errorf("stdout = %q, quer conter \"valid\":true", got)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, quer 0", result.ExitCode)
	}
	if len(result.Stderr) != 0 {
		t.Errorf("Stderr = %q, quer vazio", result.Stderr)
	}
}

func TestInvoke_NonZeroExitPreservesPayload(t *testing.T) {
	withFakeExec(t)
	withFakeJar(t)
	t.Setenv("HELPER_STDOUT", `{"valid":false,"message":"conteúdo ausente"}`)
	t.Setenv("HELPER_STDERR", "")
	t.Setenv("HELPER_EXIT", "1")

	result, err := Invoke("sign")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, quer 1", result.ExitCode)
	}
	if !strings.Contains(string(result.Stdout), "conteúdo ausente") {
		t.Errorf("Stdout = %q, quer conter mensagem de erro", result.Stdout)
	}
}

func TestInvoke_SeparatesStdoutFromStderr(t *testing.T) {
	withFakeExec(t)
	withFakeJar(t)
	t.Setenv("HELPER_STDOUT", "RESULTADO")
	t.Setenv("HELPER_STDERR", "DIAGNOSTICO")
	t.Setenv("HELPER_EXIT", "0")

	result, err := Invoke("sign")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if string(result.Stdout) != "RESULTADO" {
		t.Errorf("Stdout = %q, quer %q", result.Stdout, "RESULTADO")
	}
	if string(result.Stderr) != "DIAGNOSTICO" {
		t.Errorf("Stderr = %q, quer %q", result.Stderr, "DIAGNOSTICO")
	}
}

func TestInvoke_MissingJarReturnsError(t *testing.T) {
	withFakeExec(t)
	t.Setenv("ASSINADOR_JAR", filepath.Join(t.TempDir(), "nao-existe.jar"))

	_, err := Invoke("sign")
	if err == nil {
		t.Fatal("esperava erro, recebeu nil")
	}
	if !strings.Contains(err.Error(), "nao encontrado") {
		t.Errorf("mensagem = %q, quer conter 'nao encontrado'", err.Error())
	}
}

func TestInvoke_PreservesArgumentsWithSpacesAndAccents(t *testing.T) {
	withFakeExec(t)
	jar := withFakeJar(t)
	t.Setenv("HELPER_EXIT", "0")

	// O helper apenas ecoa os args recebidos via HELPER_STDOUT. Como
	// não controlamos isso pelo helper diretamente, validamos
	// indiretamente: o invoker deve montar a linha de comando correta.
	// Verificamos via os.Args do subprocess através de um helper customizado.
	original := execCommand
	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedArgs = append([]string{name}, args...)
		return fakeExecCommand(name, args...)
	}
	t.Cleanup(func() { execCommand = original })

	_, err := Invoke("sign", "--content", "olá mundo com aspas \"")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	want := []string{"java", "-jar", jar, "sign", "--content", `olá mundo com aspas "`}
	if len(capturedArgs) != len(want) {
		t.Fatalf("args capturados = %v, quer %v", capturedArgs, want)
	}
	for i := range want {
		if capturedArgs[i] != want[i] {
			t.Errorf("args[%d] = %q, quer %q", i, capturedArgs[i], want[i])
		}
	}
}
