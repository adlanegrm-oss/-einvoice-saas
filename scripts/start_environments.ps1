# Démarre les 3 environnements en natif (sans Docker) pour le développement Windows.
# Les secrets ne sont PAS dans ce fichier : définissez-les dans votre session avant de lancer,
# par exemple :
#   $env:ADMIN_PASSWORD = "un-mot-de-passe-solide"
#   $env:JWT_SECRET     = "<au moins 32 caractères aléatoires>"
# Sinon, des valeurs temporaires sont générées pour cette session.

function New-Secret([int]$bytes) {
    $b = New-Object byte[] $bytes
    [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($b)
    return [Convert]::ToBase64String($b)
}

if (-not $env:JWT_SECRET)     { $env:JWT_SECRET = New-Secret 48; Write-Host "JWT_SECRET temporaire généré pour cette session." -ForegroundColor DarkYellow }
if (-not $env:ADMIN_PASSWORD) { $env:ADMIN_PASSWORD = New-Secret 12; Write-Host "ADMIN_PASSWORD temporaire (session uniquement) : $env:ADMIN_PASSWORD" -ForegroundColor DarkYellow }
if (-not $env:ADMIN_EMAIL)    { $env:ADMIN_EMAIL = "admin@example.com" }

$root = Split-Path -Parent $PSScriptRoot

$environments = @(
    @{ Name = "RECETTE";  Port = "8081"; Db = "invoices_recette.db";  Color = "Green"  },
    @{ Name = "PREPROD";  Port = "8082"; Db = "invoices_preprod.db";  Color = "Yellow" },
    @{ Name = "PROD";     Port = "8080"; Db = "invoices.db";          Color = "Red"    }
)

foreach ($e in $environments) {
    Write-Host "-> Lancement $($e.Name) sur le port $($e.Port)..." -ForegroundColor $e.Color
    # Chaque environnement a son dossier de données (base + archives) isolé.
    $cmd = "`$env:PORT='$($e.Port)'; `$env:APP_ENV='$($e.Name)'; `$env:DB_NAME='$($e.Db)'; " +
           "`$env:DATA_DIR='data\$($e.Name.ToLower())'; " +
           "`$env:PUBLIC_BASE_URL='http://localhost:$($e.Port)'; " +
           "Set-Location '$root'; go run .\cmd\server"
    Start-Process powershell -ArgumentList "-NoExit", "-Command", $cmd
}

Write-Host "`nEnvironnements lancés :" -ForegroundColor Cyan
foreach ($e in $environments) { Write-Host "  - $($e.Name) : http://localhost:$($e.Port)" -ForegroundColor $e.Color }
