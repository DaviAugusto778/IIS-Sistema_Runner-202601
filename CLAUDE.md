# Sistema Runner - Contexto para o Claude Code

## Sobre o projeto
Trabalho prático da disciplina de Implementação e Integração (UFG, 2026-1).
Sistema CLI multiplataforma para invocar aplicações Java (assinador.jar e
simulador HubSaúde) sem que o usuário precise configurar JDK manualmente.

## Stack
- Go 1.25 para os dois CLIs (assinatura, simulador)
- Java 21 + Maven para o assinador.jar
- Cobra para o parsing de comandos CLI
- GitHub Actions para CI/CD multiplataforma
- Cosign + SHA256 para integridade dos releases

## Referência
A especificação oficial e o plano vêm do repositório do professor em:
../runner-referencia/ (Diretorio: C:\Users\USER\Desktop\runner-referencia)

Sempre consulte:
- ../runner-referencia/especificacao.md (épicos US-01 a US-05)
- ../runner-referencia/docs/plano-revisitado-v2.md (sprints e histórias)
- ../runner-referencia/design.md (modelo C4)

## Convenções
- Branches: feature/US-XX.Y-descricao-curta
- Commits: Conventional Commits (feat:, fix:, docs:, test:, refactor:)
- Toda história precisa de testes unitários antes do merge
- Não usar lib externa quando algo da std lib resolve

## Coisas que o Claude NÃO deve fazer
- Não fazer commit automático sem eu pedir
- Não criar dependências novas sem me consultar
- Não modificar arquivos em ../runner-referencia/ (read-only)