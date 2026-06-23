package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignatureResponse;
import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

class JsonTest {

    @Test
    void parseObjetoSimples() {
        Map<String, String> m = Json.parseObject("{\"content\":\"abc\",\"token\":\"pin\"}");
        assertEquals("abc", m.get("content"));
        assertEquals("pin", m.get("token"));
    }

    @Test
    void parseValorNull() {
        Map<String, String> m = Json.parseObject("{\"signature\":null}");
        assertTrue(m.containsKey("signature"));
        assertNull(m.get("signature"));
    }

    @Test
    void parseObjetoVazio() {
        assertTrue(Json.parseObject("{}").isEmpty());
        assertTrue(Json.parseObject("  {  }  ").isEmpty());
    }

    @Test
    void parsePreservaAcentosEEscapes() {
        Map<String, String> m = Json.parseObject(
            "{\"content\":\"olá \\\"mundo\\\"\\n\\tfim\"}");
        assertEquals("olá \"mundo\"\n\tfim", m.get("content"));
    }

    @Test
    void parseUnicodeEscape() {
        Map<String, String> m = Json.parseObject("{\"content\":\"\\u00e1gua\"}");
        assertEquals("água", m.get("content"));
    }

    @Test
    void parseIgnoraEspacosEntreTokens() {
        Map<String, String> m = Json.parseObject("{ \"content\" : \"x\" , \"token\" : \"y\" }");
        assertEquals("x", m.get("content"));
        assertEquals("y", m.get("token"));
    }

    @Test
    void parseRejeitaJsonMalformado() {
        assertThrows(Json.JsonException.class, () -> Json.parseObject("{"));
        assertThrows(Json.JsonException.class, () -> Json.parseObject("nao json"));
        assertThrows(Json.JsonException.class, () -> Json.parseObject("{\"k\":}"));
        assertThrows(Json.JsonException.class, () -> Json.parseObject("{\"k\" \"v\"}"));
        assertThrows(Json.JsonException.class, () -> Json.parseObject("{\"k\":\"v\""));
    }

    @Test
    void toJsonSerializaResposta() {
        String s = Json.toJson(new SignatureResponse("SIG==", true, "ok"));
        assertEquals("{\"signature\":\"SIG==\",\"valid\":true,\"message\":\"ok\"}", s);
    }

    @Test
    void toJsonSerializaSignatureNull() {
        String s = Json.toJson(new SignatureResponse(null, false, "erro"));
        assertEquals("{\"signature\":null,\"valid\":false,\"message\":\"erro\"}", s);
    }

    @Test
    void toJsonEscapaCaracteresEspeciais() {
        String s = Json.toJson(new SignatureResponse(null, false, "aspas \" e\nquebra"));
        assertTrue(s.contains("\\\""), "aspas devem ser escapadas: " + s);
        assertTrue(s.contains("\\n"), "quebra de linha deve ser escapada: " + s);
    }

    @Test
    void toJsonERoundTripDePreservaAcentos() {
        String json = Json.toJson(new SignatureResponse(null, false, "parâmetro inválido"));
        // acentos não são escapados (>= 0x20), seguem literais
        assertTrue(json.contains("parâmetro inválido"), json);
    }
}
