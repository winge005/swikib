param(
    [string]$TursoToken,
    [string]$AuthToken = $env:SWIKI_AUTH_TOKEN,
    [string]$CorsOrigins = $env:SWIKI_CORS_ORIGINS,
    [switch]$SkipRun
)

$ErrorActionPreference = 'Stop'

$configPath = Join-Path -Path $PSScriptRoot -ChildPath 'config.yml'

if ([string]::IsNullOrWhiteSpace($TursoToken)) {
    if (-not (Test-Path -LiteralPath $configPath)) {
        throw "config.yml not found at $configPath"
    }

    $config = Get-Content -LiteralPath $configPath -Raw
    $match = [regex]::Match($config, '(?m)^\s*accesstoken\s*:\s*([^\r\n#]+)')
    if ($match.Success) {
        $TursoToken = $match.Groups[1].Value.Trim().Trim('"').Trim("'")
    }
}

if ([string]::IsNullOrWhiteSpace($TursoToken)) {
    throw 'No Turso token found. Set it in config.yml or pass -TursoToken.'
}

$env:SWIKI_TURSO_TOKEN = $TursoToken

if (-not [string]::IsNullOrWhiteSpace($AuthToken)) {
    $env:SWIKI_AUTH_TOKEN = $AuthToken
}

if (-not [string]::IsNullOrWhiteSpace($CorsOrigins)) {
    $env:SWIKI_CORS_ORIGINS = $CorsOrigins
}

if ($SkipRun) {
    return
}

Push-Location -LiteralPath $PSScriptRoot
try {
    & go run .
}
finally {
    Pop-Location
}
