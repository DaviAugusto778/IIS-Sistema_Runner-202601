# ADR 0002 — Suporte a dois modos de invocação: local e servidor HTTP

- **Data:** 2026-05-19
- **Status:** Aceito

## Contexto

O CLI `assinatura` precisa invocar o `assinador.jar` para criar e validar assinaturas.
A JVM tem um custo de inicialização mensurável: cada invocação no modo direto (`java -jar`)
incorre em ~200–500 ms só para inicializar o runtime, antes de qualquer trabalho real.
Para um único comando esporádico isso é aceitável; para scripts que chamam o assinador
dezenas de vezes em sequência, o overhead acumula e a experiência degrada.

A especificação (US-01) descreve dois modos de operação e define que o sistema deve
suportar ambos. A decisão aqui é como implementar a escolha entre eles e qual usar por
padrão.

## Decisão

Implementar ambos os modos e **usar o servidor HTTP por padrão**: o CLI reutiliza uma
instância ativa ou **inicia uma automaticamente** quando ausente, e então faz a requisição
via HTTP. O modo local (`java -jar`) é forçado explicitamente com `--local`, e também serve
de **fallback** quando o servidor não pode subir ou a chamada HTTP falha (com aviso em
stderr) — a operação sempre produz um resultado.

> **Nota (2026-06-23, US-01.6):** a redação original desta seção era ambígua — citava
> "fallback automático para local quando nenhuma instância estiver em execução", o que
> conflitava com o fluxograma abaixo ("não → inicia servidor → POST"). A ambiguidade foi
> resolvida em favor do **auto-start** (o fluxograma): ausência de servidor dispara o
> startup automático, não o modo local. O modo local permanece como override (`--local`) e
> como degradação em caso de falha do servidor.

**Modo local (invocação direta):**
O CLI executa `java -jar assinador.jar <args>` via `os/exec`, aguarda o término e captura
a saída. Cada chamada paga o custo de inicialização da JVM (*cold start*).

**Modo servidor (invocação via HTTP):**
O `assinador.jar` é iniciado uma vez e permanece em execução, respondendo a requisições
`POST /sign` e `POST /validate`. O CLI detecta se há uma instância ativa consultando
`~/.hubsaude/assinador.pid` e fazendo um health check HTTP; se encontrar, envia a
requisição diretamente (*warm start*); se não encontrar, inicia o servidor antes de
prosseguir.

**Comportamento padrão resumido:**
```
assinatura sign ...
  └─ instância em execução? ──sim──► POST /sign
                             └─não──► inicia servidor → POST /sign
```

O modo local pode ser forçado explicitamente com a flag `--local` para cenários de
automação que não devem depender de estado externo.

## Justificativa

### Por que suportar os dois modos

Cada modo tem um caso de uso legítimo e mutuamente exclusivo com o outro:

| Cenário | Modo adequado |
|---------|---------------|
| Uma assinatura ocasional, script simples | Local — sem dependência de processo em background |
| Múltiplas assinaturas em sequência (integração) | Servidor — elimina ~N×400 ms de cold start |
| Ambiente CI/CD efêmero | Local — sem estado persistente entre execuções |
| Uso interativo contínuo | Servidor — resposta imediata após a primeira chamada |

Forçar um único modo descartaria um dos dois casos de uso sem ganho arquitetural.

### Por que servidor como padrão

O caso de uso mais comum descrito na especificação (integrador da plataforma HubSaúde)
envolve múltiplas operações em uma sessão de trabalho. O modo servidor entrega latência
perceptivelmente menor sem exigir nenhuma ação extra do usuário — a primeira chamada
inicia o servidor automaticamente.

### Por que fallback automático em vez de erro

Exigir que o usuário inicie o servidor manualmente antes de qualquer comando quebraria a
premissa central do sistema: "o usuário não precisa conhecer detalhes técnicos". O CLI
deve ser a única interface necessária.

## Alternativas consideradas

**Apenas modo local:** mais simples de implementar (sem gestão de processos em background,
sem PID files, sem HTTP server no .jar), mas impõe latência acumulada inaceitável para
uso intensivo. Descartado.

**Apenas modo servidor:** elimina o cold start em todos os casos, mas adiciona estado
obrigatório (o servidor precisa estar rodando). Torna o uso em CI/CD frágil e quebra
cenários de uso único. Descartado.

**Modo servidor sem fallback automático (erro se não iniciado):** mais previsível do
ponto de vista de operações, mas viola o princípio de não exigir conhecimento técnico do
usuário. Descartado.

**gRPC em vez de HTTP:** menor overhead de serialização, mas adiciona complexidade de
setup (protobuf, geração de código) desproporcional ao volume de chamadas esperado.
HTTP com JSON é suficiente e mais fácil de inspecionar durante desenvolvimento.

## Consequências

**Positivas:**
- Experiência do usuário é consistente: o mesmo comando funciona independentemente do
  estado do sistema.
- Modo servidor elimina overhead de JVM para uso intensivo sem configuração manual.
- Modo local (`--local`) permite uso previsível e sem estado em automações.

**Negativas:**
- O CLI precisa gerenciar ciclo de vida de processo externo: PID files em
  `~/.hubsaude/`, health check HTTP, lógica de inicialização e encerramento. Isso
  representa ~4 histórias adicionais (US-01.5 a US-01.8) que não existiriam no modelo
  de modo único.
- O `assinador.jar` precisa de um servidor HTTP embutido (US-02.4), o que introduz
  decisão de biblioteca e possivelmente nova dependência no `pom.xml`.
- Concorrência: se dois processos CLI tentarem iniciar o servidor simultaneamente há
  risco de conflito de porta. Mitigação: verificar a porta antes de iniciar e usar o
  arquivo `.pid` como lock implícito.

## Arquivos de estado em runtime

```
~/.hubsaude/
├── assinador.pid     ← PID e porta da instância em execução
└── jdk/              ← JRE provisionado (ver ADR 0001)
```

O diretório `~/.hubsaude/` é criado pelo CLI na primeira execução. Sua existência não
é pré-condição para nenhum comando — ausência do `.pid` equivale a "sem instância ativa".
