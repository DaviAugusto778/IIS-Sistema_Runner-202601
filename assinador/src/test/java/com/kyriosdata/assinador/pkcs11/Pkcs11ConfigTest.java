package com.kyriosdata.assinador.pkcs11;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

class Pkcs11ConfigTest {

    @Test
    void rejeitaBibliotecaNula() {
        assertThrows(IllegalArgumentException.class,
            () -> new Pkcs11Config("Runner", null, 0));
    }

    @Test
    void rejeitaBibliotecaEmBranco() {
        assertThrows(IllegalArgumentException.class,
            () -> new Pkcs11Config("Runner", "   ", 0));
    }

    @Test
    void usaNomePadraoQuandoOmitido() {
        Pkcs11Config config = new Pkcs11Config(null, "/lib/libsofthsm2.so", 0);
        assertEquals(Pkcs11Config.DEFAULT_NAME, config.name());
    }

    @Test
    void configComSlotExplicito() {
        Pkcs11Config config = new Pkcs11Config("Runner", "/lib/libsofthsm2.so", 3);
        String cfg = config.toProviderConfig();
        assertTrue(cfg.startsWith("--"), "config inline deve iniciar com --");
        assertTrue(cfg.contains("name = Runner"));
        assertTrue(cfg.contains("library = /lib/libsofthsm2.so"));
        assertTrue(cfg.contains("slot = 3"));
        assertTrue(cfg.indexOf("slotListIndex") < 0, "não deve usar slotListIndex quando slot é dado");
    }

    @Test
    void configSemSlotUsaPrimeiroSlot() {
        Pkcs11Config config = new Pkcs11Config("Runner", "/lib/libsofthsm2.so", null);
        assertNull(config.slot());
        String cfg = config.toProviderConfig();
        assertTrue(cfg.contains("slotListIndex = 0"));
        assertTrue(cfg.indexOf("slot = ") < 0, "sem slot explícito não deve emitir 'slot ='");
    }

    @Test
    void aparaEspacosDaBibliotecaEDoNome() {
        Pkcs11Config config = new Pkcs11Config("  Tok  ", "  /lib/x.so  ", null);
        assertEquals("Tok", config.name());
        assertEquals("/lib/x.so", config.library());
    }
}
