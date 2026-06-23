package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.ValidateRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;

import java.util.List;

public class FakeSignatureService implements SignatureService {

    private static final String FAKE_SIGNATURE = "MOCKED_SIGNATURE_BASE64_==";

    @Override
    public SignatureResponse sign(SignRequest request) {
        List<String> errors = SignRequestValidator.validate(request);
        if (!errors.isEmpty()) {
            return new SignatureResponse(null, false, String.join("; ", errors));
        }
        return new SignatureResponse(FAKE_SIGNATURE, true, "Assinatura criada com sucesso");
    }

    @Override
    public SignatureResponse validate(ValidateRequest request) {
        List<String> errors = ValidateRequestValidator.validate(request);
        if (!errors.isEmpty()) {
            return new SignatureResponse(null, false, String.join("; ", errors));
        }

        boolean isValid = FAKE_SIGNATURE.equals(request.getSignature());
        return new SignatureResponse(request.getSignature(), isValid, isValid ? "Assinatura é válida" : "Assinatura é inválida");
    }
}
