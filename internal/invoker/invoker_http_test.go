package invoker

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// httptestPort sobe um servidor de teste e devolve sua porta, religando o
// httpClient para apontar ao servidor (que escuta em 127.0.0.1).
func httptestPort(t *testing.T, srv *httptest.Server) int {
	t.Helper()
	_, portStr, err := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	p, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Atoi: %v", err)
	}
	return p
}

func TestInvokeHTTP_ValidRetornaExit0(t *testing.T) {
	var gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte(`{"signature":"SIG==","valid":true,"message":"ok"}`))
	}))
	defer srv.Close()

	res, err := InvokeHTTP(httptestPort(t, srv), "sign", map[string]string{"content": "ola"})
	if err != nil {
		t.Fatalf("InvokeHTTP: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("ExitCode = %d, quer 0", res.ExitCode)
	}
	if !strings.Contains(string(res.Stdout), "SIG==") {
		t.Errorf("Stdout = %q, quer conter corpo da resposta", res.Stdout)
	}
	if gotPath != "/sign" {
		t.Errorf("path = %q, quer /sign", gotPath)
	}
	// payload serializado corretamente
	var sent map[string]string
	if err := json.Unmarshal([]byte(gotBody), &sent); err != nil {
		t.Fatalf("corpo enviado nao e JSON: %q", gotBody)
	}
	if sent["content"] != "ola" {
		t.Errorf("content enviado = %q, quer 'ola'", sent["content"])
	}
}

func TestInvokeHTTP_ValidFalseRetornaExit1(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"signature":"X","valid":false,"message":"invalida"}`))
	}))
	defer srv.Close()

	res, err := InvokeHTTP(httptestPort(t, srv), "validate", map[string]string{"content": "x", "signature": "y"})
	if err != nil {
		t.Fatalf("InvokeHTTP: %v", err)
	}
	if res.ExitCode != 1 {
		t.Errorf("ExitCode = %d, quer 1 (valid:false)", res.ExitCode)
	}
}

func TestInvokeHTTP_400DeValidacaoRetornaExit1(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"signature":null,"valid":false,"message":"content: ausente"}`))
	}))
	defer srv.Close()

	res, err := InvokeHTTP(httptestPort(t, srv), "sign", map[string]string{})
	if err != nil {
		t.Fatalf("InvokeHTTP: %v", err)
	}
	if res.ExitCode != 1 {
		t.Errorf("ExitCode = %d, quer 1 (400 de validacao)", res.ExitCode)
	}
	if !strings.Contains(string(res.Stdout), "ausente") {
		t.Errorf("Stdout = %q, quer repassar a mensagem de erro", res.Stdout)
	}
}

func TestInvokeHTTP_ConexaoRecusadaRetornaErro(t *testing.T) {
	// porta sem servidor: httptest aberto e fechado libera a porta
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	port := httptestPort(t, srv)
	srv.Close()

	if _, err := InvokeHTTP(port, "sign", map[string]string{"content": "x"}); err == nil {
		t.Error("queria erro de conexao recusada, got nil")
	}
}
