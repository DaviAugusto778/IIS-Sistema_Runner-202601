# Sistema Runner 🏃‍♂️


O **Sistema Runner** é um conjunto de ferramentas de linha de comando (CLI) desenvolvido para simplificar e abstrair a execução de aplicações Java associadas à Plataforma HubSaúde. Ele elimina a necessidade de configurações manuais de ambiente (como instalação do JDK) por parte dos usuários e integradores.

## 🎯 Escopo da Primeira Iteração

Nesta primeira fase de construção, nosso foco está em estabelecer a infraestrutura base e implementar a comunicação essencial:

- [ ] **Setup do repositório**: Estruturação dos diretórios para o código CLI (Go/Rust/Python/etc.) e o código Java (`assinador.jar`).
- [ ] **CLI Base (`assinatura`)**: Criação da estrutura de comandos no terminal (parsing de argumentos).
- [ ] **Módulo `assinador.jar` (Esqueleto)**: Implementação da classe principal em Java capaz de receber parâmetros via linha de comando e responder com um log de sucesso/erro.
- [ ] **US-01 (Parcial)**: Invocação local direta do `assinador.jar` a partir do CLI.

## 🏗️ Arquitetura (Visão Geral)

O projeto baseia-se no modelo C4 e é composto por três componentes principais na visão de contêineres:

1. **CLI de Assinatura (`assinatura`)**: Interface nativa para o usuário.
2. **CLI do Simulador (`simulador`)**: Gerenciador do ciclo de vida da aplicação HubSaúde.
3. **Módulo Core (`assinador.jar`)**: Motor de validação e simulação escrito em Java.

### Pré-requisitos
* [Ferramenta de compilação da linguagem escolhida para o CLI]
* Java Development Kit (JDK) 17 ou superior
* Maven ou Gradle (para o `assinador.jar`)