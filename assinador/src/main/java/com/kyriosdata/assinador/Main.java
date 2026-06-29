package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;
import com.kyriosdata.assinador.pkcs11.Pkcs11Config;
import com.kyriosdata.assinador.pkcs11.Pkcs11SignatureService;
import com.kyriosdata.assinador.pkcs11.SunPkcs11Gateway;

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
                + " [--pkcs11-lib <caminho> [--pkcs11-slot <n>]]"
                + " | server [--port <porta>] [--timeout <minutos>] [--pkcs11-lib <caminho> [--pkcs11-slot <n>]]");
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
        String pkcs11Lib = null;
        Integer pkcs11Slot = null;

        for (int i = 1; i < args.length - 1; i++) {
            switch (args[i]) {
                case "--content":     content    = args[++i]; break;
                case "--signature":   signature  = args[++i]; break;
                case "--token":       token      = args[++i]; break;
                case "--pkcs11-lib":  pkcs11Lib  = args[++i]; break;
                case "--pkcs11-slot": pkcs11Slot = parseSlot(args[++i]); break;
            }
        }

        SignatureService service = buildService(pkcs11Lib, pkcs11Slot);
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
        int timeoutMin = 0;
        String pkcs11Lib = null;
        Integer pkcs11Slot = null;
        for (int i = 1; i < args.length; i++) {
            if ("--port".equals(args[i]) && i + 1 < args.length) {
                try {
                    port = Integer.parseInt(args[++i]);
                } catch (NumberFormatException e) {
                    printError("Porta invalida: " + args[i]);
                    System.exit(1);
                }
            } else if ("--timeout".equals(args[i]) && i + 1 < args.length) {
                try {
                    timeoutMin = Integer.parseInt(args[++i]);
                } catch (NumberFormatException e) {
                    printError("Timeout invalido: " + args[i]);
                    System.exit(1);
                }
            } else if ("--pkcs11-lib".equals(args[i]) && i + 1 < args.length) {
                pkcs11Lib = args[++i];
            } else if ("--pkcs11-slot".equals(args[i]) && i + 1 < args.length) {
                pkcs11Slot = parseSlot(args[++i]);
            }
        }

        long idleMillis = timeoutMin > 0 ? (long) timeoutMin * 60_000L : 0L;
        try {
            HttpSignatureServer server = new HttpSignatureServer(port, buildService(pkcs11Lib, pkcs11Slot), idleMillis);
            server.start();
            String msg = idleMillis > 0
                ? "Servidor ouvindo na porta " + server.port() + " (timeout de inatividade: " + timeoutMin + " min)"
                : "Servidor ouvindo na porta " + server.port();
            System.out.println(Json.toJson(new SignatureResponse(null, true, msg)));
            server.awaitTermination();
        } catch (IOException e) {
            printError("Falha ao iniciar servidor: " + e.getMessage());
            System.exit(1);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    /**
     * Seleciona o serviço de assinatura. Sem {@code --pkcs11-lib}, usa o
     * {@link FakeSignatureService} (fluxo padrão, sem dispositivo). Com a
     * biblioteca informada, embrulha o serviço simulado em
     * {@link Pkcs11SignatureService}, que interage com o token via PKCS#11
     * (US-02.5). O PKCS#11 é, portanto, estritamente opcional e nunca bloqueia
     * o fluxo padrão.
     */
    private static SignatureService buildService(String pkcs11Lib, Integer pkcs11Slot) {
        FakeSignatureService fake = new FakeSignatureService();
        if (pkcs11Lib == null) {
            return fake;
        }
        Pkcs11Config config = new Pkcs11Config(Pkcs11Config.DEFAULT_NAME, pkcs11Lib, pkcs11Slot);
        return new Pkcs11SignatureService(config, new SunPkcs11Gateway(), fake);
    }

    /** Converte o argumento de {@code --pkcs11-slot} em inteiro; encerra com erro claro se inválido. */
    private static Integer parseSlot(String raw) {
        try {
            return Integer.valueOf(raw);
        } catch (NumberFormatException e) {
            printError("Slot PKCS#11 invalido: " + raw + " (informe um inteiro)");
            System.exit(1);
            return null; // inalcançável
        }
    }

    private static void printError(String message) {
        System.err.println(Json.toJson(new SignatureResponse(null, false, message)));
    }
}
