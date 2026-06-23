# ADR-0004 — Servidor HTTP do assinador.jar com `com.sun.net.httpserver`

**Status:** Aceito
**Data:** 2026-06-23
**Relacionado:** [ADR-0002](0002-modo-local-vs-servidor.md) (modo local vs servidor), [ADR-0003](0003-contrato-cli-jar.md) (contrato JSON)

## Contexto

A US-02.4 exige que o `assinador.jar` exponha `POST /sign` e `POST /validate`
para o modo servidor descrito no ADR-0002. O ADR-0002 já antecipava que isso
"introduz decisão de biblioteca e possivelmente nova dependência no `pom.xml`".

Até aqui o `assinador.jar` é um CLI puro, empacotado com o maven-shade-plugin,
sem nenhum framework. O `pom.xml` tem apenas JUnit (escopo de teste). O CLAUDE.md
define duas restrições relevantes: **preferir a biblioteca padrão** quando ela
resolve, e **não criar dependências novas sem consulta**.

## Decisão

Implementar o servidor com **`com.sun.net.httpserver.HttpServer`**, embutido no
JDK (módulo `jdk.httpserver`, API suportada e exportada), sem nenhuma dependência
nova no `pom.xml`.

- Roteamento via `server.createContext("/sign" | "/validate" | "/shutdown", ...)`.
- Leitura/escrita de JSON por um utilitário próprio mínimo (`Json`), suficiente
  para os payloads planos (`content`/`signature`/`token`) — sem Jackson/Gson.
- Respostas seguem o contrato de uma linha do ADR-0003
  (`{"signature":...,"valid":...,"message":...}`), com
  `Content-Type: application/json; charset=utf-8`.
- A mesma lógica de validação e simulação do modo CLI é reutilizada
  (`SignatureService`, `SignRequestValidator`, `ValidateRequestValidator`) —
  o JAR continua sendo autoridade única de validação (critério E3).

**Códigos de status:**

| Situação | Status |
|----------|--------|
| Requisição bem formada e processada (mesmo `valid:false`) | `200` |
| JSON malformado ou parâmetros inválidos | `400` |
| Método diferente de POST | `405` |

`POST /shutdown` encerra o servidor graciosamente (consumido pela US-01.8 do CLI).

## Justificativa

- **Zero dependências novas:** mantém o JAR leve (segue leve o shade), reduz
  superfície de manutenção/segurança e respeita a política std-lib-first.
- **Baixa latência:** o objetivo do modo servidor (ADR-0002) é eliminar o cold
  start da JVM. Um servidor embutido sobe em milissegundos; Spring Boot
  adicionaria segundos de inicialização e dezenas de MB ao artefato, contrariando
  a própria motivação do modo servidor.
- **Suficiência:** os payloads são triviais e o volume de endpoints é pequeno
  (3 contextos). Roteamento e (de)serialização à mão são simples e testáveis.

## Alternativas consideradas

- **Spring Boot:** casa com o termo `SignatureController` da especificação e dá
  `@RestController`/JSON automático, mas infla o JAR (~20 MB+), aumenta o cold
  start e introduz dependência pesada — desproporcional para 3 endpoints.
  Descartado.
- **Javalin / Spark / Helidon SE:** mais leves que Spring, porém ainda assim
  dependências externas para um problema que a std lib resolve. Descartado.
- **gRPC:** já descartado no ADR-0002 (complexidade de protobuf desproporcional).

## Consequências

**Positivas**
- Nenhuma dependência nova; build e distribuição inalterados.
- Endpoints reaproveitam a validação do modo CLI — uma fonte de verdade.
- `Json` reutilizável também pelo modo CLI (o `Main` passou a serializar por ele).

**Negativas**
- (De)serialização manual de JSON: o utilitário `Json` cobre apenas objetos
  planos string→string|null. Payloads mais ricos exigiriam estendê-lo (ou, aí
  sim, reavaliar uma biblioteca).
- `com.sun.net.httpserver` é de baixo nível: roteamento, métodos e tratamento de
  erro são responsabilidade do código (mitigado por testes de integração em
  `HttpSignatureServerTest`).

## Testes que cobrem esta decisão

- `HttpSignatureServerTest`: `/sign` e `/validate` (sucesso e 400 de validação),
  `/validate` com assinatura errada (200, `valid:false`), JSON malformado (400),
  método errado (405), `Content-Type`, e `/shutdown` encerrando o servidor.
- `JsonTest`: parse de objetos, valores null, escapes, unicode, acentos e
  rejeição de JSON malformado; serialização no contrato do ADR-0003.
