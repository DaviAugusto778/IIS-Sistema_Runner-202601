package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;

import java.io.PrintStream;
import java.nio.charset.StandardCharsets;

public class Main {

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
            printError("Uso: assinador.jar sign|validate --content <conteudo> [--signature <assinatura>] [--token <token>]");
            System.exit(1);
        }

        String command = args[0];
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

        System.out.println(toJson(response));
        System.exit(response.isValid() ? 0 : 1);
    }

    private static String toJson(SignatureResponse r) {
        String sig = r.getSignature() != null
            ? "\"" + r.getSignature() + "\""
            : "null";
        String msg = r.getMessage() != null
            ? r.getMessage().replace("\"", "\\\"")
            : "";
        return "{\"signature\":" + sig + ",\"valid\":" + r.isValid() + ",\"message\":\"" + msg + "\"}";
    }

    private static void printError(String message) {
        System.err.println("{\"signature\":null,\"valid\":false,\"message\":\"" + message + "\"}");
    }
}
