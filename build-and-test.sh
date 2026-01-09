#!/bin/bash

# 🚀 Neon Watchdog v2.0 - Build and Test Script
# Compila el proyecto y ejecuta verificaciones completas

set -e

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

LOG_FILE="build-test-$(date +%Y%m%d-%H%M%S).log"
PASS=0
FAIL=0
WARN=0

# Función para logging
log() {
    echo -e "$@" | tee -a "$LOG_FILE"
}

log_section() {
    log ""
    log "${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    log "${PURPLE}$1${NC}"
    log "${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

log_test() {
    local name="$1"
    local result="$2"
    local details="$3"
    
    if [ "$result" = "PASS" ]; then
        log "  ${GREEN}✅ $name${NC}"
        [ -n "$details" ] && log "     ${CYAN}→ $details${NC}"
        ((PASS++))
    elif [ "$result" = "FAIL" ]; then
        log "  ${RED}❌ $name${NC}"
        [ -n "$details" ] && log "     ${RED}→ $details${NC}"
        ((FAIL++))
    else
        log "  ${YELLOW}⚠️  $name${NC}"
        [ -n "$details" ] && log "     ${YELLOW}→ $details${NC}"
        ((WARN++))
    fi
}

# Banner
clear
log "${CYAN}"
log "╔════════════════════════════════════════════════════════════════╗"
log "║                                                                ║"
log "║           🐺 Neon Watchdog v2.0 - Build & Test               ║"
log "║                                                                ║"
log "╚════════════════════════════════════════════════════════════════╝"
log "${NC}"
log "${BLUE}Started: $(date '+%Y-%m-%d %H:%M:%S')${NC}"
log "${BLUE}Log file: $LOG_FILE${NC}"
log ""

# ============================================================================
# PASO 1: Verificar Dependencias
# ============================================================================
log_section "📋 PASO 1: Verificando Dependencias"

if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    log_test "Go instalado" "PASS" "$GO_VERSION"
else
    log_test "Go instalado" "FAIL" "Go no está instalado"
    exit 1
fi

if command -v git &> /dev/null; then
    GIT_VERSION=$(git --version | awk '{print $3}')
    log_test "Git instalado" "PASS" "$GIT_VERSION"
else
    log_test "Git instalado" "WARN" "Git no está instalado (opcional)"
fi

# ============================================================================
# PASO 2: Verificar Estructura del Proyecto
# ============================================================================
log_section "📁 PASO 2: Verificando Estructura del Proyecto"

check_file() {
    if [ -f "$1" ]; then
        local size=$(du -h "$1" | cut -f1)
        log_test "$2" "PASS" "$size"
    else
        log_test "$2" "FAIL" "Archivo no encontrado: $1"
    fi
}

check_dir() {
    if [ -d "$1" ]; then
        local count=$(find "$1" -type f | wc -l)
        log_test "$2" "PASS" "$count archivos"
    else
        log_test "$2" "FAIL" "Directorio no encontrado: $1"
    fi
}

# Verificar módulos internos
check_file "internal/config/config.go" "Config module"
check_file "internal/logger/logger.go" "Logger module"
check_file "internal/checks/checks.go" "Checks module"
check_file "internal/actions/actions.go" "Actions module"
check_file "internal/engine/engine.go" "Engine module"
check_file "internal/notifications/notifications.go" "Notifications module (v2)"
check_file "internal/metrics/metrics.go" "Metrics module (v2)"
check_file "internal/dashboard/dashboard.go" "Dashboard module (v2)"
check_file "internal/history/history.go" "History module (v2)"

# Verificar cmd
check_file "cmd/neon-watchdog/main.go" "Main application"

# Verificar ejemplos y docs
check_file "examples/config.yml" "Example config"
check_file "examples/config-v2-full.yml" "v2 Full config"
check_file "README.md" "README"
check_file "README-V2.md" "README v2"

# ============================================================================
# PASO 3: Limpiar Build Anterior
# ============================================================================
log_section "🧹 PASO 3: Limpiando Build Anterior"

if [ -f "neon-watchdog" ]; then
    rm -f neon-watchdog
    log_test "Binario anterior eliminado" "PASS"
else
    log_test "No hay binario anterior" "PASS"
fi

# ============================================================================
# PASO 4: Actualizar Dependencias
# ============================================================================
log_section "📦 PASO 4: Actualizando Dependencias"

if go mod download 2>&1 | tee -a "$LOG_FILE"; then
    log_test "go mod download" "PASS"
else
    log_test "go mod download" "FAIL"
fi

if go mod tidy 2>&1 | tee -a "$LOG_FILE"; then
    log_test "go mod tidy" "PASS"
else
    log_test "go mod tidy" "FAIL"
fi

if go mod verify 2>&1 | tee -a "$LOG_FILE"; then
    log_test "go mod verify" "PASS"
else
    log_test "go mod verify" "WARN" "Verificación de módulos falló"
fi

# ============================================================================
# PASO 5: Compilación
# ============================================================================
log_section "🔨 PASO 5: Compilando Proyecto"

log "  ${CYAN}Compilando...${NC}"
START_TIME=$(date +%s)

if go build -o neon-watchdog ./cmd/neon-watchdog 2>&1 | tee -a "$LOG_FILE"; then
    END_TIME=$(date +%s)
    COMPILE_TIME=$((END_TIME - START_TIME))
    
    if [ -f "neon-watchdog" ]; then
        SIZE=$(du -h neon-watchdog | cut -f1)
        log_test "Compilación exitosa" "PASS" "$SIZE en ${COMPILE_TIME}s"
        
        # Verificar permisos de ejecución
        if [ -x "neon-watchdog" ]; then
            log_test "Binario ejecutable" "PASS"
        else
            chmod +x neon-watchdog
            log_test "Permisos de ejecución añadidos" "PASS"
        fi
    else
        log_test "Binario generado" "FAIL" "No se creó el archivo"
    fi
else
    log_test "Compilación" "FAIL" "Ver errores arriba"
    exit 1
fi

# ============================================================================
# PASO 6: Verificar Features Implementadas
# ============================================================================
log_section "🔍 PASO 6: Verificando Features Implementadas"

check_code() {
    local feature="$1"
    local file="$2"
    local pattern="$3"
    
    if grep -q "$pattern" "$file" 2>/dev/null; then
        log_test "$feature" "PASS"
    else
        log_test "$feature" "FAIL" "Pattern '$pattern' no encontrado"
    fi
}

log "  ${BLUE}TIER 1 Features:${NC}"
check_code "  Email Notifications" "internal/notifications/notifications.go" "EmailNotifier"
check_code "  Webhook Notifications" "internal/notifications/notifications.go" "WebhookNotifier"
check_code "  Telegram Notifications" "internal/notifications/notifications.go" "TelegramNotifier"
check_code "  Prometheus Metrics" "internal/metrics/metrics.go" "neon_watchdog"
check_code "  HTTP Health Checks" "internal/checks/checks.go" "HTTPChecker"
check_code "  Action Hooks" "internal/actions/actions.go" "ActionWithHooks"

log "  ${BLUE}TIER 2 Features:${NC}"
check_code "  Dependency Chains" "internal/config/config.go" "DependsOn"
check_code "  Graceful Shutdown" "internal/config/config.go" "IgnoreExitCodes"
check_code "  Backoff Strategy" "internal/config/config.go" "BackoffStrategy"

log "  ${BLUE}TIER 3 Features:${NC}"
check_code "  Logic Checker (AND/OR)" "internal/checks/checks.go" "LogicChecker"
check_code "  Web Dashboard" "internal/dashboard/dashboard.go" "Dashboard"
check_code "  Script Checker" "internal/checks/checks.go" "ScriptChecker"
check_code "  History System" "internal/history/history.go" "RecordEvent"

# ============================================================================
# PASO 7: Validar Configuraciones
# ============================================================================
log_section "⚙️  PASO 7: Validando Configuraciones"

# Crear config temporal para test
cat > /tmp/test-neon-config.yml << 'EOF'
interval_seconds: 30
timeout_seconds: 10
log_level: INFO

default_policy:
  fail_threshold: 1
  restart_cooldown_seconds: 60
  max_restarts_per_hour: 10

targets:
  - name: test-target
    enabled: false
    checks:
      - type: tcp_port
        tcp_port: "127.0.0.1:9999"
    action:
      type: exec
      exec:
        restart: ["/bin/true"]
EOF

if ./neon-watchdog test-config -c /tmp/test-neon-config.yml 2>&1 | tee -a "$LOG_FILE"; then
    log_test "test-config comando funciona" "PASS"
else
    log_test "test-config comando" "FAIL"
fi

if [ -f "examples/config.yml" ]; then
    if ./neon-watchdog test-config -c examples/config.yml 2>&1 | tee -a "$LOG_FILE"; then
        log_test "Validación config.yml" "PASS"
    else
        log_test "Validación config.yml" "FAIL"
    fi
fi

rm -f /tmp/test-neon-config.yml

# ============================================================================
# PASO 8: Test Version Command
# ============================================================================
log_section "🏷️  PASO 8: Verificando Comandos"

if ./neon-watchdog version 2>&1 | tee -a "$LOG_FILE" | grep -q "neon-watchdog version"; then
    VERSION=$(./neon-watchdog version 2>&1 | head -1)
    log_test "version comando" "PASS" "$VERSION"
else
    log_test "version comando" "FAIL"
fi

if ./neon-watchdog help 2>&1 | tee -a "$LOG_FILE" | grep -q "USAGE"; then
    log_test "help comando" "PASS"
else
    log_test "help comando" "FAIL"
fi

# ============================================================================
# PASO 9: Verificar Tamaño y Símbolos
# ============================================================================
log_section "📊 PASO 9: Análisis del Binario"

if [ -f "neon-watchdog" ]; then
    SIZE_BYTES=$(stat -c%s "neon-watchdog" 2>/dev/null || stat -f%z "neon-watchdog" 2>/dev/null || echo "0")
    if [ "$SIZE_BYTES" != "0" ]; then
        SIZE_MB=$((SIZE_BYTES / 1024 / 1024))
        log_test "Tamaño del binario" "PASS" "${SIZE_MB}MB"
    else
        SIZE_H=$(du -h neon-watchdog | cut -f1)
        log_test "Tamaño del binario" "PASS" "$SIZE_H"
    fi
    
    if command -v file &> /dev/null; then
        FILE_INFO=$(file neon-watchdog 2>/dev/null | cut -d: -f2- || echo "ELF executable")
        log_test "Tipo de binario" "PASS" "$FILE_INFO"
    fi
    
    if command -v nm &> /dev/null; then
        SYMBOL_COUNT=$(nm neon-watchdog 2>/dev/null | wc -l || echo "N/A")
        if [ "$SYMBOL_COUNT" != "N/A" ]; then
            log_test "Símbolos en binario" "PASS" "$SYMBOL_COUNT símbolos"
        fi
    fi
fi

# ============================================================================
# PASO 10: Tests de Integración Básicos
# ============================================================================
log_section "🧪 PASO 10: Tests de Integración"

# Test: Verificar que el binario no crashea con --help
if timeout 5 ./neon-watchdog --help &>/dev/null; then
    log_test "Binario no crashea con --help" "PASS"
else
    log_test "Binario estabilidad básica" "WARN" "Timeout o error con --help"
fi

# Test: Verificar códigos de salida
if ./neon-watchdog version &>/dev/null; then
    log_test "Exit code correcto (version)" "PASS"
else
    log_test "Exit code (version)" "FAIL"
fi

# ============================================================================
# PASO 11: Verificar Git Status
# ============================================================================
log_section "📝 PASO 11: Estado del Repositorio"

if command -v git &> /dev/null && [ -d ".git" ]; then
    BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
    log_test "Git branch" "PASS" "$BRANCH"
    
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    log_test "Git commit" "PASS" "$COMMIT"
    
    if git diff --quiet 2>/dev/null; then
        log_test "Working directory limpio" "PASS"
    else
        CHANGED=$(git diff --name-only | wc -l)
        log_test "Archivos modificados" "WARN" "$CHANGED archivos sin commit"
    fi
else
    log_test "Git repository" "WARN" "No es un repositorio git"
fi

# ============================================================================
# RESUMEN FINAL
# ============================================================================
log_section "📊 RESUMEN FINAL"

TOTAL=$((PASS + FAIL + WARN))
PASS_PERCENT=$((PASS * 100 / TOTAL))

log ""
log "  ${GREEN}✅ Pasados:     $PASS${NC}"
log "  ${YELLOW}⚠️  Warnings:    $WARN${NC}"
log "  ${RED}❌ Fallidos:    $FAIL${NC}"
log "  ${BLUE}━━━━━━━━━━━━━━━━━━━━${NC}"
log "  ${CYAN}📊 Total:       $TOTAL tests${NC}"
log "  ${CYAN}📈 Success:     ${PASS_PERCENT}%${NC}"
log ""

if [ -f "neon-watchdog" ]; then
    SIZE=$(du -h neon-watchdog | cut -f1)
    log "  ${PURPLE}🎯 Binario generado: ${GREEN}neon-watchdog${NC} ${PURPLE}($SIZE)${NC}"
fi

log "  ${PURPLE}📋 Log completo: ${CYAN}$LOG_FILE${NC}"
log ""

# Resultado final
if [ $FAIL -eq 0 ]; then
    log "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    log "${GREEN}║                                                                ║${NC}"
    log "${GREEN}║           ✅ BUILD Y TESTS COMPLETADOS EXITOSAMENTE           ║${NC}"
    log "${GREEN}║                                                                ║${NC}"
    log "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    log ""
    log "${CYAN}🚀 Listo para ejecutar:${NC}"
    log "   ${YELLOW}./neon-watchdog version${NC}"
    log "   ${YELLOW}./neon-watchdog test-config -c examples/config.yml${NC}"
    log "   ${YELLOW}./neon-watchdog check -c examples/config.yml${NC}"
    log ""
    exit 0
else
    log "${RED}╔════════════════════════════════════════════════════════════════╗${NC}"
    log "${RED}║                                                                ║${NC}"
    log "${RED}║              ⚠️  ALGUNOS TESTS FALLARON                        ║${NC}"
    log "${RED}║                                                                ║${NC}"
    log "${RED}╚════════════════════════════════════════════════════════════════╝${NC}"
    log ""
    log "${YELLOW}⚠️  Revisa los errores en: $LOG_FILE${NC}"
    log ""
    exit 1
fi
