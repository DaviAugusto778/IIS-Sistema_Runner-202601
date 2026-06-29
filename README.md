# Sistema Runner

**Disciplina:** Implementação e Integração — BES/UFG (2026-1)  
**Aluno:** `Davi Augusto Fererira de Carvalho`  
**Matrícula:** `202403066`

---

## O que é

O Sistema Runner é uma ferramenta de linha de comandos (CLI) multiplataforma que permite executar aplicações Java — especificamente o **assinador.jar** e o **Simulador HubSaúde** — sem que o usuário precise instalar ou configurar o JDK manualmente. O sistema detecta (ou baixa) o Java necessário, invoca as aplicações com os parâmetros corretos e apresenta os resultados de forma legível no terminal.

O contexto é a plataforma **HubSaúde**, um esforço conjunto da SES-GO e UFG para interoperabilidade de dados em saúde.

## Documentação de referência

Os documentos oficiais do trabalho estão no repositório do professor, fixados no commit `4d7d40f` para garantir rastreabilidade:

| Documento | Descrição |
|-----------|-----------|
| [especificacao.md](https://github.com/kyriosdata/runner/blob/4d7d40f/especificacao.md) | Épicos, critérios de aceitação e escopo |
| [docs/plano-revisitado-v2.md](https://github.com/kyriosdata/runner/blob/4d7d40f/docs/plano-revisitado-v2.md) | Histórias derivadas e sprints |
| [design.md](https://github.com/kyriosdata/runner/blob/4d7d40f/design.md) | Modelo C4 (contexto e contêineres) |
| [docs/criterios.md](https://github.com/kyriosdata/runner/blob/4d7d40f/docs/criterios.md) | Critérios de aceitação (definição de "pronto") |
| [release.json](https://github.com/kyriosdata/runner/blob/4d7d40f/release.json) | Manifesto dos artefatos validador/simulador + JREs |


## Épicos

| Épico | Descrição | Situação |
|-------|-----------|----------|
| US-01 | Invocar assinador.jar via CLI (modo local e modo servidor HTTP) | ✅ Concluído |
| US-02 | Simular assinatura digital com validação rigorosa de parâmetros + PKCS#11 | ✅ Concluído |
| US-03 | Gerenciar ciclo de vida do Simulador HubSaúde via CLI | ✅ Concluído |
| US-04 | Provisionar JDK automaticamente (detecção e download) | ✅ Concluído |
| US-05 | Disponibilizar binários multiplataforma via GitHub Releases | ✅ Concluído |

## Documentação

| Documento | Para quê |
|-----------|----------|
| [**Guia de Instalação**](docs/instalacao.md) | Baixar, verificar (SHA256 + Cosign) e instalar nas três plataformas |
| [**Manual de Usuário**](docs/manual-usuario.md) | Comandos, flags, modos local/servidor, PKCS#11, exit codes e troubleshooting |
| [ADRs](docs/adr/) | Decisões de arquitetura (Go+Java, modo local vs servidor, contrato CLI↔JAR, servidor HTTP, PKCS#11) |

## Como rodar

O guia abaixo é um resumo; consulte o [Manual de Usuário](docs/manual-usuario.md)
e o [Guia de Instalação](docs/instalacao.md) para os detalhes.

### Pré-requisitos

- Go 1.25 ou superior (para compilar a CLI)
- JDK 21 e Maven (para compilar o `assinador.jar`)
- Em runtime, o usuário final **não** precisa instalar Java: o CLI detecta um JDK 21 ou provisiona um JRE Temurin automaticamente (US-04)

### Compilar localmente

```bash
# Compilar o assinador.jar (a partir da raiz do repo)
cd assinador && ./mvnw package && cd ..

# Compilar a CLI assinatura
go build -o assinatura ./

# Colocar o jar ao lado do binário (ou exportar ASSINADOR_JAR=<caminho>)
cp assinador/target/assinador.jar .
```

### Executar o CLI `assinatura`

```bash
# Exibir versão
./assinatura version

# Criar assinatura (modo local)
./assinatura sign --content "texto a assinar"

# Validar assinatura
./assinatura validate --content "texto a assinar" --signature "<assinatura>"

# Iniciar modo servidor HTTP (planejado para Sprint 3 — US-01.5+)
# ./assinatura start
```

### Assinar com dispositivo criptográfico (PKCS#11 — US-02.5)

O `assinador.jar` interage de fato com um token/smart card via a interface
padrão **PKCS#11** (provider `SunPKCS11` do JDK, sem dependências externas). A
*interação* com o dispositivo é real (login no token com o PIN); o **valor da
assinatura permanece simulado**, conforme o escopo do projeto (ver
[ADR-0005](docs/adr/0005-pkcs11-dispositivo-criptografico.md)).

O recurso é **opcional**: só é acionado com `--pkcs11-lib`. Sem ele, o fluxo
padrão segue inalterado. O PIN é passado por `--token`.

Pelo CLI `assinatura` (o dispositivo é por invocação, então `--pkcs11-lib`
implica modo local):

```bash
# Assina interagindo com o dispositivo (ex.: SoftHSM2)
./assinatura sign --content "documento" \
  --token 1234 \
  --pkcs11-lib /usr/lib/softhsm/libsofthsm2.so --pkcs11-slot 0

# Modo servidor: o dispositivo fica vinculado à instância iniciada
./assinatura start --pkcs11-lib /usr/lib/softhsm/libsofthsm2.so
./assinatura sign --content "documento" --token 1234   # usa o dispositivo do servidor
```

Invocando o JAR diretamente (equivalente, sem o CLI):

```bash
java -jar assinador.jar sign --content "documento" \
  --token 1234 --pkcs11-lib /usr/lib/softhsm/libsofthsm2.so --pkcs11-slot 0
java -jar assinador.jar server --pkcs11-lib /usr/lib/softhsm/libsofthsm2.so
```

O teste de integração que comprova chamadas PKCS#11 reais (critério E5) é
condicional: roda apenas quando há um dispositivo provisionado, apontado pelas
variáveis `PKCS11_LIBRARY`/`PKCS11_PIN` (instruções em `Pkcs11IntegrationTest`).

### Executar o CLI `simulador`

```bash
# A preencher na Sprint 4
# simulador start
# simulador stop
# simulador status
```

### Verificar integridade dos binários baixados

Cada artefato publicado na release acompanha um `.sha256` (integridade) e os arquivos
`.sig` + `.pem` (autenticidade via Cosign keyless). Os nomes seguem o padrão
`assinatura-<tag>-<os>-<arch>` (ex.: `assinatura-v1.0.0-linux-amd64`).

```bash
# Verificar checksum SHA256
sha256sum -c assinatura-v1.0.0-linux-amd64.sha256

# Verificar assinatura Cosign (keyless / OIDC).
# Os binários são assinados pelo workflow release.yml via OIDC do GitHub Actions,
# então a verificação exige a identidade do assinante e o emissor do token.
cosign verify-blob \
  --certificate           assinatura-v1.0.0-linux-amd64.pem \
  --signature             assinatura-v1.0.0-linux-amd64.sig \
  --certificate-identity-regexp "^https://github.com/DaviAugusto778/IIS-Sistema_Runner-202601/.github/workflows/release.yml@refs/tags/v.*$" \
  --certificate-oidc-issuer     "https://token.actions.githubusercontent.com" \
  assinatura-v1.0.0-linux-amd64
```

### Publicar uma release (mantenedores)

O workflow [`.github/workflows/release.yml`](.github/workflows/release.yml) dispara ao
empurrar uma tag SemVer. Ele compila os seis binários (assinatura + simulador para
Linux, Windows e macOS amd64) e o `assinador.jar` (independente de plataforma), gera
os checksums SHA256, assina cada artefato com Cosign (keyless) e cria a GitHub Release
com todos os arquivos anexados. O `assinador.jar` é publicado com esse nome exato — basta
colocá-lo ao lado do binário `assinatura` para que seja detectado automaticamente.

```bash
git tag v1.0.0
git push origin v1.0.0
```

A versão exibida por `version` vem da tag (injetada via `-ldflags`); não há versão
hardcoded no código nem auto-commit/auto-tag — a release é sempre intencional.

## Status do projeto

| Sprint | Foco | Histórias | Estado |
|--------|------|-----------|--------|
| 1 | Fundação: estrutura Go, CI/CD, releases, Cosign | US-01.1, US-05.1, US-05.2, US-05.3 | ✅ Concluída |
| 2 | Fluxo ponta-a-ponta local: sign/validate via `java -jar` | US-02.1–02.3, US-01.2–01.4, US-04.1 | ✅ Concluída  |
| 3 | Modo servidor HTTP e dispositivo criptográfico (PKCS#11) | US-02.4–02.5, US-01.5–01.9 | ✅ Concluída |
| 4 | CLI simulador e download dinâmico do simulador.jar | US-03.1–03.4 | ✅ Concluída |

## Licença

Distribuído sob a [Licença MIT](LICENSE).
