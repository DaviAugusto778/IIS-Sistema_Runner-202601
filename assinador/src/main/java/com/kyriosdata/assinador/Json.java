package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignatureResponse;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Utilitário mínimo de JSON para os payloads do assinador — objetos planos
 * cujos valores são strings ou null (ex.: {@code {"content":"...","token":null}}).
 *
 * <p>Escrito à mão deliberadamente: os payloads são triviais e a política do
 * projeto é preferir a std lib a introduzir dependência (Jackson/Gson) só
 * para isso (ver CLAUDE.md e ADR-0004).</p>
 */
final class Json {

    private Json() {}

    /** Sinaliza JSON malformado durante o parse. */
    static final class JsonException extends RuntimeException {
        JsonException(String message) {
            super(message);
        }
    }

    /**
     * Serializa uma {@link SignatureResponse} no contrato único de uma linha
     * {@code {"signature":...,"valid":...,"message":...}} (ADR-0003).
     * Campos string nulos viram {@code null} literal.
     */
    static String toJson(SignatureResponse r) {
        return "{\"signature\":" + quoteOrNull(r.getSignature())
            + ",\"valid\":" + r.isValid()
            + ",\"message\":" + quoteOrNull(r.getMessage()) + "}";
    }

    private static String quoteOrNull(String s) {
        return s == null ? "null" : "\"" + escape(s) + "\"";
    }

    /** Escapa uma string para inclusão literal em JSON. */
    static String escape(String s) {
        StringBuilder sb = new StringBuilder(s.length() + 8);
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            switch (c) {
                case '"':  sb.append("\\\""); break;
                case '\\': sb.append("\\\\"); break;
                case '\n': sb.append("\\n"); break;
                case '\r': sb.append("\\r"); break;
                case '\t': sb.append("\\t"); break;
                case '\b': sb.append("\\b"); break;
                case '\f': sb.append("\\f"); break;
                default:
                    if (c < 0x20) {
                        sb.append(String.format("\\u%04x", (int) c));
                    } else {
                        sb.append(c);
                    }
            }
        }
        return sb.toString();
    }

    /**
     * Parseia um objeto JSON plano cujos valores são string ou null.
     * Preserva a ordem de inserção. Lança {@link JsonException} se a entrada
     * não for um objeto bem formado nesse subconjunto.
     */
    static Map<String, String> parseObject(String text) {
        Map<String, String> out = new LinkedHashMap<>();
        int i = skipWs(text, 0);
        if (i >= text.length() || text.charAt(i) != '{') {
            throw new JsonException("esperado '{' no inicio do objeto");
        }
        i = skipWs(text, i + 1);
        if (i < text.length() && text.charAt(i) == '}') {
            return out; // objeto vazio
        }

        int[] end = new int[1];
        while (true) {
            i = skipWs(text, i);
            if (i >= text.length() || text.charAt(i) != '"') {
                throw new JsonException("esperada chave entre aspas");
            }
            String key = parseString(text, i, end);
            i = skipWs(text, end[0]);
            if (i >= text.length() || text.charAt(i) != ':') {
                throw new JsonException("esperado ':' apos a chave \"" + key + "\"");
            }
            i = skipWs(text, i + 1);

            if (i < text.length() && text.charAt(i) == '"') {
                out.put(key, parseString(text, i, end));
                i = end[0];
            } else if (text.startsWith("null", i)) {
                out.put(key, null);
                i += 4;
            } else {
                throw new JsonException("valor de \"" + key + "\" deve ser string ou null");
            }

            i = skipWs(text, i);
            if (i < text.length() && text.charAt(i) == ',') {
                i++;
                continue;
            }
            if (i < text.length() && text.charAt(i) == '}') {
                return out;
            }
            throw new JsonException("esperado ',' ou '}'");
        }
    }

    private static int skipWs(String s, int i) {
        while (i < s.length()) {
            char c = s.charAt(i);
            if (c == ' ' || c == '\t' || c == '\n' || c == '\r') {
                i++;
            } else {
                break;
            }
        }
        return i;
    }

    /** Lê uma string JSON a partir da aspa em {@code start}; grava em
     *  {@code end[0]} o índice logo após a aspa de fechamento. */
    private static String parseString(String s, int start, int[] end) {
        StringBuilder sb = new StringBuilder();
        int i = start + 1; // pula a aspa de abertura
        while (i < s.length()) {
            char c = s.charAt(i++);
            if (c == '"') {
                end[0] = i;
                return sb.toString();
            }
            if (c == '\\') {
                if (i >= s.length()) {
                    throw new JsonException("sequencia de escape incompleta");
                }
                char esc = s.charAt(i++);
                switch (esc) {
                    case '"':  sb.append('"'); break;
                    case '\\': sb.append('\\'); break;
                    case '/':  sb.append('/'); break;
                    case 'n':  sb.append('\n'); break;
                    case 't':  sb.append('\t'); break;
                    case 'r':  sb.append('\r'); break;
                    case 'b':  sb.append('\b'); break;
                    case 'f':  sb.append('\f'); break;
                    case 'u':
                        if (i + 4 > s.length()) {
                            throw new JsonException("escape \\u incompleto");
                        }
                        sb.append((char) Integer.parseInt(s.substring(i, i + 4), 16));
                        i += 4;
                        break;
                    default:
                        throw new JsonException("escape invalido: \\" + esc);
                }
            } else {
                sb.append(c);
            }
        }
        throw new JsonException("string nao terminada");
    }
}
