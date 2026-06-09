package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/**
 * Valida parâmetros de criação de assinatura (US-02.2).
 *
 * <p>Autoridade única de validação: o CLI não deve replicar estas regras
 * (criterio E3). Retorna lista de violações; vazia quando a requisição é
 * válida. As mensagens identificam o parâmetro, o motivo e — quando
 * aplicável — como corrigir (criterio "falhar bem").</p>
 */
public final class SignRequestValidator {

    public static final int CONTENT_MAX_LENGTH = 1_048_576;
    public static final int TOKEN_MAX_LENGTH = 256;

    private SignRequestValidator() {}

    public static List<String> validate(SignRequest request) {
        if (request == null) {
            return Collections.singletonList(
                "request: parâmetro obrigatório ausente (forneça SignRequest não nulo)");
        }

        List<String> errors = new ArrayList<>();
        validateContent(request.getContent(), errors);
        validateToken(request.getToken(), errors);
        return errors;
    }

    private static void validateContent(String content, List<String> errors) {
        if (content == null) {
            errors.add("content: parâmetro obrigatório ausente (use --content <conteudo>)");
            return;
        }
        if (content.isBlank()) {
            errors.add("content: parâmetro vazio (forneça conteúdo a assinar)");
            return;
        }
        if (content.length() > CONTENT_MAX_LENGTH) {
            errors.add("content: excede tamanho máximo de " + CONTENT_MAX_LENGTH + " caracteres");
        }
    }

    private static void validateToken(String token, List<String> errors) {
        if (token == null) {
            return;
        }
        if (token.isBlank()) {
            errors.add("token: não pode estar em branco (omita o parâmetro ou forneça um valor)");
            return;
        }
        if (token.length() > TOKEN_MAX_LENGTH) {
            errors.add("token: excede tamanho máximo de " + TOKEN_MAX_LENGTH + " caracteres");
        }
    }
}
