public class FakeSignatureService implements SignatureService {

    @Override
    public String sign(String message, String privateKey) {
        // Validação estrita simulada (US-02.2)
        if (message == null || message.trim().isEmpty()) {
            throw new IllegalArgumentException("Erro: O parâmetro 'message' é obrigatório.");
        }
        if (privateKey == null || privateKey.trim().isEmpty()) {
            throw new IllegalArgumentException("Erro: O parâmetro 'privateKey' é obrigatório.");
        }

        // Retorna uma assinatura pré-construída para simular sucesso
        return "ass_simulada_7f8b9a0c1d2e3f4g5h6i7j8k9l0m";
    }

    @Override
    public boolean validate(String message, String signature, String publicKey) {
        // Validação estrita simulada (US-02.3)
        if (message == null || signature == null || publicKey == null) {
            throw new IllegalArgumentException("Erro: message, signature e publicKey são obrigatórios para validação.");
        }

        // Resultado pré-determinado baseado em um critério simples para simulação
        // Exemplo: se a assinatura tiver um tamanho esperado, consideramos válida.
        return signature.startsWith("ass_simulada_");
    }
}