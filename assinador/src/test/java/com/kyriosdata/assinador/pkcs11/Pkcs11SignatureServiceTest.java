package com.kyriosdata.assinador.pkcs11;

import com.kyriosdata.assinador.FakeSignatureService;
import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Testa a lógica de decisão do {@link Pkcs11SignatureService} com um
 * {@link Pkcs11Gateway} dublê — sem token físico nem SoftHSM2. A interação
 * PKCS#11 real é coberta, condicionalmente, por {@code Pkcs11IntegrationTest}.
 */
class Pkcs11SignatureServiceTest {

    private static final String MOCK_SIGNATURE = "MOCKED_SIGNATURE_BASE64_==";
    private static final Pkcs11Config CONFIG = new Pkcs11Config("Runner", "/lib/fake.so", 0);

    /** Gateway que registra se foi chamado e qual PIN recebeu, sem tocar em hardware. */
    private static final class StubGateway implements Pkcs11Gateway {
        boolean called;
        String pinSeen;
        private final TokenInfo result;
        private final Pkcs11Exception failure;

        static StubGateway succeeding(TokenInfo info) {
            return new StubGateway(info, null);
        }

        static StubGateway failing(Pkcs11Exception e) {
            return new StubGateway(null, e);
        }

        private StubGateway(TokenInfo result, Pkcs11Exception failure) {
            this.result = result;
            this.failure = failure;
        }

        @Override
        public TokenInfo authenticate(Pkcs11Config config, char[] pin) throws Pkcs11Exception {
            this.called = true;
            this.pinSeen = new String(pin); // copia: o serviço zera o array depois
            if (failure != null) {
                throw failure;
            }
            return result;
        }
    }

    private static SignRequest signReq(String content, String token) {
        SignRequest req = new SignRequest();
        req.setContent(content);
        req.setToken(token);
        return req;
    }

    @Test
    void assinaturaValidaInteragecomDispositivoEMantemAssinaturaSimulada() {
        StubGateway gateway = StubGateway.succeeding(
            new Pkcs11Gateway.TokenInfo("SunPKCS11-Runner", "chave-1", "CN=Fulano, OU=SES, O=GO"));
        Pkcs11SignatureService service =
            new Pkcs11SignatureService(CONFIG, gateway, new FakeSignatureService());

        SignatureResponse resp = service.sign(signReq("documento", "1234"));

        assertTrue(resp.isValid());
        assertEquals(MOCK_SIGNATURE, resp.getSignature(), "saída permanece simulada");
        assertTrue(resp.getMessage().contains("PKCS#11"));
        assertTrue(resp.getMessage().contains("CN=Fulano"), "evidencia o certificado do token");
        assertTrue(gateway.called);
        assertEquals("1234", gateway.pinSeen, "PIN (token) é propagado ao dispositivo");
    }

    @Test
    void parametroInvalidoNaoAcionaDispositivo() {
        StubGateway gateway = StubGateway.succeeding(
            new Pkcs11Gateway.TokenInfo("SunPKCS11-Runner", null, null));
        Pkcs11SignatureService service =
            new Pkcs11SignatureService(CONFIG, gateway, new FakeSignatureService());

        SignatureResponse resp = service.sign(signReq(null, "1234"));

        assertFalse(resp.isValid());
        assertFalse(gateway.called, "erro de validação não deve tocar no token");
    }

    @Test
    void falhaDoDispositivoViraErroDeSistema() {
        StubGateway gateway = StubGateway.failing(new Pkcs11Exception("token ausente no slot 0"));
        Pkcs11SignatureService service =
            new Pkcs11SignatureService(CONFIG, gateway, new FakeSignatureService());

        SignatureResponse resp = service.sign(signReq("documento", "1234"));

        assertFalse(resp.isValid());
        assertNull(resp.getSignature());
        assertTrue(resp.getMessage().contains("PKCS#11"));
        assertTrue(resp.getMessage().contains("token ausente"));
        assertTrue(gateway.called);
    }

    @Test
    void semCertificadoUsaAliasNaEvidencia() {
        StubGateway gateway = StubGateway.succeeding(
            new Pkcs11Gateway.TokenInfo("SunPKCS11-Runner", "chave-sem-cert", null));
        Pkcs11SignatureService service =
            new Pkcs11SignatureService(CONFIG, gateway, new FakeSignatureService());

        SignatureResponse resp = service.sign(signReq("documento", null));

        assertTrue(resp.isValid());
        assertTrue(resp.getMessage().contains("alias: chave-sem-cert"));
        assertEquals("", gateway.pinSeen, "token nulo vira PIN vazio");
    }

    @Test
    void validacaoDelegaSemAcionarDispositivo() {
        StubGateway gateway = StubGateway.succeeding(
            new Pkcs11Gateway.TokenInfo("SunPKCS11-Runner", "x", null));
        Pkcs11SignatureService service =
            new Pkcs11SignatureService(CONFIG, gateway, new FakeSignatureService());

        ValidateRequest req = new ValidateRequest();
        req.setContent("documento");
        req.setSignature(MOCK_SIGNATURE);
        SignatureResponse resp = service.validate(req);

        assertTrue(resp.isValid());
        assertFalse(gateway.called, "validar assinatura não exige o dispositivo");
    }
}
