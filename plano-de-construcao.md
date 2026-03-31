# Plano de Construção: Sistema Runner (Tarefas Operacionais)

## Premissas Arquiteturais e Técnicas
* **CLI**: Desenvolvimento em Go 1.25, utilizando pacotes robustos para parsing de comandos (como `cobra`) e gerenciamento de rede/processos.
* **Core**: Desenvolvimento em Java 21 para o `assinador.jar`.
* **Estado Local**: Utilização do diretório `~/.hubsaude/` como repositório de dados principal para rastrear PIDs, portas em uso e instâncias em execução, garantindo a integridade do ciclo de vida.

---

## Sprint 1: Fundação, Estrutura e Pipeline CI/CD
**Foco:** Criar o esqueleto arquitetural dos projetos e automatizar a entrega.

* **Tarefa 1.1 (Go): Inicializar Workspace.** Configurar o módulo principal (`go mod init`) e definir a arquitetura de pastas (ex: `cmd/` para o CLI, `internal/` para lógica de negócios).
* **Tarefa 1.2 (Go): Implementar Parsing Básico.** Criar o comando raiz e o subcomando `version` para validar o build multiplataforma.
* **Tarefa 1.3 (DevOps): Configurar GitHub Actions (Build).** Criar workflow para realizar cross-compilation automática para `windows/amd64`, `linux/amd64` e `darwin/amd64` a cada push na branch principal.
* **Tarefa 1.4 (DevOps): Configurar GitHub Actions (Release).** Estender o workflow para publicar binários no GitHub Releases automaticamente ao criar uma tag baseada em SemVer (ex: `v0.1.0`), seguindo o padrão de nomenclatura `assinatura-<versão>-<os>-<arch>`.

---

## Sprint 2: Motor de Simulação e Invocação Local
**Foco:** Desenvolver a validação rigorosa no Java e o gerenciamento de processos no Go.

* **Tarefa 2.1 (Java): Estruturar `SignatureService`.** Criar o projeto Java 21, definir a interface principal com os métodos `sign` e `validate`, e implementar a classe `FakeSignatureService` que retornará dados estáticos pré-construídos para cenários de sucesso.
* **Tarefa 2.2 (Java): Implementar Motor de Validação.** Codificar a lógica de sanitização e validação estrita de todos os parâmetros de entrada. O sistema deve rejeitar requisições malformadas e retornar mensagens detalhando o erro e o parâmetro problemático.
* **Tarefa 2.3 (Go): Desenvolver Subcomandos CLI.** Implementar o parsing seguro para os comandos `sign` e `validate`, garantindo que parâmetros ausentes acionem mensagens de ajuda (`--help`).
* **Tarefa 2.4 (Go): Orquestrar Processo Local.** Desenvolver a função em Go responsável por localizar o executável Java, montar a string de execução (`java -jar assinador.jar`) com os argumentos corretos, invocar o processo e capturar o `stdout`/`stderr`.
* **Tarefa 2.5 (Go): Lógica de Provisionamento JDK.** Criar um módulo para verificar a existência do JDK no `PATH` ou em `~/.hubsaude/`. Se ausente, implementar rotina para download seguro e extração do JDK compatível com o SO atual.

---

## Sprint 3: Modo Servidor, Redes e Criptografia
**Foco:** Implementar comunicação cliente-servidor e integração com PKCS#11.

* **Tarefa 3.1 (Java): Construir Camada Web.** Implementar um servidor embutido com o `SignatureController` expondo os endpoints `POST /sign` e `POST /validate`, reaproveitando a lógica de validação criada na Sprint 2.
* **Tarefa 3.2 (Java): Configurar Provider Criptográfico.** Integrar a classe `SunPKCS11` para possibilitar a comunicação com dispositivos físicos (tokens/smart cards), criando testes de integração com o SoftHSM2.
* **Tarefa 3.3 (Go): Gerenciamento de Estado (Startup).** Implementar rotina para iniciar o Java em background. O Go deve capturar o PID e a porta alocada, e persistir esses dados em um arquivo de estado local dentro de `~/.hubsaude/` para consultas futuras.
* **Tarefa 3.4 (Go): Cliente HTTP.** Construir o módulo de rede no Go para enviar requisições HTTP para a porta registrada no estado local, reduzindo a latência nas operações (*warm start*).
* **Tarefa 3.5 (Go): Comandos de Ciclo de Vida.** Implementar o comando `stop` para ler o PID do estado local e encerrar o processo educadamente, e a flag `--timeout` para encerrar o servidor após inatividade.

---

## Sprint 4: Orquestração do Simulador e Segurança da Cadeia de Suprimentos
**Foco:** Entregar o CLI do Simulador e selar os artefatos com criptografia.

* **Tarefa 4.1 (Go): Estruturar CLI `simulador`.** Criar o projeto independente para o segundo CLI, contemplando os comandos `start`, `stop` e `status`.
* **Tarefa 4.2 (Go): Rotina de Download Dinâmico.** Implementar cliente HTTP para consultar a API do GitHub Releases, baixar o artefato mais recente do `simulador.jar` (caso não exista no cache local) e validar sua integridade via checksum.
* **Tarefa 4.3 (Go): Validação de Portas.** Antes do `start`, implementar verificação de rede (socket) para garantir que as portas exigidas pelo HubSaúde estão livres.
* **Tarefa 4.4 (DevOps): Segurança com Cosign.** Atualizar o pipeline de CI/CD para incluir a etapa de assinatura de artefatos com o Cosign (identidade OIDC e log de transparência), garantindo a geração dos arquivos `.sig` e `.pem` exigidos na especificação. Incluir também a geração e publicação de arquivos de checksum SHA256.