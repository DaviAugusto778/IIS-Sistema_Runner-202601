## História

<!-- Ex.: US-01.2 — Parsing de comandos sign e validate no CLI Go -->

## O que foi feito

<!-- Descrição objetiva das mudanças. O que o código faz agora que não fazia antes. -->

## Como testar

<!-- Passos para reproduzir o comportamento manualmente, se aplicável. -->

```bash
# Ex.:
go run . sign --content "hello"
```

## Checklist

### Geral
- [ ] Branch segue o padrão `feature/US-XX.Y-descricao`
- [ ] Commits seguem Conventional Commits
- [ ] Nenhum arquivo de credencial, `.env` ou binário commitado

### Go (se aplicável)
- [ ] `go vet ./...` passa sem warnings
- [ ] `go test ./...` passa
- [ ] Novo código tem teste cobrindo caminho feliz e pelo menos um caso de erro
- [ ] Comandos novos têm `--help` preenchido

### Java (se aplicável)
- [ ] `mvn test` passa
- [ ] Validações novas têm testes para: nulo, vazio, formato inválido, e caso válido
- [ ] Nenhuma dependência adicionada ao `pom.xml` sem justificativa nos comentários do PR

## Histórias desbloqueadas

<!-- Quais histórias do meu-plano.md ficam disponíveis para iniciar após este merge? -->
<!-- Deixe em branco se nenhuma. -->
