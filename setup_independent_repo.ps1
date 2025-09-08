# Script para convertir el proyecto a un repositorio independiente
# Ejecutar desde la raíz del proyecto

Write-Host "=== Convirtiendo proyecto a repositorio independiente ==="

# Eliminar historial de git existente
if (Test-Path ".git") {
    Write-Host "Eliminando historial de git existente..."
    Remove-Item -Recurse -Force ".git"
}

# Inicializar nuevo repositorio git
Write-Host "Inicializando nuevo repositorio git..."
git init

# Crear .gitignore si no existe
if (-not (Test-Path ".gitignore")) {
    Write-Host "Creando .gitignore..."
    @"
# Binarios
*.exe
*.dll
*.so
*.dylib

# Archivos de test
*.test

# Archivos de cobertura
*.out

# Dependencias de vendor
vendor/

# Archivos de configuración local
.env
.env.local

# Archivos de base de datos
*.db
*.sqlite

# Logs
*.log

# Directorio de datos
data/

# Archivos temporales
tmp/
temp/

# IDEs
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db
"@ | Out-File -FilePath ".gitignore" -Encoding utf8
}

# Agregar todos los archivos
Write-Host "Agregando archivos al repositorio..."
git add .

# Commit inicial
Write-Host "Creando commit inicial..."
git commit -m "Initial commit: Personal StremThru project"

Write-Host ""
Write-Host "=== Próximos pasos ==="
Write-Host "1. Crear un nuevo repositorio en GitHub"
Write-Host "2. Ejecutar: git remote add origin https://github.com/tuusuario/tu-proyecto-stremthru.git"
Write-Host "3. Ejecutar: git branch -M main"
Write-Host "4. Ejecutar: git push -u origin main"
Write-Host ""
Write-Host "¡Proyecto convertido exitosamente!"
