# ADR-0005 — Interação com dispositivo criptográfico via PKCS#11 (`SunPKCS11`)

**Status:** Aceito
**Data:** 2026-06-29
**Relacionado:** [ADR-0001](0001-uso-de-go-e-java.md) (Go + Java), [ADR-0003](0003-contrato-cli-jar.md) (contrato JSON), [ADR-0004](0004-servidor-http-jdk.md) (servidor HTTP std-lib)

## Contexto

A US-02 exige que o `assinador.jar` "suporte interação com dispositivo
criptográfico (token/smart card) via interface PKCS#11". O critério de aceitação
E5 (`docs/criterios.md` da referência) pede explicitamente "testes de integração
que comprovam que [o assinador] responde a **chamadas PKCS#11 reais**".

Há uma tensão aparente com o escopo: a especificação lista "implementação real de
assinatura digital criptográfica" como **fora do escopo**. O `design.md` da
referência reconcilia os dois ao descrever o assinador como um processo que
"valida parâmetros e **simula assinaturas interagindo com o dispositivo
criptográfico**".

## Decisão

Implementar a **interação real** com o dispositivo via provider **`SunPKCS11`**,
embutido no JDK (módulo `jdk.crypto.cryptoki`), mantendo a **saída simulada**.

- **Sem dependência nova:** `SunPKCS11` faz parte do JDK; o `pom.xml` continua só
  com JUnit (escopo de teste). Respeita a política std-lib-first do CLAUDE.md e o
  critério F (supply chain) — coerente com o ADR-0004.
- **Interação real (satisfaz E5):** `Provider.configure(...)` configura o provider
  a partir do caminho da biblioteca e do slot; abre-se `KeyStore("PKCS11")` e
  executa-se `load(null, pin)` — um `C_Login` real no token usando o PIN. O
  primeiro alias/certificado é lido apenas para evidenciar a comunicação.
- **Saída simulada (respeita o escopo):** o valor da assinatura continua sendo o
  do `FakeSignatureService`. Quem decide o conteúdo é o serviço base; a camada
  PKCS#11 responde só pela *interação* com o hardware.
- **Estritamente opcional, nunca bloqueia o fluxo padrão:** o caminho PKCS#11 só é
  ativado quando o usuário passa `--pkcs11-lib <caminho>` (e, opcionalmente,
  `--pkcs11-slot <n>`), no modo CLI ou servidor. Sem essa flag, usa-se o
  `FakeSignatureService` como antes. O PIN vem do parâmetro existente `--token`.

### Desenho

```
SignatureService (interface)
 ├── FakeSignatureService            (validação + assinatura simulada)
 └── Pkcs11SignatureService          (decorator: simula + interage com o token)
        └── Pkcs11Gateway (interface)
              └── SunPkcs11Gateway   (SunPKCS11 real: configure + KeyStore + login)
```

O `Pkcs11Gateway` isola a chamada de baixo nível, permitindo testar a lógica de
decisão do serviço com um dublê (sem token/SoftHSM2) e exercitar a implementação
real em um teste de integração condicional.

### Distinção de erros (critério E3)

| Situação | Resposta | `exit` (CLI) |
|----------|----------|--------------|
| Parâmetro inválido (erro do usuário) | validação do serviço base; dispositivo **não** é acionado | 1 |
| Dispositivo configurado, mas indisponível/recusado (erro de sistema) | `valid:false`, mensagem `dispositivo PKCS#11: ...` | 1 |
| Interação bem-sucedida | assinatura simulada + evidência do token na mensagem | 0 |

## Justificativa

- **`SunPKCS11` é a interface PKCS#11 padrão do Java** — exatamente o que a US-02 e
  o `design.md` pedem — sem trazer bibliotecas de terceiros.
- **Decorator** preserva o `FakeSignatureService` intacto e reutiliza sua
  validação/simulação, mantendo uma única fonte de verdade (critério E3).
- **Opt-in explícito** garante que o fluxo padrão (a esmagadora maioria dos usos,
  já que o projeto simula) nunca dependa de hardware — alinhado ao "nunca bloqueie
  o fluxo principal" do plano.

## Alternativas consideradas

- **Biblioteca PKCS#11 de terceiros (ex.: IAIK, BouncyCastle PKCS#11):** dependência
  externa para algo que o JDK já provê. Descartada (std-lib-first).
- **Assinatura criptográfica real com a chave do token** (`Signature` sobre o
  provider): cruzaria o limite de escopo "sem criptografia real" e exigiria chave
  provisionada no dispositivo. Descartada — o E5 pede *chamadas PKCS#11 reais*
  (login/acesso), não assinatura real.
- **Instalar SoftHSM2 no CI** para rodar o teste real sempre: aumenta o tempo e a
  fragilidade do pipeline (instalação por SO). Optou-se por teste **gated**
  (`assumeTrue`), que roda quando o dispositivo está provisionado; a instalação no
  CI fica como evolução futura, se desejada.

## Consequências

**Positivas**
- Nenhuma dependência nova; build e distribuição inalterados.
- Interação PKCS#11 real e testável; lógica de decisão coberta sem hardware.
- Fluxo padrão inalterado; PKCS#11 é aditivo e opcional.

**Negativas**
- O teste que comprova chamadas reais (E5) só roda em ambiente com token/SoftHSM2
  provisionado (condicional). No CI atual ele é pulado — cobre-se o caminho de
  erro do gateway real (biblioteca inexistente) de forma incondicional.
- Configurar SoftHSM2 localmente para o teste de integração exige passos manuais
  (documentados no `Pkcs11IntegrationTest`).

## Testes que cobrem esta decisão

- `Pkcs11ConfigTest`: montagem da config inline do `SunPKCS11` (slot explícito vs.
  `slotListIndex`), nome padrão, validação de biblioteca obrigatória.
- `Pkcs11SignatureServiceTest`: interação com dublê — assinatura válida mantém a
  saída simulada e propaga o PIN; parâmetro inválido não aciona o dispositivo;
  falha do dispositivo vira erro de sistema; validação delega sem tocar no token.
- `SunPkcs11GatewayTest`: gateway real com biblioteca inexistente falha com
  `Pkcs11Exception` legível (incondicional, roda no CI).
- `Pkcs11IntegrationTest`: **gated** (`assumeTrue` em `PKCS11_LIBRARY`/`PKCS11_PIN`)
  — comprova `C_Login` real contra SoftHSM2/token (critério E5).
