public interface SignatureService {
    /**
     * Simula a criação de uma assinatura digital.
     * * @param message A mensagem ou hash a ser assinado.
     * @param privateKey Identificador da chave privada (simulado).
     * @return A assinatura digital gerada (simulada).
     */
    String sign(String message, String privateKey);

    /**
     * Simula a validação de uma assinatura digital.
     * * @param message A mensagem original.
     * @param signature A assinatura a ser validada.
     * @param publicKey A chave pública (simulada).
     * @return true se a assinatura for válida, false caso contrário.
     */
    boolean validate(String message, String signature, String publicKey);
}