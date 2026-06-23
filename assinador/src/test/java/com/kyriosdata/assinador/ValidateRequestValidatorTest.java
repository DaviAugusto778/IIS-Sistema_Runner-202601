package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.ValidateRequest;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

class ValidateRequestValidatorTest {

    @Test
    void shouldReturnNoErrorsForValidRequest() {
        ValidateRequest req = new ValidateRequest();
        req.setContent("conteudo original");
        req.setSignature("MOCKED_SIGNATURE_BASE64_==");

        List<String> errors = ValidateRequestValidator.validate(req);

        assertTrue(errors.isEmpty(), "esperado sem erros, obtido: " + errors);
    }

    @Test
    void shouldReportErrorWhenRequestIsNull() {
        List<String> errors = ValidateRequestValidator.validate(null);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("request: "),
            "mensagem deve identificar parametro 'request': " + errors.get(0));
    }

    @Test
    void shouldReportErrorWhenContentIsNull() {
        ValidateRequest req = new ValidateRequest();
        req.setSignature("MOCKED_SIGNATURE_BASE64_==");

        List<String> errors = ValidateRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("content: "));
        assertTrue(errors.get(0).contains("ausente"));
        assertTrue(errors.get(0).contains("--content"), "mensagem deve orientar correcao");
    }

    @Test
    void shouldReportErrorWhenContentIsEmpty() {
        ValidateRequest req = new ValidateRequest();
        req.setContent("");
        req.setSignature("MOCKED_SIGNATURE_BASE64_==");

        List<String> errors = ValidateRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("content: "));
        assertTrue(errors.get(0).contains("vazio"));
    }

    @Test
    void shouldReportErrorWhenContentIsOnlyWhitespace() {
        ValidateRequest req = new ValidateRequest();
        req.setContent("   \t\n  ");
        req.setSignature("MOCKED_SIGNATURE_BASE64_==");

        List<String> errors = ValidateRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("content: "));
        assertTrue(errors.get(0).contains("vazio"));
    }

    @Test
    void shouldReportErrorWhenContentExceedsMaxLength() {
        ValidateRequest req = new ValidateRequest();
        req.setContent("a".repeat(ValidateRequestValidator.CONTENT_MAX_LENGTH + 1));
        req.setSignature("MOCKED_SIGNATURE_BASE64_==");

        List<String> errors = ValidateRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("content: "));
        assertTrue(errors.get(0).contains("excede tamanho máximo"));
        assertTrue(errors.get(0).contains(String.valueOf(ValidateRequestValidator.CONTENT_MAX_LENGTH)));
    }

    @Test
    void shouldAcceptContentExactlyAtMaxLength() {
        ValidateRequest req = new ValidateRequest();
        req.setContent("a".repeat(ValidateRequestValidator.CONTENT_MAX_LENGTH));
        req.setSignature("MOCKED_SIGNATURE_BASE64_==");

        List<String> errors = ValidateRequestValidator.validate(req);

        assertTrue(errors.isEmpty(), "tamanho exato no limite deve ser aceito: " + errors);
    }

    @Test
    void shouldReportErrorWhenSignatureIsNull() {
        ValidateRequest req = new ValidateRequest();
        req.setContent("conteudo original");

        List<String> errors = ValidateRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("signature: "));
        assertTrue(errors.get(0).contains("ausente"));
        assertTrue(errors.get(0).contains("--signature"), "mensagem deve orientar correcao");
    }

    @Test
    void shouldReportErrorWhenSignatureIsEmpty() {
        ValidateRequest req = new ValidateRequest();
        req.setContent("conteudo original");
        req.setSignature("");

        List<String> errors = ValidateRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("signature: "));
        assertTrue(errors.get(0).contains("vazio"));
    }

    @Test
    void shouldReportErrorWhenSignatureIsOnlyWhitespace() {
        ValidateRequest req = new ValidateRequest();
        req.setContent("conteudo original");
        req.setSignature("   ");

        List<String> errors = ValidateRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("signature: "));
        assertTrue(errors.get(0).contains("vazio"));
    }

    @Test
    void shouldReportErrorWhenSignatureExceedsMaxLength() {
        ValidateRequest req = new ValidateRequest();
        req.setContent("conteudo original");
        req.setSignature("a".repeat(ValidateRequestValidator.SIGNATURE_MAX_LENGTH + 1));

        List<String> errors = ValidateRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("signature: "));
        assertTrue(errors.get(0).contains("excede tamanho máximo"));
    }

    @Test
    void shouldAcceptSignatureExactlyAtMaxLength() {
        ValidateRequest req = new ValidateRequest();
        req.setContent("conteudo original");
        req.setSignature("a".repeat(ValidateRequestValidator.SIGNATURE_MAX_LENGTH));

        List<String> errors = ValidateRequestValidator.validate(req);

        assertTrue(errors.isEmpty(), "assinatura no limite deve ser aceita: " + errors);
    }

    @Test
    void shouldReportAllErrorsWhenMultipleParametersInvalid() {
        ValidateRequest req = new ValidateRequest();
        req.setContent(null);
        req.setSignature(null);

        List<String> errors = ValidateRequestValidator.validate(req);

        assertEquals(2, errors.size(),
            "deve relatar content e signature simultaneamente: " + errors);
        assertTrue(errors.stream().anyMatch(e -> e.startsWith("content: ")));
        assertTrue(errors.stream().anyMatch(e -> e.startsWith("signature: ")));
    }
}
