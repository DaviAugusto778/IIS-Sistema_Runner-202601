package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.ValidateRequest;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/**
 * Valida parâmetros de validação de assinatura (US-02.3).
 *
 * <p>Autoridade única de validação: o CLI não deve replicar estas regras
 * (criterio E3). Retorna lista de violações; vazia quando a requisição é
 * válida. As mensagens identificam o parâmetro, o motivo e — quando
 * aplicável — como corrigir (criterio "falhar bem").</p>
 *
 * <p>Espelha {@link SignRequestValidator}: ambos os fluxos compartilham o
 * mesmo estilo de mensagens e a mesma estratégia de relatar todas as
 * violações de uma vez.</p>
 */
public final class ValidateRequestValidator {

    public static final int CONTENT_MAX_LENGTH = 1_048_576;
    public static final int SIGNATURE_MAX_LENGTH = 8_192;

    private ValidateRequestValidator() {}

    public static List<String> validate(ValidateRequest request) {
        if (request == null) {
            return Collections.singletonList(
                "request: parâmetro obrigatório ausente (forneça ValidateRequest não nulo)");
        }

        List<String> errors = new ArrayList<>();
        validateContent(request.getContent(), errors);
        validateSignature(request.getSignature(), errors);
        return errors;
    }

    private static void validateContent(String content, List<String> errors) {
        if (content == null) {
            errors.add("content: parâmetro obrigatório ausente (use --content <conteudo>)");
            return;
        }
        if (content.isBlank()) {
            errors.add("content: parâmetro vazio (forneça o conteúdo original)");
            return;
        }
        if (content.length() > CONTENT_MAX_LENGTH) {
            errors.add("content: excede tamanho máximo de " + CONTENT_MAX_LENGTH + " caracteres");
        }
    }

    private static void validateSignature(String signature, List<String> errors) {
        if (signature == null) {
            errors.add("signature: parâmetro obrigatório ausente (use --signature <assinatura>)");
            return;
        }
        if (signature.isBlank()) {
            errors.add("signature: parâmetro vazio (forneça a assinatura a validar)");
            return;
        }
        if (signature.length() > SIGNATURE_MAX_LENGTH) {
            errors.add("signature: excede tamanho máximo de " + SIGNATURE_MAX_LENGTH + " caracteres");
        }
    }
}
