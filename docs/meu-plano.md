# Meu Plano de Implementação

Adaptação pessoal do `plano-revisitado-v2.md` do professor. Referência original em
`../runner-referencia/runner-referencia/docs/plano-revisitado-v2.md`.

---

## Situação de partida

O repositório de referência entrega Sprint 1 completa e **US-02.1** da Sprint 2.
Isso significa que posso começar do código já existente para as histórias marcadas
como **"adaptar"** abaixo — estudar, entender, e replicar no meu repositório com
as devidas adaptações.

| O que existe no repo de referência | Status |
|------------------------------------|--------|
| CLI `assinatura` com `version` (Cobra) | ✅ |
| CI/CD: cross-compile + SHA256 + Cosign + GitHub Releases | ✅ |
| `SignatureService` + `FakeSignatureService` + DTOs Java | ✅ |
| Testes JUnit 5 do `FakeSignatureService` | ✅ |
| **Qualquer coisa além disso** | ❌ |

---

## Definition of Done

### Histórias de código Go

- [ ] `go build ./...` sem erros
- [ ] `go vet ./...` sem warnings
- [ ] `go test ./...` passa (incluindo o novo teste)
- [ ] Todo novo comando tem `--help` com descrição útil
- [ ] Cada função pública tem pelo menos um teste: caminho feliz + um caso de erro
- [ ] Nenhum `panic` sem recover em código de produção

### Histórias de código Java

- [ ] `mvn compile` sem erros
- [ ] `mvn test` passa
- [ ] Cada método validador tem testes para: parâmetro nulo, parâmetro vazio, formato inválido, e caso válido
- [ ] Erros retornam em `SignatureResponse` (campo `message`), nunca como exceção não tratada
- [ ] Nenhuma dependência nova adicionada ao `pom.xml` sem justificativa documentada

---

## Mapa de dependências

```
US-02.1 ──► US-02.2 ──► US-01.3 ──► US-01.4
              │                        ▲
            US-02.3 ──────────────────┘
              
US-01.2 ──────────────► US-01.3

US-04.1 ──────────────► US-01.3
              
US-02.4 ──► US-01.5 ──► US-01.6
              │
              └────────► US-01.7 ──► US-01.6
                           │
                           └────────► US-01.8
              
US-03.3 ──► US-03.4 ──► US-03.1
              │
              └────────► US-03.2
              
US-01.9   (independente — adiciona flag ao US-01.5)
US-02.5   (independente — adiciona provider ao US-02.4)
```

**Leitura:** uma seta `A ──► B` significa "A precisa estar pronto antes de B ser iniciado."

---

## Análise por história

### Sprint 1 — Fundação *(já pronta no repo de referência)*

---

#### US-01.1 — Estrutura base do CLI em Go
**Origem:** adaptar do repo de referência (`projetos/assinatura/`)  
**Esforço estimado:** 1–2 h  
**O que fazer:** criar o repositório pessoal, copiar a estrutura (`main.go`, `cmd/root.go`, `cmd/version.go`, `go.mod`), ajustar o nome do módulo Go para o seu repositório GitHub.  
**Atenção:** `go.mod` no repo de referência declara `go 1.26.1` mas a CI configura `go-version: '1.25'`. Alinhe os dois para `1.25`.

---

#### US-05.1 / US-05.2 / US-05.3 — Pipeline CI/CD + Releases + Cosign
**Origem:** adaptar do repo de referência (`.github/workflows/ci.yml`)  
**Esforço estimado:** 1–2 h  
**O que fazer:** copiar o workflow, ajustar os caminhos de `working-directory` para a sua estrutura de diretórios, garantir que `id-token: write` está nas permissões (necessário para Cosign keyless).  
**Atenção:** o workflow lê a versão com `grep` diretamente de `cmd/version.go`. Se você renomear o arquivo ou mover a variável, o pipeline quebra silenciosamente.

---

### Sprint 2 — Fluxo ponta-a-ponta local

**Ordem recomendada:** US-02.1 → US-02.2 → US-02.3 → US-01.2 → US-04.1 → US-01.3 → US-01.4

---

#### US-02.1 — FakeSignatureService
**Origem:** adaptar do repo de referência  
**Esforço estimado:** 1 h  
**O que fazer:** copiar `SignatureService`, `FakeSignatureService`, os três DTOs e o teste existente. Certificar que o projeto Maven compila e os testes passam antes de avançar. Este é o ponto de partida do assinador.jar — tudo em Sprint 2 e 3 é construído sobre ele.

---

#### US-02.2 — Validação de parâmetros de criação de assinatura
**Origem:** escrever do zero  
**Esforço estimado:** 3–5 h  
**O que fazer:** estender `FakeSignatureService.sign()` (ou extrair um validador separado) para checar todos os campos de `SignRequest` exigidos pela especificação FHIR. A referência atual só checa se `content` é nulo/vazio — isso não é suficiente.  
**Dependência:** requer leitura das especificações linkadas na seção 10 da `especificacao.md` para entender quais campos são obrigatórios e quais formatos são aceitos.  
**Risco:** a especificação FHIR é densa. Reserve tempo para leitura antes de escrever código. Comece listando os campos obrigatórios em comentário antes de implementar.

---

#### US-02.3 — Validação de parâmetros de validação de assinatura
**Origem:** escrever do zero  
**Esforço estimado:** 2–3 h  
**O que fazer:** mesma lógica de US-02.2, mas para `ValidateRequest`. `FakeSignatureService.validate()` já valida presença de `content` e `signature`; precisa adicionar validação de formato.  
**Dica:** faça US-02.2 primeiro — a estrutura de validação que você criar lá (uma classe `Validator`, por exemplo) pode ser reusada aqui quase sem modificações.

---

#### US-01.2 — Parsing de comandos `sign` e `validate` no CLI Go
**Origem:** escrever do zero (mas segue o padrão Cobra de `version.go`)  
**Esforço estimado:** 3–4 h  
**O que fazer:** criar `cmd/sign.go` e `cmd/validate.go` com Cobra. Definir as flags necessárias (`--content`, `--signature`, etc.). Por enquanto, os comandos só precisam parsear e imprimir o que receberam — a invocação do .jar vem em US-01.3.  
**Dica:** use `cmd/version.go` do repo de referência como template para criar os novos subcomandos.

---

#### US-04.1 — Detecção e provisionamento automático do JDK
**Origem:** escrever do zero  
**Esforço estimado:** 5–8 h *(história mais trabalhosa da Sprint 2)*  
**O que fazer:** lógica em `internal/jdk/` que (1) procura `java` no PATH e verifica se é versão 21+, (2) se ausente, baixa o JRE Eclipse Temurin para a plataforma correta via URL do Adoptium, (3) descompacta em `~/.hubsaude/jdk/`, (4) retorna o caminho para o executável `java`.  
**Risco:** o código precisa funcionar em Windows (`.zip`), Linux e macOS (`.tar.gz`) — os formatos de arquivo e os paths internos do JRE são diferentes em cada plataforma. Escreva testes que mockam o "java no PATH" para não precisar baixar 200 MB nos testes.  
**Dica:** separe "detectar java" de "baixar JRE" em funções distintas desde o início — facilita testar cada parte isoladamente.

---

#### US-01.3 — Invocação do assinador.jar no modo local
**Origem:** escrever do zero  
**Esforço estimado:** 4–6 h  
**Dependências:** US-01.2 + US-04.1 + US-02.1 (o .jar precisa existir para testar)  
**O que fazer:** em `internal/invoker/`, usar `os/exec` para construir e executar `java -jar assinador.jar <args>`, capturar stdout/stderr, e retornar o resultado estruturado para o CLI. O caminho do `java` vem de US-04.1.  
**Risco:** serialização dos parâmetros para linha de comando é frágil se houver espaços ou caracteres especiais nos valores. Prefira passar argumentos como slice (`exec.Command(java, "-jar", jar, arg1, arg2)`) — nunca concatenar uma string e passar para shell.

---

#### US-01.4 — Exibição legível de resultados
**Origem:** escrever do zero  
**Esforço estimado:** 1–2 h  
**Dependências:** US-01.3  
**O que fazer:** formatar a resposta do assinador.jar de forma legível no terminal. Sucesso imprime a assinatura; erro imprime a mensagem de forma destacada. Não precisa ser sofisticado — clareza é mais importante que estética.

---

### Sprint 3 — Modo Servidor HTTP e PKCS#11

**Ordem recomendada:** US-02.4 → US-01.5 → US-01.7 → US-01.6 → US-01.8 → US-01.9 → US-02.5 (por último, é o mais arriscado)

---

#### US-02.4 — Endpoints HTTP no assinador.jar
**Origem:** escrever do zero  
**Esforço estimado:** 4–6 h  
**O que fazer:** adicionar um servidor HTTP embutido ao assinador.jar com endpoints `POST /sign` e `POST /validate`. **Decisão pendente:** qual biblioteca usar? Opções sem Spring Boot: `com.sun.net.httpserver` (zero dependências, verboso), Javalin (simples, uma dependência), Undertow (pesado demais para este projeto). Recomendo **Javalin** ou o `httpserver` nativo para manter o `pom.xml` limpo.  
**Risco:** adicionar servidor HTTP significa adicionar dependência ao `pom.xml` — consulte o CLAUDE.md antes de escolher.

---

#### US-01.5 — Iniciar assinador.jar no modo servidor
**Origem:** escrever do zero  
**Esforço estimado:** 3–4 h  
**O que fazer:** comando `assinatura start [--port N]` que inicia o .jar como processo em background e salva PID + porta em `~/.hubsaude/assinador.pid`. Em Windows o conceito de "background process" requer cuidado com `SysProcAttr`.

---

#### US-01.7 — Detectar instância em execução
**Origem:** escrever do zero  
**Esforço estimado:** 2–3 h  
**O que fazer:** ler o arquivo `.pid`, verificar se o processo existe e fazer um GET em `/health` (ou equivalente) para confirmar que está respondendo. Processo registrado mas não respondente = inativo.

---

#### US-01.6 — Invocar assinador.jar via HTTP
**Origem:** escrever do zero  
**Esforço estimado:** 3–4 h  
**Dependências:** US-02.4 + US-01.5 + US-01.7  
**O que fazer:** no `internal/invoker/`, adicionar um modo HTTP que faz `POST /sign` ou `POST /validate` ao servidor em execução. A lógica de escolha (HTTP vs local) usa US-01.7 para decidir.

---

#### US-01.8 — Interromper o assinador.jar
**Origem:** escrever do zero  
**Esforço estimado:** 1–2 h  
**O que fazer:** `assinatura stop [--port N]` — envia sinal de encerramento (via HTTP `/shutdown` ou kill do PID) e limpa o arquivo `.pid`.

---

#### US-01.9 — Timeout por inatividade
**Origem:** escrever do zero  
**Esforço estimado:** 2–3 h  
**O que fazer:** parâmetro `--timeout <minutos>` no `assinatura start`. Pode ser implementado no lado do servidor Java (contando requisições) ou no CLI (timer que chama stop). O lado Java é mais correto; o lado Go é mais fácil.

---

#### US-02.5 — Integração PKCS#11
**Origem:** escrever do zero  
**Esforço estimado:** 8–12 h *(história de maior risco do projeto)*  
**O que fazer:** configurar `SunPKCS11` provider no Java, integrar com SoftHSM2 nos testes. Esta é a única história com risco real de não entregar — requer instalação de software externo (SoftHSM2) e entendimento de criptografia de hardware.  
**Estratégia recomendada:** implemente como funcionalidade opcional. Se PKCS#11 não estiver disponível, o serviço continua funcionando com `FakeSignatureService`. Nunca bloqueie o fluxo principal por conta desta integração.  
**Deixe para o final da Sprint 3.** Se o tempo apertar, documente o que falta em vez de entregar código quebrado.

---

### Sprint 4 — Simulador HubSaúde

**Ordem recomendada:** US-03.3 → US-03.4 → US-03.1 → US-03.2

---

#### US-03.3 — Estrutura base do CLI `simulador`
**Origem:** adaptar do CLI `assinatura`  
**Esforço estimado:** 2–3 h  
**O que fazer:** criar `cmd/simulador/` (ou `projetos/simulador/`) copiando a estrutura do `assinatura`. Adicionar stubs dos comandos `start`, `stop`, `status`. Atualizar o CI para compilar e publicar o segundo binário.

---

#### US-03.4 — Download dinâmico do simulador.jar
**Origem:** escrever do zero  
**Esforço estimado:** 4–6 h  
**O que fazer:** em `internal/release/`, buscar `release.json` na URL fixa do repo da disciplina, comparar versão com o que há em `~/.hubsaude/`, baixar apenas se necessário. Verificar checksum SHA256 do arquivo baixado.  
**Dica:** reuse a lógica de download de US-04.1 (baixar arquivo de URL + verificar hash). A diferença é que aqui você também compara versões.

---

#### US-03.1 — Iniciar o Simulador
**Origem:** escrever do zero  
**Esforço estimado:** 3–4 h  
**Dependências:** US-03.4 + US-04.1  
**O que fazer:** `simulador start` — verificar se porta 8443 está livre, iniciar `java -jar simulador.jar` em background, registrar PID.

---

#### US-03.2 — Parar e monitorar o Simulador
**Origem:** escrever do zero  
**Esforço estimado:** 2–3 h  
**O que fazer:** `simulador stop` → chama `GET /shutdown`; `simulador status` → chama `GET /api/info` e exibe o resultado.

---

## Resumo: adaptar vs. escrever do zero

| História | Estratégia | Justificativa |
|----------|------------|---------------|
| US-01.1 | **Adaptar** | Estrutura e código idênticos; só muda nome do módulo |
| US-05.1/5.2/5.3 | **Adaptar** | CI é configuração, não lógica; ajustar caminhos e permissões |
| US-02.1 | **Adaptar** | `FakeSignatureService` e DTOs completos no repo de referência |
| US-02.2 | **Do zero** | Validação completa não existe no repo de referência |
| US-02.3 | **Do zero** | Idem |
| US-01.2 | **Do zero** | Comandos `sign`/`validate` não existem; `version.go` serve de template |
| US-01.3 | **Do zero** | Nenhum código de invocação no repo de referência |
| US-01.4 | **Do zero** | Nenhum código de formatação no repo de referência |
| US-04.1 | **Do zero** | Nenhum código de JDK no repo de referência |
| US-02.4 | **Do zero** | Servidor HTTP não existe no .jar de referência |
| US-02.5 | **Do zero** | PKCS#11 não existe no repo de referência |
| US-01.5–1.9 | **Do zero** | Nenhum código de servidor/lifecycle no repo de referência |
| US-03.3 | **Adaptar** | Estrutura idêntica ao CLI `assinatura` |
| US-03.1/3.2/3.4 | **Do zero** | Sem código de simulador no repo de referência |

---

## Estimativa total

| Sprint | Histórias restantes | Esforço estimado |
|--------|--------------------|--------------------|
| 1 | Configurar próprio repo + adaptar código | 2–4 h |
| 2 | US-02.2, 02.3, 01.2, 01.3, 01.4, 04.1 | 18–28 h |
| 3 | US-02.4, 02.5, 01.5–01.9 | 23–34 h |
| 4 | US-03.1–03.4 | 11–16 h |
| **Total** | | **54–82 h de trabalho focado** |

US-02.5 (PKCS#11) responde por ~15% do esforço total e tem o maior risco de atraso. Se o cronograma apertar, é a primeira candidata a virar entrega parcial com documentação do que falta.
