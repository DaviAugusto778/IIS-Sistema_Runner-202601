package com.kyriosdata.assinador.pkcs11;

import org.junit.jupiter.api.Test;

import java.nio.file.Files;
import java.nio.file.Path;

import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.junit.jupiter.api.Assumptions.assumeTrue;

/**
 * Teste de integração que comprova chamadas PKCS#11 <em>reais</em> ao dispositivo
 * (critério E5). É condicional ({@code assumeTrue}): roda apenas quando um
 * dispositivo/SoftHSM2 está provisionado e apontado por variáveis de ambiente,
 * mantendo o CI verde sem dependência externa.
 *
 * <p>Para executá-lo localmente com SoftHSM2:</p>
 * <pre>{@code
 * softhsm2-util --init-token --slot 0 --label runner --pin 1234 --so-pin 1234
 * export PKCS11_LIBRARY=/usr/lib/softhsm/libsofthsm2.so
 * export PKCS11_PIN=1234
 * # opcional: export PKCS11_SLOT=<id do slot inicializado>
 * mvn -Dtest=Pkcs11IntegrationTest test
 * }</pre>
 */
class Pkcs11IntegrationTest {

    @Test
    void autenticaContraDispositivoReal() throws Exception {
        String library = System.getenv("PKCS11_LIBRARY");
        String pin = System.getenv("PKCS11_PIN");
        String slotEnv = System.getenv("PKCS11_SLOT");

        assumeTrue(library != null && !library.isBlank(),
            "PKCS11_LIBRARY não definido; teste de integração PKCS#11 ignorado");
        assumeTrue(Files.exists(Path.of(library)),
            "biblioteca PKCS#11 inexistente: " + library);

        Integer slot = (slotEnv != null && !slotEnv.isBlank()) ? Integer.valueOf(slotEnv) : null;
        Pkcs11Config config = new Pkcs11Config("RunnerIT", library, slot);
        Pkcs11Gateway gateway = new SunPkcs11Gateway();

        Pkcs11Gateway.TokenInfo info = gateway.authenticate(
            config, pin != null ? pin.toCharArray() : new char[0]);

        assertNotNull(info, "interação bem-sucedida deve retornar dados do token");
        assertNotNull(info.provider());
        assertTrue(info.provider().startsWith("SunPKCS11"),
            "provider esperado SunPKCS11-*, obtido: " + info.provider());
    }
}
