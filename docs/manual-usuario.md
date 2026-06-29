# Manual de Usuário — Sistema Runner

O Sistema Runner é composto por dois CLIs multiplataforma:

- **`assinatura`** — cria e valida assinaturas digitais (simuladas) invocando o
  `assinador.jar`, em modo local (`java -jar`) ou servidor (HTTP).
- **`simulador`** — gerencia o ciclo de vida do Simulador HubSaúde
  (`simulador.jar`): baixa, inicia, consulta e encerra.

> Instalação e verificação de integridade: veja o [Guia de Instalação](instalacao.md).

---

## Sumário

- [Conceitos e convenções](#conceitos-e-convenções)
- [CLI `assinatura`](#cli-assinatura)
  - [sign — criar assinatura](#sign--criar-assinatura)
  - [validate — validar assinatura](#validate--validar-assinatura)
  - [start — iniciar o servidor](#start--iniciar-o-servidor)
  - [status / stop](#status--stop)
  - [Assinar com dispositivo PKCS#11](#assinar-com-dispositivo-pkcs11)
- [CLI `simulador`](#cli-simulador)
- [Solução de problemas](#solução-de-problemas)

---

## Conceitos e convenções

**Modos de invocação do assinador** (US-01):

| Modo | Como ativar | Quando usar |
|------|-------------|-------------|
| **Servidor (padrão)** | comportamento default | várias chamadas — evita o *cold start* da JVM a cada uso |
| **Local** | flag `--local` | uso esporádico, scripts, ou sem servidor disponível |

No modo servidor, o CLI reutiliza uma instância ativa (verificada por *health
check*) ou inicia uma automaticamente. Se o servidor não puder subir ou a chamada
HTTP falhar, o CLI **degrada para o modo local** com um aviso em `stderr` — a
operação sempre produz um resultado.

**Contrato de saída** (ADR-0003): o resultado é um **JSON de uma linha** impresso
em `stdout`:

```json
{"signature":"MOCKED_SIGNATURE_BASE64_==","valid":true,"message":"Assinatura criada com sucesso"}
```

Mensagens de progresso e avisos vão para `stderr`, mantendo o `stdout` limpo para
*piping* (ex.: `... | jq`).

**Códigos de saída:**

| Código | Significado |
|--------|-------------|
| `0` | Operação bem-sucedida (`valid: true`) |
| `1` | Requisição processada, mas inválida — parâmetro inválido ou assinatura não confere (`valid: false`) |
| `2` | Falha de execução no modo local (ex.: `assinador.jar` ausente, `java` indisponível) |

**Validação:** quem valida os parâmetros é **exclusivamente o `assinador.jar`**
(autoridade única, critério E3). O CLI apenas repassa o que você informar.

**Estado e diretórios:** o estado dos processos em segundo plano (PID/porta) e os
logs ficam em **`~/.hubsaude/`** (`assinador.json`, `simulador.json`,
`assinador.log`, etc.). Portas padrão: **assinador `8080`**, **simulador `8443`**.

---

## CLI `assinatura`

```
assinatura [comando] [flags]

Comandos: sign · validate · start · stop · status · version
```

### sign — criar assinatura

Cria uma assinatura digital simulada para um conteúdo.

| Flag | Descrição |
|------|-----------|
| `--content <texto>` | Conteúdo a ser assinado |
| `--token <pin>` | Token de autenticação (PIN do dispositivo PKCS#11, quando aplicável) |
| `--local` | Força a invocação direta `java -jar` (sem modo servidor) |
| `--pkcs11-lib <caminho>` | Ativa o modo dispositivo (PKCS#11); implica `--local` — veja [seção](#assinar-com-dispositivo-pkcs11) |
| `--pkcs11-slot <n>` | Slot do dispositivo PKCS#11 (padrão: primeiro slot) |

```bash
# Modo servidor (padrão) — rápido em chamadas repetidas
assinatura sign --content "ola mundo"

# Conteúdo de um arquivo
assinatura sign --content "$(cat documento.txt)"

# Modo local explícito
assinatura sign --content "ola" --local
```

Saída (stdout):

```json
{"signature":"MOCKED_SIGNATURE_BASE64_==","valid":true,"message":"Assinatura criada com sucesso"}
```

### validate — validar assinatura

Verifica se uma assinatura corresponde ao conteúdo.

| Flag | Descrição |
|------|-----------|
| `--content <texto>` | Conteúdo original |
| `--signature <valor>` | Assinatura a ser validada |
| `--local` | Força a invocação direta `java -jar` |

```bash
assinatura validate --content "ola mundo" --signature "MOCKED_SIGNATURE_BASE64_=="
```

```json
{"signature":"MOCKED_SIGNATURE_BASE64_==","valid":true,"message":"Assinatura é válida"}
```

Uma assinatura que não confere retorna `"valid":false` com `exit 1` — a chamada
foi processada com sucesso, mas o resultado é negativo.

> `validate` **não** expõe flags PKCS#11: a validação não interage com o
> dispositivo criptográfico (ADR-0005).

### start — iniciar o servidor

Sobe o `assinador.jar` como servidor HTTP em segundo plano e aguarda ficar
pronto (`GET /health`). Reutiliza uma instância já ativa, se houver.

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `--port <n>` | `8080` | Porta do servidor |
| `--timeout <min>` | `0` | Minutos de inatividade até o auto-encerramento (`0` = desabilitado) |
| `--pkcs11-lib <caminho>` | — | Vincula um dispositivo PKCS#11 a esta instância |
| `--pkcs11-slot <n>` | primeiro | Slot do dispositivo PKCS#11 |

```bash
assinatura start                  # porta 8080
assinatura start --port 9090      # porta alternativa
assinatura start --timeout 10     # encerra após 10 min sem requisições
```

O auto-encerramento por inatividade reinicia a contagem a cada requisição de
`sign`/`validate` (probes de `/health` não contam).

### status / stop

```bash
assinatura status        # informa se há instância em execução e respondendo
assinatura stop          # encerra a instância registrada (POST /shutdown)
assinatura stop --port 9090   # encerra a instância de uma porta específica
```

`stop` é idempotente: sem instância registrada, informa que já estava parado.

### Assinar com dispositivo PKCS#11

O `assinador.jar` interage de fato com um token/smart card via a interface padrão
**PKCS#11** (provider `SunPKCS11` do JDK). A *interação* com o dispositivo é real
(login no token com o PIN); o **valor da assinatura permanece simulado**, conforme
o escopo do projeto ([ADR-0005](adr/0005-pkcs11-dispositivo-criptografico.md)).

O recurso é **opcional** — só é acionado com `--pkcs11-lib`. O **PIN** vai em
`--token`. Como a configuração do dispositivo é *por invocação*, há dois caminhos:

**Por invocação (modo local).** Usar `--pkcs11-lib` no `sign` implica modo local:

```bash
assinatura sign --content "documento" \
  --token 1234 \
  --pkcs11-lib /usr/lib/softhsm/libsofthsm2.so --pkcs11-slot 0
```

**Vinculado ao servidor.** Inicie o servidor com o dispositivo; as chamadas
`sign` subsequentes (modo servidor) usam o token, informando apenas o PIN:

```bash
assinatura start --pkcs11-lib /usr/lib/softhsm/libsofthsm2.so
assinatura sign  --content "documento" --token 1234
```

Em caso de sucesso, a mensagem evidencia a interação:

```json
{"signature":"MOCKED_SIGNATURE_BASE64_==","valid":true,"message":"assinatura simulada após interação com dispositivo PKCS#11 (certificado: CN=...) [provider SunPKCS11-Runner]"}
```

Falha no dispositivo (biblioteca inexistente, PIN incorreto, token ausente) é um
erro de sistema, distinto de erro de parâmetro (critério E3):

```json
{"signature":null,"valid":false,"message":"dispositivo PKCS#11: falha ao interagir com o token: ..."}
```

---

## CLI `simulador`

Gerencia o Simulador HubSaúde (`simulador.jar`), baixado sob demanda do
`release.json` da disciplina. Estado persistido em `~/.hubsaude/`.

```
simulador [comando]

Comandos: start · status · stop · version
```

### start

Baixa o `simulador.jar` (se ainda não estiver em cache na versão mais recente),
verifica se a porta **8443** está livre, inicia o processo em segundo plano e
aguarda ficar pronto (`GET /api/info`).

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `--source <url>` | `release.json` da disciplina | URL alternativa do `release.json` |

```bash
simulador start
```

### status / stop

```bash
simulador status     # mostra se está em execução (consulta /api/info)
simulador stop       # encerra a instância (POST /shutdown)
```

---

## Solução de problemas

| Sintoma | Causa provável | O que fazer |
|---------|----------------|-------------|
| `assinador.jar nao encontrado` (exit 2) | JAR não está ao lado do binário | Coloque `assinador.jar` no diretório do executável ou defina `ASSINADOR_JAR` |
| Primeira execução lenta | Download do JRE Temurin para `~/.hubsaude/` | Aguarde; chamadas seguintes reutilizam o JRE |
| `porta N nao disponivel` no `start` | Outra aplicação usa a porta | Use `--port` (assinatura) ou libere a porta 8443 (simulador) |
| `assinador registrado ... nao responde` | Instância órfã/travada | Rode `assinatura stop` e inicie de novo |
| `aviso: modo servidor indisponivel ...` | Servidor não subiu (ex.: sem Java) | A operação seguiu em modo local; verifique o Java/logs em `~/.hubsaude/assinador.log` |
| `dispositivo PKCS#11: ...` (exit 1) | Biblioteca/slot/PIN do token incorretos | Confira `--pkcs11-lib`, `--pkcs11-slot` e o `--token` (PIN) |
| SmartScreen/Gatekeeper bloqueia o binário | Executável não notarizado | Veja as notas por plataforma no [Guia de Instalação](instalacao.md) |

Para detalhes de qualquer comando, use `--help`:

```bash
assinatura sign --help
simulador start --help
```
