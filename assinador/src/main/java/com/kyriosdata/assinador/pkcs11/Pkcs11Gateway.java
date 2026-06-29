package com.kyriosdata.assinador.pkcs11;

/**
 * Abstração da interação de baixo nível com o dispositivo PKCS#11.
 *
 * <p>Isola a chamada real ao provider {@code SunPKCS11}
 * ({@link SunPkcs11Gateway}) do serviço de assinatura, permitindo testar a
 * lógica de decisão ({@link Pkcs11SignatureService}) com um dublê — sem exigir
 * um token físico ou SoftHSM2 instalado. O teste de integração que comprova as
 * chamadas PKCS#11 reais (critério E5) exercita a implementação concreta e é
 * condicional à presença do dispositivo.</p>
 */
public interface Pkcs11Gateway {

    /**
     * Autentica no dispositivo com o PIN e confirma acesso ao seu conteúdo.
     *
     * @param config parâmetros do provider PKCS#11
     * @param pin    PIN/credencial do token; pode ser vazio quando o dispositivo
     *               não exige autenticação
     * @return informações do token efetivamente acessado
     * @throws Pkcs11Exception se a interação com o dispositivo falhar
     */
    TokenInfo authenticate(Pkcs11Config config, char[] pin) throws Pkcs11Exception;

    /**
     * Dados observados do dispositivo após uma interação bem-sucedida, usados
     * apenas para evidenciar (em mensagem ao usuário) que a comunicação
     * ocorreu de fato.
     *
     * @param provider nome do provider configurado (ex.: {@code SunPKCS11-Runner})
     * @param alias    primeiro alias do keystore do token, ou {@code null} se vazio
     * @param subject  subject do certificado X.509 associado, ou {@code null}
     */
    record TokenInfo(String provider, String alias, String subject) {
    }
}
