# Verification de git-filter-repo
if (-not (Get-Command git-filter-repo -ErrorAction SilentlyContinue)) {
    Write-Error "git-filter-repo n'est pas installe. Lancez : pip install git-filter-repo"
    exit 1
}

$replaceFile = [System.IO.Path]::GetTempFileName()
@"
hadahowana==>REDACTED_SECRET_P0_1
total2026==>REDACTED_SECRET_P0_2
"@ | Set-Content -Path $replaceFile -Encoding utf8

Write-Host "==> Reecriture de l'historique Git..." -ForegroundColor Yellow
git filter-repo --replace-text $replaceFile --force
Remove-Item -Path $replaceFile -Force

git reflog expire --expire=now --all
git gc --prune=now --aggressive

Write-Host "==> Historique purge. Lancez : git push origin --force --all" -ForegroundColor Green
