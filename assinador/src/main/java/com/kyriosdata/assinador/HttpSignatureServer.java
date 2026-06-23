package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

/**
 * Servidor HTTP do assinador (US-02.4). Expõe {@code POST /sign} e
 * {@code POST /validate}, reutilizando exatamente a mesma lógica de
 * validação e simulação do modo CLI ({@link SignatureService} +
 * {@link SignRequestValidator}/{@link ValidateRequestValidator}), e
 * {@code POST /shutdown} para encerramento gracioso (consumido pela
 * US-01.8 do CLI).
 *
 * <p>Usa {@code com.sun.net.httpserver.HttpServer}, embutido no JDK, sem
 * dependência externa (ADR-0004). As respostas seguem o contrato JSON de
 * uma linha {@code {"signature":...,"valid":...,"message":...}} do ADR-0003,
 * com Content-Type {@code application/json; charset=utf-8}.</p>
 *
 * <p>Códigos de status: {@code 200} quando a requisição é bem formada e
 * processada (mesmo que a assinatura seja inválida); {@code 400} para JSON
 * malformado ou parâmetros que falham na validação; {@code 405} para método
 * diferente de POST.</p>
 */
public final class HttpSignatureServer {

    private final HttpServer server;
    private final SignatureService service;
    private final CountDownLatch terminated = new CountDownLatch(1);
    private final AtomicBoolean stopped = new AtomicBoolean(false);

    // Watchdog de inatividade (US-01.9): quando idleMillis > 0, o servidor se
    // encerra após esse período sem requisições de /sign ou /validate. Probes
    // de /health não contam como atividade.
    private final long idleMillis;
    private final ScheduledExecutorService watchdog;
    private ScheduledFuture<?> pendingShutdown;

    public HttpSignatureServer(int port) throws IOException {
        this(port, new FakeSignatureService(), 0);
    }

    public HttpSignatureServer(int port, long idleMillis) throws IOException {
        this(port, new FakeSignatureService(), idleMillis);
    }

    public HttpSignatureServer(int port, SignatureService service) throws IOException {
        this(port, service, 0);
    }

    public HttpSignatureServer(int port, SignatureService service, long idleMillis) throws IOException {
        this.service = service;
        this.idleMillis = idleMillis;
        this.watchdog = idleMillis > 0
            ? Executors.newSingleThreadScheduledExecutor(r -> {
                Thread t = new Thread(r, "assinador-watchdog");
                t.setDaemon(true);
                return t;
            })
            : null;
        this.server = HttpServer.create(new InetSocketAddress(port), 0);
        this.server.createContext("/health", this::handleHealth);
        this.server.createContext("/sign", this::handleSign);
        this.server.createContext("/validate", this::handleValidate);
        this.server.createContext("/shutdown", this::handleShutdown);
        this.server.setExecutor(null); // executor padrão (síncrono)
    }

    /** Inicia o servidor em background e arma o watchdog de inatividade. */
    public void start() {
        server.start();
        rearmWatchdog();
    }

    /** (Re)agenda o auto-shutdown por inatividade. Chamado no start e a cada
     *  requisição de assinatura/validação. No-op quando o watchdog está
     *  desabilitado (idleMillis == 0). */
    private synchronized void rearmWatchdog() {
        if (watchdog == null) {
            return;
        }
        if (pendingShutdown != null) {
            pendingShutdown.cancel(false);
        }
        pendingShutdown = watchdog.schedule(this::stop, idleMillis, TimeUnit.MILLISECONDS);
    }

    /** Porta efetivamente vinculada (resolve porta efêmera quando 0). */
    public int port() {
        return server.getAddress().getPort();
    }

    /** Encerra o servidor (idempotente) e libera quem aguarda
     *  {@link #awaitTermination()}. */
    public void stop() {
        if (!stopped.compareAndSet(false, true)) {
            return;
        }
        if (watchdog != null) {
            watchdog.shutdownNow();
        }
        server.stop(0);
        terminated.countDown();
    }

    /** Bloqueia até o servidor ser encerrado via {@code /shutdown} ou {@link #stop()}. */
    public void awaitTermination() throws InterruptedException {
        terminated.await();
    }

    /** Liveness/readiness: responde 200 a qualquer método, indicando que o
     *  servidor já está aceitando requisições (alvo do readiness do CLI,
     *  US-01.5). */
    private void handleHealth(HttpExchange exchange) throws IOException {
        writeResponse(exchange, 200, new SignatureResponse(null, true, "ok"));
    }

    private void handleSign(HttpExchange exchange) throws IOException {
        if (!ensurePost(exchange)) {
            return;
        }
        rearmWatchdog();
        Map<String, String> json = parseBodyOrFail(exchange);
        if (json == null) {
            return;
        }
        SignRequest req = new SignRequest();
        req.setContent(json.get("content"));
        req.setToken(json.get("token"));

        List<String> errors = SignRequestValidator.validate(req);
        if (!errors.isEmpty()) {
            writeResponse(exchange, 400, new SignatureResponse(null, false, String.join("; ", errors)));
            return;
        }
        writeResponse(exchange, 200, service.sign(req));
    }

    private void handleValidate(HttpExchange exchange) throws IOException {
        if (!ensurePost(exchange)) {
            return;
        }
        rearmWatchdog();
        Map<String, String> json = parseBodyOrFail(exchange);
        if (json == null) {
            return;
        }
        ValidateRequest req = new ValidateRequest();
        req.setContent(json.get("content"));
        req.setSignature(json.get("signature"));

        List<String> errors = ValidateRequestValidator.validate(req);
        if (!errors.isEmpty()) {
            writeResponse(exchange, 400, new SignatureResponse(null, false, String.join("; ", errors)));
            return;
        }
        writeResponse(exchange, 200, service.validate(req));
    }

    private void handleShutdown(HttpExchange exchange) throws IOException {
        if (!ensurePost(exchange)) {
            return;
        }
        writeResponse(exchange, 200, new SignatureResponse(null, true, "Encerrando servidor"));
        // stop() fora da thread de dispatch: server.stop interrompe os
        // handlers ativos, e este é justamente um deles.
        new Thread(this::stop, "assinador-shutdown").start();
    }

    /** Garante POST; caso contrário responde 405 e devolve false. */
    private boolean ensurePost(HttpExchange exchange) throws IOException {
        if (!"POST".equalsIgnoreCase(exchange.getRequestMethod())) {
            writeResponse(exchange, 405, new SignatureResponse(null, false,
                "Metodo " + exchange.getRequestMethod() + " nao suportado: use POST"));
            return false;
        }
        return true;
    }

    /** Lê e parseia o corpo; em JSON malformado responde 400 e devolve null. */
    private Map<String, String> parseBodyOrFail(HttpExchange exchange) throws IOException {
        String body = readBody(exchange);
        try {
            return Json.parseObject(body);
        } catch (Json.JsonException e) {
            writeResponse(exchange, 400, new SignatureResponse(null, false, "JSON invalido: " + e.getMessage()));
            return null;
        }
    }

    private static String readBody(HttpExchange exchange) throws IOException {
        try (InputStream in = exchange.getRequestBody()) {
            return new String(in.readAllBytes(), StandardCharsets.UTF_8);
        }
    }

    private static void writeResponse(HttpExchange exchange, int status, SignatureResponse response) throws IOException {
        byte[] payload = Json.toJson(response).getBytes(StandardCharsets.UTF_8);
        exchange.getResponseHeaders().set("Content-Type", "application/json; charset=utf-8");
        exchange.sendResponseHeaders(status, payload.length);
        try (OutputStream out = exchange.getResponseBody()) {
            out.write(payload);
        }
    }
}
