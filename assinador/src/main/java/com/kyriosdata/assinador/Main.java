package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;

import java.io.IOException;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;

public class Main {

    /** Porta padrão do modo servidor quando --port não é informado. */
    static final int DEFAULT_SERVER_PORT = 8080;

    static {
        // Garante UTF-8 nos streams independente do encoding do console
        // (critério D da especificacao@4d7d40f: "Encoding UTF-8 declarado").
        // Em Windows, o default de System.out é cp1252, o que corrompe
        // acentos no JSON de saída.
        System.setOut(new PrintStream(System.out, true, StandardCharsets.UTF_8));
        System.setErr(new PrintStream(System.err, true, StandardCharsets.UTF_8));
    }

    public static void main(String[] args) {
        if (args.length == 0) {
            printError("Uso: assinador.jar sign|validate --content <conteudo> [--signature <assinatura>] [--token <token>]"
                + " | server [--port <porta>]");
            System.exit(1);
        }

        String command = args[0];

        if ("server".equals(command)) {
            runServer(args);
            return;
        }

        String content = null;
        String signature = null;
        String token = null;

        for (int i = 1; i < args.length - 1; i++) {
            switch (args[i]) {
                case "--content":   content   = args[++i]; break;
                case "--signature": signature = args[++i]; break;
                case "--token":     token     = args[++i]; break;
            }
        }

        FakeSignatureService service = new FakeSignatureService();
        SignatureResponse response;

        switch (command) {
            case "sign": {
                SignRequest req = new SignRequest();
                req.setContent(content);
                req.setToken(token);
                response = service.sign(req);
                break;
            }
            case "validate": {
                ValidateRequest req = new ValidateRequest();
                req.setContent(content);
                req.setSignature(signature);
                response = service.validate(req);
                break;
            }
            default:
                printError("Comando desconhecido: " + command);
                System.exit(1);
                return;
        }

        System.out.println(Json.toJson(response));
        System.exit(response.isValid() ? 0 : 1);
    }

    /**
     * Inicia o modo servidor HTTP (US-02.4) e bloqueia até o encerramento
     * via POST /shutdown. Aceita {@code --port <porta>}; usa
     * {@link #DEFAULT_SERVER_PORT} quando omitido.
     */
    private static void runServer(String[] args) {
        int port = DEFAULT_SERVER_PORT;
        for (int i = 1; i < args.length; i++) {
            if ("--port".equals(args[i]) && i + 1 < args.length) {
                try {
                    port = Integer.parseInt(args[++i]);
                } catch (NumberFormatException e) {
                    printError("Porta invalida: " + args[i]);
                    System.exit(1);
                }
            }
        }

        try {
            HttpSignatureServer server = new HttpSignatureServer(port);
            server.start();
            System.out.println(Json.toJson(new SignatureResponse(
                null, true, "Servidor ouvindo na porta " + server.port())));
            server.awaitTermination();
        } catch (IOException e) {
            printError("Falha ao iniciar servidor: " + e.getMessage());
            System.exit(1);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    private static void printError(String message) {
        System.err.println(Json.toJson(new SignatureResponse(null, false, message)));
    }
}
