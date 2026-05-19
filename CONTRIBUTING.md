# Contributing

## Branches

```
feature/US-XX.Y-descricao-curta   → nova funcionalidade
fix/US-XX.Y-descricao-curta       → correção de bug
docs/descricao-curta              → documentação apenas
refactor/descricao-curta          → refatoração sem mudança de comportamento
```

Exemplos: `feature/US-01.2-comandos-sign-validate`, `fix/US-04.1-deteccao-jdk-windows`

## Commits

Seguimos [Conventional Commits](https://www.conventionalcommits.org/):

```
<tipo>(<escopo>): <descrição curta no imperativo>
```

| Tipo | Quando usar |
|------|-------------|
| `feat` | nova funcionalidade |
| `fix` | correção de bug |
| `test` | adicionar ou corrigir testes |
| `refactor` | mudança de código sem alterar comportamento |
| `docs` | documentação |
| `ci` | mudanças no pipeline |
| `chore` | tarefas de manutenção (dependências, config) |

Exemplos:
```
feat(cli): adicionar comando sign com flags --content e --token
fix(jdk): corrigir detecção de java no PATH no Windows
test(assinador): cobrir validação de content nulo em FakeSignatureService
```

**Regras:**
- Descrição em português, minúsculas, sem ponto final
- Escopo é o componente afetado: `cli`, `assinador`, `simulador`, `jdk`, `ci`
- Limite de 72 caracteres na primeira linha

## Antes de abrir PR

### Go
```bash
go vet ./...
go test ./...
```

### Java (assinador-java)
```bash
cd projetos/assinador-java
mvn test
```

Nenhum PR deve ser aberto com `go vet` ou `mvn test` falhando.

## Pull Request

Use o template em `.github/PULL_REQUEST_TEMPLATE.md`. Preencha todos os campos — PRs sem checklist preenchida não serão revisados.
