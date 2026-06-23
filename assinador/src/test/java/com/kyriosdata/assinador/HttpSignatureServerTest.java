package com.kyriosdata.assinador;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;

import static org.junit.jupiter.api.Assertions.*;

class HttpSignatureServerTest {

    private static final String FAKE_SIGNATURE = "MOCKED_SIGNATURE_BASE64_==";

    private HttpSignatureServer server;
    private HttpClient client;
    private String base;

    @BeforeEach
    void setUp() throws Exception {
        server = new HttpSignatureServer(0); // porta efêmera
        server.start();
        base = "http://localhost:" + server.port();
        client = HttpClient.newHttpClient();
    }

    @AfterEach
    void tearDown() {
        server.stop();
    }

    private HttpResponse<String> post(String path, String body) throws Exception {
        HttpRequest req = HttpRequest.newBuilder(URI.create(base + path))
            .timeout(Duration.ofSeconds(5))
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build();
        return client.send(req, HttpResponse.BodyHandlers.ofString());
    }

    private HttpResponse<String> get(String path) throws Exception {
        HttpRequest req = HttpRequest.newBuilder(URI.create(base + path))
            .timeout(Duration.ofSeconds(5))
            .GET()
            .build();
        return client.send(req, HttpResponse.BodyHandlers.ofString());
    }

    @Test
    void signComParametrosValidosRetorna200EAssinatura() throws Exception {
        HttpResponse<String> resp = post("/sign", "{\"content\":\"ola mundo\"}");

        assertEquals(200, resp.statusCode());
        assertTrue(resp.body().contains("\"valid\":true"), resp.body());
        assertTrue(resp.body().contains(FAKE_SIGNATURE), resp.body());
    }

    @Test
    void signSemContentRetorna400() throws Exception {
        HttpResponse<String> resp = post("/sign", "{}");

        assertEquals(400, resp.statusCode());
        assertTrue(resp.body().contains("\"valid\":false"), resp.body());
        assertTrue(resp.body().contains("content:"), resp.body());
    }

    @Test
    void validateComAssinaturaCorretaRetorna200Valido() throws Exception {
        HttpResponse<String> resp = post("/validate",
            "{\"content\":\"ola\",\"signature\":\"" + FAKE_SIGNATURE + "\"}");

        assertEquals(200, resp.statusCode());
        assertTrue(resp.body().contains("\"valid\":true"), resp.body());
    }

    @Test
    void validateComAssinaturaErradaRetorna200Invalido() throws Exception {
        HttpResponse<String> resp = post("/validate",
            "{\"content\":\"ola\",\"signature\":\"ERRADA==\"}");

        // requisição bem formada e processada → 200, mesmo com valid:false
        assertEquals(200, resp.statusCode());
        assertTrue(resp.body().contains("\"valid\":false"), resp.body());
    }

    @Test
    void validateSemSignatureRetorna400() throws Exception {
        HttpResponse<String> resp = post("/validate", "{\"content\":\"ola\"}");

        assertEquals(400, resp.statusCode());
        assertTrue(resp.body().contains("signature:"), resp.body());
    }

    @Test
    void corpoJsonMalformadoRetorna400() throws Exception {
        HttpResponse<String> resp = post("/sign", "{nao eh json");

        assertEquals(400, resp.statusCode());
        assertTrue(resp.body().contains("JSON invalido"), resp.body());
    }

    @Test
    void metodoGetRetorna405() throws Exception {
        HttpResponse<String> resp = get("/sign");

        assertEquals(405, resp.statusCode());
        assertTrue(resp.body().contains("POST"), resp.body());
    }

    @Test
    void respostaTemContentTypeJson() throws Exception {
        HttpResponse<String> resp = post("/sign", "{\"content\":\"x\"}");

        String contentType = resp.headers().firstValue("Content-Type").orElse("");
        assertTrue(contentType.contains("application/json"), contentType);
        assertTrue(contentType.contains("utf-8"), contentType);
    }

    @Test
    void healthRetorna200() throws Exception {
        HttpResponse<String> resp = get("/health");

        assertEquals(200, resp.statusCode());
        assertTrue(resp.body().contains("\"valid\":true"), resp.body());
    }

    @Test
    void shutdownRetorna200EEncerraServidor() throws Exception {
        HttpResponse<String> resp = post("/shutdown", "");

        assertEquals(200, resp.statusCode());
        assertTrue(resp.body().contains("Encerrando"), resp.body());

        // após o shutdown o servidor deve parar de aceitar conexões
        boolean down = false;
        for (int i = 0; i < 20 && !down; i++) {
            try {
                post("/sign", "{\"content\":\"x\"}");
                Thread.sleep(100);
            } catch (Exception e) {
                down = true;
            }
        }
        assertTrue(down, "servidor deveria ter encerrado após /shutdown");
    }
}
