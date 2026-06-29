package com.kyriosdata.assinador.pkcs11;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Testa o gateway PKCS#11 real ({@link SunPkcs11Gateway}) no caminho de erro,
 * sem exigir dispositivo: uma biblioteca inexistente deve virar
 * {@link Pkcs11Exception} legível, nunca uma exceção não tratada. Roda no CI
 * (depende apenas do provider SunPKCS11 do JDK). O caminho feliz com token real
 * está em {@code Pkcs11IntegrationTest} (condicional).
 */
class SunPkcs11GatewayTest {

    @Test
    void bibliotecaInexistenteFalhaComErroDeDominio() {
        SunPkcs11Gateway gateway = new SunPkcs11Gateway();
        Pkcs11Config config = new Pkcs11Config("Runner",
            "/caminho/que/nao/existe/libsofthsm2.so", 0);

        Pkcs11Exception ex = assertThrows(Pkcs11Exception.class,
            () -> gateway.authenticate(config, "1234".toCharArray()));

        assertNotNull(ex.getMessage());
        assertTrue(ex.getMessage().toLowerCase().contains("token"),
            "mensagem deve contextualizar a falha no token: " + ex.getMessage());
    }
}
