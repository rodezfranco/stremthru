# Script para actualizar todas las importaciones del módulo Go
# Reemplaza github.com/MunifTanjim/stremthru con tu nuevo repositorio

param(
    [Parameter(Mandatory=$true)]
    [string]$NewRepo
)

Write-Host "Actualizando importaciones de github.com/MunifTanjim/stremthru a $NewRepo..."

# Buscar todos los archivos .go
$goFiles = Get-ChildItem -Path . -Recurse -Filter "*.go"

foreach ($file in $goFiles) {
    $content = Get-Content $file.FullName -Raw
    if ($content -match "github\.com/MunifTanjim/stremthru") {
        Write-Host "Actualizando: $($file.FullName)"
        $newContent = $content -replace "github\.com/MunifTanjim/stremthru", $NewRepo
        Set-Content -Path $file.FullName -Value $newContent -NoNewline
    }
}

Write-Host "Actualización completada!"
Write-Host "Ejecuta 'go mod tidy' para actualizar las dependencias."
