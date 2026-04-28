# Plano de Gerenciamento do Projeto: Sistema Runner

## 1. Visão Geral e Escopo
O **Sistema Runner** tem como objetivo facilitar a execução de aplicações Java vinculadas à Plataforma HubSaúde através de uma Interface de Linha de Comandos (CLI) simples, eliminando a necessidade de configuração manual do ambiente (como a instalação do JDK) por parte dos usuários. 

O escopo compreende a entrega de três componentes principais executáveis nas plataformas Windows, Linux e macOS:
* A aplicação CLI `assinatura` em Go.
* A aplicação CLI `simulador` em Go.
* A aplicação `assinador.jar` em Java 21.

## 2. Abordagem Metodológica
Para garantir entregas contínuas de valor e adaptação rápida a riscos técnicos, o projeto adota um ciclo de vida **iterativo e incremental**. 
* **Cadência:** O desenvolvimento está estruturado em **4 Sprints**.
* **Duração da Sprint:** 1 semana cada, focando na entrega de valor ou remoção de riscos específicos.
* **Integração:** Toda submissão de código acionará pipelines de CI/CD automatizados.

---

## 3. Cronograma Macro (Roadmap)

A alocação das histórias de usuário (US) foi distribuída para construir a fundação nas primeiras semanas e integrar as lógicas de negócio e segurança nas finais:

| Sprint | Duração | Foco Principal | Marcos e Entregáveis |
| :--- | :--- | :--- | :--- |
| **Sprint 1** | 1 Semana | Fundação e CI/CD | Estrutura base de pacotes em Go; Pipeline GitHub Actions configurado; Binários CLI "Hello World" publicados via Releases multiplataforma. |
| **Sprint 2** | 1 Semana | Assinatura Simulada (Modo Local) | Implementação do `FakeSignatureService` em Java; CLI invocando o arquivo `.jar` via comando local; Download e provisionamento automático do JDK. |
| **Sprint 3** | 1 Semana | Modo Servidor e Material Criptográfico | `assinador.jar` expondo endpoints HTTP (`/sign`, `/validate`); CLI gerenciando o processo em background; Integração com interface PKCS#11. |
| **Sprint 4** | 1 Semana | Orquestração do Simulador e Segurança | CLI `simulador` desenvolvido para controle de ciclo de vida; Download dinâmico do `simulador.jar`; Artefatos selados com Cosign (Sigstore) e checksums. |

---

## 4. Definição de Pronto (Definition of Done - DoD)
Para que uma história de usuário ou funcionalidade seja considerada oficialmente concluída, ela deve satisfazer rigorosamente os seguintes critérios:
* O código fonte foi implementado de forma limpa e adicionado ao repositório.
* O código atende a todos os Critérios de Aceitação descritos na especificação da história correspondente.
* Testes unitários e de integração foram desenvolvidos com boa cobertura de código, focando especialmente em cenários de erro e validações restritas.
* A compilação e os testes passam com sucesso na esteira do CI (GitHub Actions).
* Os artefatos binários foram gerados para as arquiteturas `amd64` nos sistemas Windows, Linux e macOS.
* Documentação e exemplos de uso em linha de comando foram atualizados no repositório.

---

## 5. Gerenciamento de Riscos Técnicos

Ao longo do desenvolvimento da arquitetura e das integrações de software, alguns riscos foram mapeados, bem como as estratégias para mitigá-los de forma proativa:

* **Integração Criptográfica (PKCS#11):** * *Risco:* Interagir com hardware físico (tokens) pode gerar impeditivos locais de desenvolvimento. 
  * *Mitigação:* Usar o simulador em software **SoftHSM2** nas fases iniciais para testes de integração com o `SunPKCS11` (Java).
* **Gestão de Processos Órfãos (Go x Java):**
  * *Risco:* O CLI não conseguir encerrar o `assinador.jar` ou o `simulador.jar` em background, ocupando indevidamente portas do sistema operacional.
  * *Mitigação:* Desenvolver o módulo de estado no Go que grave o PID e a porta no diretório `~/.hubsaude/` assim que a aplicação iniciar. Implementar flags rígidas de `--timeout` por inatividade.
* **Provisionamento do JDK:**
  * *Risco:* Falha nos scripts de download automático do JDK corrompendo a execução.
  * *Mitigação:* Verificar integridade (checksum) logo após o download; não repetir a tentativa se o diretório configurado já possuir os binários válidos na versão 21.