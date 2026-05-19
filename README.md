# Sistema Runner

**Disciplina:** Implementação e Integração — BES/UFG (2026-1)  
**Aluno:** `Davi Augusto Fererira de Carvalho`  
**Matrícula:** `202403066`

---

## O que é

O Sistema Runner é uma ferramenta de linha de comandos (CLI) multiplataforma que permite executar aplicações Java — especificamente o **assinador.jar** e o **Simulador HubSaúde** — sem que o usuário precise instalar ou configurar o JDK manualmente. O sistema detecta (ou baixa) o Java necessário, invoca as aplicações com os parâmetros corretos e apresenta os resultados de forma legível no terminal.

O contexto é a plataforma **HubSaúde**, um esforço conjunto da SES-GO e UFG para interoperabilidade de dados em saúde.

## Documentação de referência

Os documentos oficiais do trabalho estão no repositório do professor:

| Documento | Descrição |
|-----------|-----------|
| [especificacao.md](github.com/kyriosdata/runner/blob/main/especificacao.md) | Épicos, critérios de aceitação e escopo |
| [docs/plano-revisitado-v2.md](github.com/kyriosdata/runner/blob/main/docs/plano-revisitado-v2.md) | Histórias derivadas e sprints |
| [design.md](github.com/kyriosdata/runner/blob/main/design.md) | Modelo C4 (contexto e contêineres) |


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

```
# A preencher: versão do Go, dependências externas, etc.
```

### Compilar localmente

```bash
# A preencher
```

### Executar o CLI `assinatura`

```bash
# Exibir versão
assinatura version

# Criar assinatura (a preencher conforme Sprint 2)
# assinatura sign --content <...>

# Validar assinatura (a preencher conforme Sprint 2)
# assinatura validate --content <...> --signature <...>

# Iniciar modo servidor (a preencher conforme Sprint 3)
# assinatura start
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
| 2 | Fluxo ponta-a-ponta local: sign/validate via `java -jar` | US-02.1–02.3, US-01.2–01.4, US-04.1 | 🔄 Em andamento |
| 3 | Modo servidor HTTP e dispositivo criptográfico (PKCS#11) | US-02.4–02.5, US-01.5–01.9 | ⏳ Pendente |
| 4 | CLI simulador e download dinâmico do simulador.jar | US-03.1–03.4 | ⏳ Pendente |

## Licença

A definir.
