package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

class SignRequestValidatorTest {

    @Test
    void shouldReturnNoErrorsForValidRequestWithoutToken() {
        SignRequest req = new SignRequest();
        req.setContent("conteudo a assinar");

        List<String> errors = SignRequestValidator.validate(req);

        assertTrue(errors.isEmpty(), "esperado sem erros, obtido: " + errors);
    }

    @Test
    void shouldReturnNoErrorsForValidRequestWithToken() {
        SignRequest req = new SignRequest();
        req.setContent("conteudo a assinar");
        req.setToken("PIN-1234");

        List<String> errors = SignRequestValidator.validate(req);

        assertTrue(errors.isEmpty(), "esperado sem erros, obtido: " + errors);
    }

    @Test
    void shouldReportErrorWhenRequestIsNull() {
        List<String> errors = SignRequestValidator.validate(null);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("request: "),
            "mensagem deve identificar parametro 'request': " + errors.get(0));
    }

    @Test
    void shouldReportErrorWhenContentIsNull() {
        SignRequest req = new SignRequest();

        List<String> errors = SignRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("content: "));
        assertTrue(errors.get(0).contains("ausente"));
        assertTrue(errors.get(0).contains("--content"), "mensagem deve orientar correcao");
    }

    @Test
    void shouldReportErrorWhenContentIsEmpty() {
        SignRequest req = new SignRequest();
        req.setContent("");

        List<String> errors = SignRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("content: "));
        assertTrue(errors.get(0).contains("vazio"));
    }

    @Test
    void shouldReportErrorWhenContentIsOnlyWhitespace() {
        SignRequest req = new SignRequest();
        req.setContent("   \t\n  ");

        List<String> errors = SignRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("content: "));
        assertTrue(errors.get(0).contains("vazio"));
    }

    @Test
    void shouldReportErrorWhenContentExceedsMaxLength() {
        SignRequest req = new SignRequest();
        req.setContent("a".repeat(SignRequestValidator.CONTENT_MAX_LENGTH + 1));

        List<String> errors = SignRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("content: "));
        assertTrue(errors.get(0).contains("excede tamanho máximo"));
        assertTrue(errors.get(0).contains(String.valueOf(SignRequestValidator.CONTENT_MAX_LENGTH)));
    }

    @Test
    void shouldAcceptContentExactlyAtMaxLength() {
        SignRequest req = new SignRequest();
        req.setContent("a".repeat(SignRequestValidator.CONTENT_MAX_LENGTH));

        List<String> errors = SignRequestValidator.validate(req);

        assertTrue(errors.isEmpty(), "tamanho exato no limite deve ser aceito: " + errors);
    }

    @Test
    void shouldReportErrorWhenTokenIsBlank() {
        SignRequest req = new SignRequest();
        req.setContent("ok");
        req.setToken("   ");

        List<String> errors = SignRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("token: "));
        assertTrue(errors.get(0).contains("branco"));
    }

    @Test
    void shouldReportErrorWhenTokenExceedsMaxLength() {
        SignRequest req = new SignRequest();
        req.setContent("ok");
        req.setToken("a".repeat(SignRequestValidator.TOKEN_MAX_LENGTH + 1));

        List<String> errors = SignRequestValidator.validate(req);

        assertEquals(1, errors.size());
        assertTrue(errors.get(0).startsWith("token: "));
        assertTrue(errors.get(0).contains("excede tamanho máximo"));
    }

    @Test
    void shouldAcceptTokenExactlyAtMaxLength() {
        SignRequest req = new SignRequest();
        req.setContent("ok");
        req.setToken("a".repeat(SignRequestValidator.TOKEN_MAX_LENGTH));

        List<String> errors = SignRequestValidator.validate(req);

        assertTrue(errors.isEmpty(), "token no limite deve ser aceito: " + errors);
    }

    @Test
    void shouldReportAllErrorsWhenMultipleParametersInvalid() {
        SignRequest req = new SignRequest();
        req.setContent(null);
        req.setToken("   ");

        List<String> errors = SignRequestValidator.validate(req);

        assertEquals(2, errors.size(),
            "deve relatar content e token simultaneamente: " + errors);
        assertTrue(errors.stream().anyMatch(e -> e.startsWith("content: ")));
        assertTrue(errors.stream().anyMatch(e -> e.startsWith("token: ")));
    }
}
