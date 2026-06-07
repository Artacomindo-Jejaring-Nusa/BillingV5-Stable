# run-sonar.ps1
# Helper script to run unit tests and trigger SonarScanner scan

$SONAR_TOKEN = $args[0]
if (-not $SONAR_TOKEN) {
    Write-Host "Error: SonarQube Token is required." -ForegroundColor Red
    Write-Host "Usage: .\run-sonar.ps1 <YOUR_SONAR_TOKEN>"
    exit 1
}

$PROJECT_ROOT = Get-Location

# 1. Run Backend Tests & Coverage
Write-Host "--- Running Backend (Go) Tests ---" -ForegroundColor Cyan
Push-Location backend
go test -v "./..." -coverprofile="coverage.out"
if ($LASTEXITCODE -ne 0) { Write-Host "Backend tests failed!"; Pop-Location; exit 1 }
Pop-Location

# 2. Run Frontend Tests & Coverage
Write-Host "--- Running Frontend (Vue) Tests ---" -ForegroundColor Cyan
Push-Location frontend
if (-not (Test-Path "node_modules")) { 
    Write-Host "node_modules not found, running npm install..." -ForegroundColor Yellow
    npm install
}
npm run test:coverage
if ($LASTEXITCODE -ne 0) { Write-Host "Frontend tests failed!"; Pop-Location; exit 1 }
Pop-Location

# 3. Trigger SonarScanner Scan via Docker
Write-Host "--- Triggering SonarScanner Scan (Debug Mode) ---" -ForegroundColor Cyan
docker run --rm `
    -e SONAR_HOST_URL="http://host.docker.internal:9000" `
    -e SONAR_TOKEN="$SONAR_TOKEN" `
    -v "${PROJECT_ROOT}:/usr/src" `
    sonarsource/sonar-scanner-cli -X

Write-Host "--- Scan Complete! Check http://localhost:9000 ---" -ForegroundColor Green
