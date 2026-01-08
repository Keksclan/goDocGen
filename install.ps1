# goDocGen Installer Script
# Fügt das aktuelle Verzeichnis zur PATH-Umgebungsvariable hinzu

$currentDir = $PSScriptRoot
if (-not $currentDir) { $currentDir = Get-Location }

$exeName = "godocgen.exe"
$exePath = Join-Path $currentDir $exeName

if (-not (Test-Path $exePath)) {
    Write-Host "[WARN] $exeName wurde im aktuellen Verzeichnis nicht gefunden." -ForegroundColor Yellow
    Write-Host "Versuche das Programm zu bauen..." -ForegroundColor Cyan
    
    & go build -o $exeName ./cmd/godocgen
    
    if (-not $?) {
        Write-Host "[ERROR] Build fehlgeschlagen. Bitte stelle sicher, dass Go installiert ist." -ForegroundColor Red
        exit 1
    }
    Write-Host "[OK] Build erfolgreich." -ForegroundColor Green
}

Write-Host "[INFO] Verzeichnis: $currentDir" -ForegroundColor Cyan

# Pfad zur User-Variable hinzufügen
$pathType = [EnvironmentVariableTarget]::User
$oldPath = [Environment]::GetEnvironmentVariable("Path", $pathType)

if ($oldPath -split ";" -contains $currentDir) {
    Write-Host "[INFO] Das Verzeichnis ist bereits in der PATH-Variable enthalten." -ForegroundColor Green
} else {
    Write-Host "[CONFIG] Füge Verzeichnis zum PATH hinzu..." -ForegroundColor Cyan
    $newPath = "$oldPath;$currentDir".Trim(';')
    [Environment]::SetEnvironmentVariable("Path", $newPath, $pathType)
    
    # Auch für die aktuelle Sitzung aktualisieren
    $env:Path = "$env:Path;$currentDir"
    
    Write-Host "[OK] Erfolgreich zum PATH hinzugefügt!" -ForegroundColor Green
    Write-Host "[INFO] Bitte starte dein Terminal neu, um 'godocgen' von überall aus nutzen zu können." -ForegroundColor Cyan
}
