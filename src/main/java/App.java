package br.ufg.hubsaude;

public class App {
    public static void main(String[] args) {
        if (args.length == 0) {
            System.err.println("Erro: Nenhum comando fornecido. Use 'sign' ou 'validate'.");
            System.exit(1);
        }

        SignatureService service = new FakeSignatureService();
        String command = args[0];

        try {
            if ("sign".equalsIgnoreCase(command)) {
                // Para simulação, assumimos posições fixas nos args
                String message = args.length > 1 ? args[1] : null;
                String privKey = args.length > 2 ? args[2] : null;
                
                String result = service.sign(message, privKey);
                System.out.println("SUCESSO: " + result);

            } else if ("validate".equalsIgnoreCase(command)) {
                String message = args.length > 1 ? args[1] : null;
                String signature = args.length > 2 ? args[2] : null;
                String pubKey = args.length > 3 ? args[3] : null;

                boolean isValid = service.validate(message, signature, pubKey);
                System.out.println(isValid ? "VALIDO" : "INVALIDO");

            } else {
                System.err.println("Erro: Comando desconhecido '" + command + "'.");
                System.exit(1);
            }
        } catch (IllegalArgumentException e) {
            System.err.println(e.getMessage());
            System.exit(1);
        }
    }
}