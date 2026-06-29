package com.kyriosdata.assinador.pkcs11;

/**
 * Sinaliza falha na interação com o dispositivo criptográfico via PKCS#11
 * (provider indisponível, biblioteca inexistente, slot inválido, PIN
 * incorreto, token ausente, etc.).
 *
 * <p>É uma exceção verificada para forçar o tratamento explícito do erro de
 * sistema na camada de serviço, que o converte em uma resposta de falha
 * legível — em vez de propagar uma exceção não tratada (critério E3:
 * distinguir erro do usuário de erro do sistema).</p>
 */
public class Pkcs11Exception extends Exception {

    private static final long serialVersionUID = 1L;

    public Pkcs11Exception(String message) {
        super(message);
    }

    public Pkcs11Exception(String message, Throwable cause) {
        super(message, cause);
    }
}
