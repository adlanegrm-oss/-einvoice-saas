$ErrorActionPreference = 'Stop'

$root = 'D:\einvoice-saas-release1\einvoice-saas'

if (-not (Test-Path -LiteralPath $root)) {
throw "Projet introuvable : $root"
}

Set-Location -LiteralPath $root

Write-Host ""
Write-Host "===============================================================" -ForegroundColor Cyan
Write-Host " Fix tests - einvoice-saas" -ForegroundColor Cyan
Write-Host "===============================================================" -ForegroundColor Cyan
Write-Host ""

$go = Get-Command go -ErrorAction SilentlyContinue

if (-not $go) {
throw "Go n'est pas disponible dans le PATH."
}

go version

if ($LASTEXITCODE -ne 0) {
throw "Go ne peut pas etre execute."
}

Write-Host ""
Write-Host "[1] Verification du JWT issuer" -ForegroundColor Cyan

$jwtPath = Join-Path $root 'internal\security\jwt_validator.go'

if (-not (Test-Path -LiteralPath $jwtPath)) {
throw "Fichier introuvable : $jwtPath"
}

$src = [System.IO.File]::ReadAllText($jwtPath)

if ($src.Contains('errors.New("jwt: invalid issuer")')) {
Write-Host "JWT issuer : deja corrige." -ForegroundColor Green
}
elseif ($src -match 'claims.Issuer\s*!=\s*v.expectedIssuer\s*{\s*return nil,\s*err\s*}') {

$src = [regex]::Replace(
    $src,
    'claims\.Issuer\s*!=\s*v\.expectedIssuer\s*\{\s*return nil,\s*err\s*\}',
    'claims.Issuer != v.expectedIssuer {' + [Environment]::NewLine +
    '        return nil, errors.New("jwt: invalid issuer")' + [Environment]::NewLine +
    '    }'
)

[System.IO.File]::WriteAllText(
    $jwtPath,
    $src,
    (New-Object System.Text.UTF8Encoding($false))
)

Write-Host "JWT issuer : corrige." -ForegroundColor Green

}
else {
Write-Host "JWT issuer : aucun correctif necessaire ou motif inconnu." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "[2] Verification des tests" -ForegroundColor Cyan

$files = @(
'internal\security\jwt_validator_test.go',
'internal\security\upload_guard_test.go',
'internal\security\in_memory_keystore_cov_test.go',
'internal\middleware\auth_test.go'
)

foreach ($file in $files) {

if (-not (Test-Path -LiteralPath $file)) {
    throw "Fichier manquant : $file"
}

Write-Host "OK : $file" -ForegroundColor Green

}

Write-Host ""
Write-Host "[3] GOTMPDIR" -ForegroundColor Cyan

$env:GOTMPDIR = 'D:\tmp_go'

if (-not (Test-Path -LiteralPath $env:GOTMPDIR)) {
New-Item -ItemType Directory -Path $env:GOTMPDIR -Force | Out-Null
}

Write-Host "GOTMPDIR = $env:GOTMPDIR" -ForegroundColor Green

Write-Host ""
Write-Host "[4] gofmt" -ForegroundColor Cyan

gofmt -w internal\security internal\middleware

if ($LASTEXITCODE -ne 0) {
throw "gofmt a echoue."
}

Write-Host "gofmt : OK" -ForegroundColor Green

Write-Host ""
Write-Host "[5] go vet" -ForegroundColor Cyan

go vet ./internal/security/... ./internal/middleware/...

if ($LASTEXITCODE -ne 0) {
throw "go vet a echoue."
}

Write-Host "go vet : OK" -ForegroundColor Green

Write-Host ""
Write-Host "[6] Detection de GCC / CGO" -ForegroundColor Cyan

$gcc = Get-Command gcc -ErrorAction SilentlyContinue

if ($gcc) {

Write-Host "GCC trouve : $($gcc.Source)" -ForegroundColor Green

$env:CGO_ENABLED = '1'

$cgo = (& go env CGO_ENABLED 2>$null).Trim()

if (($LASTEXITCODE -eq 0) -and ($cgo -eq '1')) {

    Write-Host "CGO = 1" -ForegroundColor Green
    Write-Host "Execution des tests avec -race." -ForegroundColor Green

    go test ./internal/security/... ./internal/middleware/... -race -cover

    if ($LASTEXITCODE -ne 0) {
        throw "Les tests avec -race ont echoue."
    }

    $testMode = "avec -race"
}
else {

    Write-Host "CGO indisponible." -ForegroundColor Yellow
    Write-Host "Execution des tests sans -race." -ForegroundColor Yellow

    go test ./internal/security/... ./internal/middleware/... -cover

    if ($LASTEXITCODE -ne 0) {
        throw "Les tests ont echoue."
    }

    $testMode = "sans -race"
}

}
else {

Write-Host "GCC absent du PATH." -ForegroundColor Yellow
Write-Host "Execution des tests sans -race." -ForegroundColor Yellow

go test ./internal/security/... ./internal/middleware/... -cover

if ($LASTEXITCODE -ne 0) {
    throw "Les tests ont echoue."
}

$testMode = "sans -race"

}

Write-Host ""
Write-Host "===============================================================" -ForegroundColor Green
Write-Host " SUCCES" -ForegroundColor Green
Write-Host "===============================================================" -ForegroundColor Green
Write-Host ""
Write-Host "Go          : OK" -ForegroundColor Green
Write-Host "JWT issuer  : OK" -ForegroundColor Green
Write-Host "Tests       : OK ($testMode)" -ForegroundColor Green
Write-Host "gofmt       : OK" -ForegroundColor Green
Write-Host "go vet      : OK" -ForegroundColor Green
Write-Host ""
