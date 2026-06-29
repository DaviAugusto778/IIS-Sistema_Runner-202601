package com.kyriosdata.assinador.pkcs11;

import com.kyriosdata.assinador.SignatureService;
import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;

import java.util.Arrays;

/**
 * {@link SignatureService} que assina <em>interagindo com o dispositivo
 * criptográfico</em> via PKCS#11 (US-02.5), mantendo a saída simulada.
 *
 * <p>Decora um serviço base (tipicamente o {@code FakeSignatureService}), que
 * responde pela validação de parâmetros e pelo valor simulado da assinatura.
 * Acima disso, esta classe realiza a interação real com o token:</p>
 *
 * <ol>
 *   <li>delega validação + simulação ao serviço base;</li>
 *   <li>se os parâmetros forem válidos, autentica no dispositivo com o PIN
 *       (campo {@code token} da requisição) via {@link Pkcs11Gateway};</li>
 *   <li>devolve a assinatura simulada, anotando na mensagem a evidência da
 *       interação com o dispositivo.</li>
 * </ol>
 *
 * <p>Distinção de erros (critério E3): erro de validação (do usuário) é
 * devolvido tal como veio do serviço base, sem tocar no dispositivo; falha na
 * interação com o token (erro de sistema) vira uma resposta de falha com
 * mensagem própria. Esta classe só é instanciada quando o usuário opta
 * explicitamente pelo modo PKCS#11 — no fluxo padrão o dispositivo nunca é
 * acionado.</p>
 */
public final class Pkcs11SignatureService implements SignatureService {

    private final Pkcs11Config config;
    private final Pkcs11Gateway gateway;
    private final SignatureService delegate;

    public Pkcs11SignatureService(Pkcs11Config config, Pkcs11Gateway gateway, SignatureService delegate) {
        this.config = config;
        this.gateway = gateway;
        this.delegate = delegate;
    }

    @Override
    public SignatureResponse sign(SignRequest request) {
        SignatureResponse simulated = delegate.sign(request);
        if (!simulated.isValid()) {
            // erro de validação (usuário): não aciona o dispositivo
            return simulated;
        }

        String pin = (request != null) ? request.getToken() : null;
        char[] secret = (pin != null) ? pin.toCharArray() : new char[0];
        try {
            Pkcs11Gateway.TokenInfo info = gateway.authenticate(config, secret);
            return new SignatureResponse(simulated.getSignature(), true, describe(info));
        } catch (Pkcs11Exception e) {
            // erro de sistema: dispositivo configurado mas indisponível/recusado
            return new SignatureResponse(null, false, "dispositivo PKCS#11: " + e.getMessage());
        } finally {
            Arrays.fill(secret, '\0');
        }
    }

    @Override
    public SignatureResponse validate(ValidateRequest request) {
        // Validar uma assinatura não exige o dispositivo: mantém a lógica simulada.
        return delegate.validate(request);
    }

    private static String describe(Pkcs11Gateway.TokenInfo info) {
        StringBuilder sb = new StringBuilder(
            "assinatura simulada após interação com dispositivo PKCS#11");
        if (info.subject() != null) {
            sb.append(" (certificado: ").append(info.subject()).append(')');
        } else if (info.alias() != null) {
            sb.append(" (alias: ").append(info.alias()).append(')');
        }
        sb.append(" [provider ").append(info.provider()).append(']');
        return sb.toString();
    }
}
