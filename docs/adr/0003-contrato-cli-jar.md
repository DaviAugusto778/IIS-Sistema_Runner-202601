# ADR-0003 — Contrato entre a CLI `assinatura` e o `assinador.jar`

**Status:** Aceito
**Data:** 2026-06-02
**Critérios afetados:** [A](https://github.com/kyriosdata/runner/blob/4d7d40f/docs/criterios.md#a-princípios-transversais), [D](https://github.com/kyriosdata/runner/blob/4d7d40f/docs/criterios.md#d-qualidade-de-código), [E1](https://github.com/kyriosdata/runner/blob/4d7d40f/docs/criterios.md#e1-invocação-local-do-assinadorjar), [E3](https://github.com/kyriosdata/runner/blob/4d7d40f/docs/criterios.md#e3-validação-de-parâmetros)

## Contexto

A CLI `assinatura` (Go) invoca o `assinador.jar` (Java) via subprocess. A fronteira entre os dois processos é uma API — argumentos de linha de comando, formato de saída e códigos de retorno. O critério D da especificação exige que esta fronteira esteja "documentada e testada, não inferida". O critério E3 estabelece o JAR como autoridade única de validação.

## Decisão

### 1. Argumentos de invocação

A CLI invoca o JAR sempre como:

```
java -jar <caminho-do-jar> <subcomando> [--flag valor]...
```

Subcomandos suportados: `sign`, `validate`.

Flags repassadas ao JAR:

| Subcomando | Flag         | Tipo   | Observação |
|------------|--------------|--------|------------|
| `sign`     | `--content`  | string | Conteúdo a assinar |
| `sign`     | `--token`    | string | Token opcional |
| `validate` | `--content`  | string | Conteúdo original |
| `validate` | `--signature`| string | Assinatura a verificar |

A CLI só repassa uma flag quando o usuário a forneceu explicitamente (via `cmd.Flags().Changed`). **Não há validação de obrigatoriedade no nível Cobra** — a autoridade única é o JAR (critério E3).

### 2. Streams

- **stdout do JAR**: resultado da operação, JSON único de uma linha:
  ```json
  {"signature":"...","valid":true|false,"message":"..."}
  ```
- **stderr do JAR**: diagnóstico (erros inesperados, traces). Nunca contém o resultado.

A CLI **não parseia o JSON**. Faz passthrough fiel:
- escreve o stdout do JAR no stdout do processo CLI;
- escreve o stderr do JAR no stderr do processo CLI;
- preserva exit code do JAR como exit code da CLI.

### 3. Códigos de saída

| Exit code | Significado                                    | Origem |
|-----------|------------------------------------------------|--------|
| `0`       | Sucesso (assinatura criada ou válida)          | JAR    |
| `1`       | Falha de domínio (`valid=false` no payload)    | JAR    |
| `2`       | Falha de execução (jar ausente, java fora do PATH, falha de I/O) | CLI |

Exit codes ≥ 3 estão reservados para uso futuro pelo JAR.

### 4. Localização do JAR

A CLI procura o JAR nesta ordem:

1. Caminho em `ASSINADOR_JAR` (override explícito);
2. Diretório do próprio executável (`os.Executable()`).

Não há fallback para o diretório de trabalho atual — isso garante que a CLI funcione independente do diretório de invocação (critério E1).

### 5. Preservação de caracteres

A CLI repassa os valores das flags via `exec.Command(name, args...)`, que no Go usa o argv do sistema operacional sem reparsing de shell. Espaços, acentos e aspas são preservados (verificado em `invoker_test.go`).

## Consequências

**Positivas**
- Validação concentrada no JAR — uma fonte de verdade.
- Streams separados permitem que ferramentas downstream consumam apenas o JSON sem misturar diagnóstico.
- Exit codes permitem orquestração em scripts.

**Negativas**
- A CLI atual não humaniza o output (mostra JSON cru). Mitigação futura: flag `--output=human|json` em US-01.6.
- Mudanças no JAR (novas flags, novos campos do JSON) exigem atualização coordenada deste contrato.

## Testes que cobrem este contrato

- `internal/invoker/invoker_test.go`:
  - separação stdout/stderr (`TestInvoke_SeparatesStdoutFromStderr`)
  - propagação de exit code não-zero (`TestInvoke_NonZeroExitPreservesPayload`)
  - jar ausente (`TestInvoke_MissingJarReturnsError`)
  - preservação de espaços/acentos/aspas (`TestInvoke_PreservesArgumentsWithSpacesAndAccents`)
- `cmd/cmd_test.go`:
  - flags não são marcadas como required (`TestSignHasOptionalFlags`, `TestValidateHasOptionalFlags`)
- `assinador/src/test/java/.../FakeSignatureServiceTest.java`:
  - validação de `content` e `signature` ausentes/vazios.
