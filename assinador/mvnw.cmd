@echo off
setlocal

set MAVEN_VERSION=3.9.9
set MAVEN_DIST_DIR=%USERPROFILE%\.m2\wrapper\dists\apache-maven-%MAVEN_VERSION%
set MVN_CMD=%MAVEN_DIST_DIR%\bin\mvn.cmd

if not exist "%MVN_CMD%" (
    echo Baixando Maven %MAVEN_VERSION%...
    powershell -NoProfile -Command ^
        "$ver = '3.9.9';" ^
        "$dst = [System.IO.Path]::Combine($env:USERPROFILE, '.m2', 'wrapper', 'dists');" ^
        "$zip = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), 'maven.zip');" ^
        "$url = 'https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/' + $ver + '/apache-maven-' + $ver + '-bin.zip';" ^
        "Write-Host ('URL: ' + $url);" ^
        "New-Item -ItemType Directory -Force -Path $dst | Out-Null;" ^
        "Invoke-WebRequest -Uri $url -OutFile $zip;" ^
        "Expand-Archive -Path $zip -DestinationPath $dst -Force;" ^
        "Remove-Item $zip;" ^
        "Write-Host 'Maven instalado com sucesso.'"
    if errorlevel 1 (
        echo Falha ao baixar o Maven. Verifique sua conexao com a internet.
        exit /b 1
    )
)

"%MVN_CMD%" %*
