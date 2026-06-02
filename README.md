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
| US-01 | Invocar assinador.jar via CLI (modo local e modo servidor HTTP) | 🔄 Em andamento |
| US-02 | Simular assinatura digital com validação rigorosa de parâmetros | 🔄 Em andamento |
| US-03 | Gerenciar ciclo de vida do Simulador HubSaúde via CLI | ⏳ Pendente |
| US-04 | Provisionar JDK automaticamente (detecção e download) | 🔄 Em andamento |
| US-05 | Disponibilizar binários multiplataforma via GitHub Releases | ✅ Concluído |

## Como rodar

> Esta seção será preenchida conforme a implementação avança.

### Pré-requisitos

- Go 1.25 ou superior (para compilar a CLI)
- JDK 21 e Maven (para compilar o `assinador.jar`)
- Em runtime, o usuário final só precisa de Java 21 instalado (ou aguardar US-04, que provisiona JRE automaticamente)

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

### Executar o CLI `simulador`

```bash
# A preencher na Sprint 4
# simulador start
# simulador stop
# simulador status
```

### Verificar integridade dos binários baixados

```bash
# Verificar checksum SHA256
sha256sum -c assinatura-<versão>-<os>-<arch>.sha256

# Verificar assinatura Cosign
cosign verify-blob \
  --certificate assinatura-<versão>-<os>-<arch>.pem \
  --signature  assinatura-<versão>-<os>-<arch>.sig \
  assinatura-<versão>-<os>-<arch>
```

## Status do projeto

| Sprint | Foco | Histórias | Estado |
|--------|------|-----------|--------|
| 1 | Fundação: estrutura Go, CI/CD, releases, Cosign | US-01.1, US-05.1, US-05.2, US-05.3 | ✅ Concluída |
| 2 | Fluxo ponta-a-ponta local: sign/validate via `java -jar` | US-02.1–02.3, US-01.2–01.4, US-04.1 | ✅ Concluída  |
| 3 | Modo servidor HTTP e dispositivo criptográfico (PKCS#11) | US-02.4–02.5, US-01.5–01.9 | ⏳ Pendente |
| 4 | CLI simulador e download dinâmico do simulador.jar | US-03.1–03.4 | ⏳ Pendente |

## Licença

Distribuído sob a [Licença MIT](LICENSE).
