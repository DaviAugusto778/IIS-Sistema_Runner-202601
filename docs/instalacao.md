# Guia de Instalação — Sistema Runner

Este guia cobre a instalação dos CLIs **`assinatura`** e **`simulador`** nas três
plataformas suportadas (Windows, Linux e macOS, amd64), a verificação de
integridade dos binários e o preparo do ambiente.

> Você **não** precisa instalar Java manualmente. Os CLIs detectam um JDK 21 já
> presente ou baixam um JRE Eclipse Temurin automaticamente para `~/.hubsaude/`
> na primeira execução que precisar dele (US-04). Veja [Requisitos](#requisitos).

---

## Requisitos

| Item | Necessário para usar (binário pronto) | Necessário para compilar |
|------|----------------------------------------|--------------------------|
| Binário do CLI (Releases) | ✅ | — |
| Java 21 | ❌ (provisionado automaticamente) | ✅ (JDK 21) |
| Go 1.25+ | ❌ | ✅ |
| Maven (`mvnw` incluso) | ❌ | ✅ (só para o `assinador.jar`) |
| Acesso à internet | Na 1ª execução (download do JRE/JARs) | Para baixar dependências |
| [Cosign](https://docs.sigstore.dev/cosign/installation/) | Opcional (verificar assinatura) | — |

O CLI cria e usa o diretório **`~/.hubsaude/`** para guardar o JRE provisionado,
os JARs baixados, os logs e o estado dos processos em segundo plano.

---

## 1. Baixar o binário

Baixe o artefato da sua plataforma na página de
[**Releases**](https://github.com/DaviAugusto778/IIS-Sistema_Runner-202601/releases)
do projeto. Os nomes seguem o padrão `<app>-<versão>-<os>-<arch>`:

| Plataforma | `assinatura` | `simulador` |
|------------|--------------|-------------|
| Windows x64 | `assinatura-v1.0.0-windows-amd64.exe` | `simulador-v1.0.0-windows-amd64.exe` |
| Linux x64 | `assinatura-v1.0.0-linux-amd64` | `simulador-v1.0.0-linux-amd64` |
| macOS x64 | `assinatura-v1.0.0-darwin-amd64` | `simulador-v1.0.0-darwin-amd64` |

Cada binário acompanha três arquivos auxiliares: `.sha256` (checksum),
`.sig` e `.pem` (assinatura Cosign). Para usar o `assinatura` com o assinador,
baixe também **`assinador.jar`** da mesma release.

---

## 2. Verificar a integridade (recomendado)

### Checksum SHA256

```bash
# Linux/macOS
sha256sum -c assinatura-v1.0.0-linux-amd64.sha256

# Windows (PowerShell) — compare manualmente com o conteúdo do .sha256
Get-FileHash .\assinatura-v1.0.0-windows-amd64.exe -Algorithm SHA256
```

### Assinatura Cosign (keyless)

Os artefatos são assinados pelo pipeline via OIDC do GitHub Actions, então a
verificação informa a identidade do assinante e o emissor do token:

```bash
cosign verify-blob \
  --certificate           assinatura-v1.0.0-linux-amd64.pem \
  --signature             assinatura-v1.0.0-linux-amd64.sig \
  --certificate-identity-regexp "^https://github.com/DaviAugusto778/IIS-Sistema_Runner-202601/.github/workflows/release.yml@refs/tags/v.*$" \
  --certificate-oidc-issuer     "https://token.actions.githubusercontent.com" \
  assinatura-v1.0.0-linux-amd64
```

A saída `Verified OK` confirma a autenticidade e a integridade do arquivo.

---

## 3. Instalar

### Linux / macOS

```bash
# Renomeie para um nome curto e torne executável
mv assinatura-v1.0.0-linux-amd64 assinatura
chmod +x assinatura

# O assinador.jar deve ficar AO LADO do binário (nome exato: assinador.jar)
mv assinador.jar ./assinador.jar

# (opcional) coloque no PATH
sudo mv assinatura /usr/local/bin/        # se mover o binário, veja a nota do JAR abaixo
```

> **macOS:** na primeira execução o Gatekeeper pode bloquear um binário não
> notarizado. Libere com `xattr -d com.apple.quarantine ./assinatura` ou em
> *Ajustes do Sistema → Privacidade e Segurança → Abrir mesmo assim*.

### Windows (PowerShell)

```powershell
# Renomeie para um nome curto
Rename-Item .\assinatura-v1.0.0-windows-amd64.exe assinatura.exe

# O assinador.jar deve ficar na mesma pasta do .exe (nome exato: assinador.jar)
# (o arquivo já deve se chamar assinador.jar ao baixar)
```

> **Windows:** o SmartScreen pode avisar sobre um executável desconhecido.
> Clique em *Mais informações → Executar assim mesmo* (a verificação Cosign do
> passo 2 já comprova a procedência).

### Onde o `assinatura` procura o `assinador.jar`

1. Variável de ambiente **`ASSINADOR_JAR`** (caminho absoluto), se definida; senão
2. um arquivo **`assinador.jar`** no **mesmo diretório do executável**.

Se você mover o binário para o PATH, aponte o JAR explicitamente:

```bash
export ASSINADOR_JAR="$HOME/.hubsaude/assinador.jar"   # Linux/macOS
$env:ASSINADOR_JAR = "C:\Users\voce\.hubsaude\assinador.jar"   # Windows
```

> O **`simulador`** **não** exige esse passo: ele baixa o `simulador.jar`
> automaticamente para `~/.hubsaude/` a partir do `release.json` da disciplina.

---

## 4. Verificar a instalação

```bash
./assinatura version
./assinatura sign --content "teste de instalacao"
./simulador version
```

A primeira chamada que precisar de Java pode demorar um pouco (download do JRE).
Chamadas seguintes reutilizam o JRE em `~/.hubsaude/`.

---

## 5. (Opcional) Compilar a partir do código-fonte

```bash
# Pré-requisitos: Go 1.25+, JDK 21

# 1. Compilar o assinador.jar (gera assinador/target/assinador.jar)
cd assinador && ./mvnw -B -ntp package && cd ..

# 2. Compilar os CLIs
go build -o assinatura ./
go build -o simulador  ./cmd/simulador

# 3. Disponibilizar o JAR ao lado do binário
cp assinador/target/assinador.jar .
```

Para publicar binários assinados de uma nova versão, veja a seção
*Publicar uma release* no [README](../README.md).

---

## Desinstalação

Remova os binários e o diretório de estado/cache:

```bash
rm -rf ~/.hubsaude            # Linux/macOS
Remove-Item -Recurse -Force $HOME\.hubsaude   # Windows
```

---

Para o uso detalhado dos comandos, consulte o
[**Manual de Usuário**](manual-usuario.md).
