package com.kyriosdata.assinador.pkcs11;

import java.security.KeyStore;
import java.security.Provider;
import java.security.Security;
import java.security.cert.Certificate;
import java.security.cert.X509Certificate;
import java.util.Enumeration;

/**
 * Implementação real da {@link Pkcs11Gateway} sobre o provider {@code SunPKCS11}
 * do JDK (módulo {@code jdk.crypto.cryptoki}) — sem dependências externas.
 *
 * <p>Configura o provider dinamicamente a partir de {@link Pkcs11Config}, abre o
 * {@code KeyStore} do tipo {@code PKCS11} e realiza o login no token com o PIN.
 * São chamadas PKCS#11 reais ao dispositivo (ou ao SoftHSM2 nos testes de
 * integração). Após o login, lê o primeiro alias e o certificado associado
 * apenas para evidenciar que a comunicação ocorreu.</p>
 *
 * <p>O resultado criptográfico em si permanece simulado (escopo do projeto):
 * quem decide o conteúdo da assinatura é {@link Pkcs11SignatureService}; esta
 * classe responde só pela <em>interação</em> com o hardware.</p>
 */
public final class SunPkcs11Gateway implements Pkcs11Gateway {

    @Override
    public TokenInfo authenticate(Pkcs11Config config, char[] pin) throws Pkcs11Exception {
        Provider base = Security.getProvider("SunPKCS11");
        if (base == null) {
            throw new Pkcs11Exception(
                "provider SunPKCS11 indisponível nesta JVM (módulo jdk.crypto.cryptoki ausente)");
        }
        try {
            Provider provider = base.configure(config.toProviderConfig());
            KeyStore keystore = KeyStore.getInstance("PKCS11", provider);
            // load(null, pin) dispara o C_Login real no dispositivo usando o PIN.
            keystore.load(null, pin);
            String alias = firstAlias(keystore);
            return new TokenInfo(provider.getName(), alias, subjectOf(keystore, alias));
        } catch (Pkcs11Exception e) {
            throw e;
        } catch (Exception e) {
            // Provider.configure / KeyStore.load encapsulam o motivo real
            // (lib inexistente, slot inválido, PIN incorreto) em ProviderException
            // ou IOException — convertemos em erro de domínio legível.
            throw new Pkcs11Exception(
                "falha ao interagir com o token: " + rootMessage(e), e);
        }
    }

    private static String firstAlias(KeyStore keystore) throws Exception {
        Enumeration<String> aliases = keystore.aliases();
        return aliases.hasMoreElements() ? aliases.nextElement() : null;
    }

    private static String subjectOf(KeyStore keystore, String alias) throws Exception {
        if (alias == null) {
            return null;
        }
        Certificate cert = keystore.getCertificate(alias);
        return (cert instanceof X509Certificate x509)
            ? x509.getSubjectX500Principal().getName()
            : null;
    }

    /** Mensagem da causa raiz, mais informativa que a do wrapper. */
    private static String rootMessage(Throwable t) {
        Throwable cause = t;
        while (cause.getCause() != null && cause.getCause() != cause) {
            cause = cause.getCause();
        }
        String msg = cause.getMessage();
        return (msg != null && !msg.isBlank()) ? msg : cause.getClass().getSimpleName();
    }
}
