# ADR 0001 — Uso de Go 1.25 para os CLIs e Java 21 para o assinador.jar

- **Data:** 2026-05-19
- **Status:** Aceito

## Contexto

O Sistema Runner é composto por três executáveis: dois CLIs (`assinatura` e `simulador`)
que o usuário interage diretamente, e um componente de processamento (`assinador.jar`)
que implementa a lógica de assinatura digital. O trabalho prático exige distribuição
multiplataforma (Windows, Linux, macOS) sem que o usuário configure ambiente de execução
manualmente.

A escolha de linguagem afeta diretamente: tamanho e portabilidade dos binários,
complexidade do pipeline de CI/CD, experiência de instalação para o usuário final, e
esforço de manutenção.

## Decisão

**CLIs (`assinatura` e `simulador`):** Go 1.25 com Cobra para parsing de comandos.

**Componente de processamento (`assinador.jar`):** Java 21 com Maven.

## Justificativa

### Go para os CLIs

Go compila para binários estáticos nativos sem dependência de runtime. Isso significa
que o usuário baixa um único arquivo executável e roda — sem instalar Go, sem configurar
PATH, sem gerenciar versões. Cross-compilation para as três plataformas-alvo acontece em
um único runner Linux no CI, com `GOOS`/`GOARCH` como variáveis de ambiente.

A biblioteca padrão cobre todas as necessidades dos CLIs sem dependências externas:
invocação de processos (`os/exec`), HTTP client (`net/http`), manipulação de arquivos e
detecção de plataforma. Cobra, a única dependência externa, é a biblioteca de facto para
CLIs em Go e resolve parsing, geração de `--help` e composição de subcomandos.

### Java para o assinador.jar

Java 21 é uma **restrição do projeto**, não uma escolha livre. O trabalho prático exige
Java por dois motivos concretos: compatibilidade futura com a plataforma HubSaúde, que
já opera com assinadores em Java, e a necessidade de integração com PKCS#11 via
`SunPKCS11` provider — uma API disponível na JVM padrão sem dependências adicionais.

### A combinação cria uma tensão gerenciada

Go resolve a experiência do usuário final (instalação zero), mas introduz a necessidade
de gerenciar o ciclo de vida do Java para o `assinador.jar`. Isso motivou diretamente
US-04.1 (provisionar JDK automaticamente) e a lógica de `~/.hubsaude/` para armazenar
o JRE baixado. Em outras palavras: Go elimina a dependência de runtime para o CLI, mas
o CLI passa a ser responsável por provisionar o runtime do componente Java.

## Alternativas consideradas

**Rust para os CLIs:** binários igualmente estáticos e performáticos, mas curva de
aprendizado significativamente maior e ecossistema de bibliotecas CLI menos maduro que
Go para este tipo de projeto.

**Go para tudo (incluindo o assinador):** eliminaria a tensão descrita acima, mas viola
a restrição do trabalho prático e inviabiliza a integração PKCS#11 com tokens reais no
futuro.

**Python para os CLIs:** distribuição multiplataforma sem compilação exige PyInstaller ou
similar, resultando em binários grandes (~50 MB) e processo de build mais complexo.

## Consequências

**Positivas:**
- Binários dos CLIs são distribuídos sem dependências; o usuário não precisa de Go instalado.
- Cross-compilation funciona de um único runner no CI, sem VMs separadas por plataforma.
- Java 21 LTS garante suporte de longo prazo e disponibilidade de JRE via Eclipse Temurin.

**Negativas:**
- O CLI precisa localizar ou baixar um JRE para invocar o `assinador.jar` — lógica que
  não existiria se o componente de processamento fosse também em Go.
- Dois ecossistemas de build (Go modules + Maven) no mesmo repositório aumentam a
  superfície de configuração do CI.
- O "cold start" da JVM (~200–500 ms) é perceptível quando o assinador.jar é invocado no
  modo local. Essa latência motivou o modo servidor (ver ADR 0002).
