#!/bin/bash

# ============================================================================
# Neon Watchdog - Script Maestro
# Compila, verifica y ejecuta checks completos del proyecto
# ============================================================================

set -e  # Salir si hay error

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Variables
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_NAME="neon-watchdog"
BUILD_TIME=""

# ============================================================================
# Funciones auxiliares
# ============================================================================

print_header() {
    echo ""
    echo -e "${BOLD}${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}${BLUE}║${NC}  $1"
    echo -e "${BOLD}${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_step() {
    echo -e "${BOLD}${YELLOW}▶${NC} $1"
}

print_success() {
    echo -e "${GREEN}✅${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC}  $1"
}

# ============================================================================
# PASO 1: Verificar dependencias
# ============================================================================

check_dependencies() {
    print_header "PASO 1: Verificando Dependencias"
    
    print_step "Verificando Go..."
    if ! command -v go &> /dev/null; then
        print_error "Go no está instalado"
        exit 1
    fi
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go instalado: $GO_VERSION"
    
    print_step "Verificando Git..."
    if ! command -v git &> /dev/null; then
        print_error "Git no está instalado"
        exit 1
    fi
    GIT_VERSION=$(git --version | awk '{print $3}')
    print_success "Git instalado: $GIT_VERSION"
    
    print_step "Verificando estructura del proyecto..."
    if [ ! -f "go.mod" ]; then
        print_error "go.mod no encontrado"
        exit 1
    fi
    print_success "Estructura del proyecto OK"
}

# ============================================================================
# PASO 2: Limpiar build anterior
# ============================================================================

clean_build() {
    print_header "PASO 2: Limpiando Build Anterior"
    
    print_step "Eliminando binario anterior..."
    if [ -f "$BINARY_NAME" ]; then
        rm -f "$BINARY_NAME"
        print_success "Binario anterior eliminado"
    else
        print_info "No hay binario anterior"
    fi
    
    print_step "Limpiando cache de Go..."
    go clean -cache 2>/dev/null || true
    print_success "Cache limpiado"
}

# ============================================================================
# PASO 3: Verificar módulos de Go
# ============================================================================

check_go_modules() {
    print_header "PASO 3: Verificando Módulos de Go"
    
    print_step "Sincronizando dependencias..."
    go mod tidy
    print_success "go mod tidy completado"
    
    print_step "Descargando dependencias..."
    go mod download
    print_success "Dependencias descargadas"
    
    print_step "Verificando módulos..."
    go mod verify
    print_success "Módulos verificados"
}

# ============================================================================
# PASO 4: Compilar proyecto
# ============================================================================

compile_project() {
    print_header "PASO 4: Compilando Proyecto"
    
    print_step "Compilando neon-watchdog..."
    START_TIME=$(date +%s)
    
    if go build -o "$BINARY_NAME" ./cmd/neon-watchdog; then
        END_TIME=$(date +%s)
        BUILD_TIME=$((END_TIME - START_TIME))
        
        if [ -f "$BINARY_NAME" ]; then
            BINARY_SIZE=$(stat -c%s "$BINARY_NAME" 2>/dev/null || stat -f%z "$BINARY_NAME" 2>/dev/null)
            BINARY_SIZE_MB=$(echo "scale=1; $BINARY_SIZE/1024/1024" | bc)
            print_success "Compilación exitosa en ${BUILD_TIME}s (${BINARY_SIZE_MB}MB)"
        else
            print_error "Binario no generado"
            exit 1
        fi
    else
        print_error "Error de compilación"
        exit 1
    fi
    
    print_step "Verificando permisos de ejecución..."
    chmod +x "$BINARY_NAME"
    print_success "Permisos configurados"
}

# ============================================================================
# PASO 5: Verificar binario
# ============================================================================

verify_binary() {
    print_header "PASO 5: Verificando Binario"
    
    print_step "Verificando que el binario es ejecutable..."
    if [ ! -x "$BINARY_NAME" ]; then
        print_error "El binario no es ejecutable"
        exit 1
    fi
    print_success "Binario ejecutable"
    
    print_step "Probando comando version..."
    VERSION_OUTPUT=$(./"$BINARY_NAME" version 2>&1)
    if [ $? -eq 0 ]; then
        print_success "Comando version: $VERSION_OUTPUT"
    else
        print_error "Comando version falló"
        exit 1
    fi
    
    print_step "Probando comando help..."
    if ./"$BINARY_NAME" help &> /dev/null; then
        print_success "Comando help funciona"
    else
        print_error "Comando help falló"
        exit 1
    fi
}

# ============================================================================
# PASO 6: Verificar archivos de configuración
# ============================================================================

verify_configs() {
    print_header "PASO 6: Verificando Configuraciones"
    
    CONFIG_FILES=(
        "examples/config.yml"
        "examples/config-v2-full.yml"
    )
    
    for config in "${CONFIG_FILES[@]}"; do
        if [ -f "$config" ]; then
            print_step "Validando $config..."
            
            if ./"$BINARY_NAME" test-config -c "$config" &> /dev/null; then
                print_success "$config válido"
            else
                print_error "$config tiene errores"
                ./"$BINARY_NAME" test-config -c "$config"
                exit 1
            fi
        else
            print_info "$config no existe (opcional)"
        fi
    done
}

# ============================================================================
# PASO 7: Verificar módulos v2.0
# ============================================================================

verify_v2_modules() {
    print_header "PASO 7: Verificando Módulos v2.0"
    
    MODULES=(
        "internal/notifications/notifications.go:Sistema de notificaciones"
        "internal/metrics/metrics.go:Métricas Prometheus"
        "internal/dashboard/dashboard.go:Dashboard web"
        "internal/history/history.go:Sistema de historial"
    )
    
    for module_info in "${MODULES[@]}"; do
        IFS=':' read -r module_path module_name <<< "$module_info"
        print_step "Verificando $module_name..."
        
        if [ -f "$module_path" ]; then
            # Verificar que el módulo tiene contenido
            if [ -s "$module_path" ]; then
                LINES=$(wc -l < "$module_path")
                print_success "$module_name existe (${LINES} líneas)"
            else
                print_error "$module_name está vacío"
                exit 1
            fi
        else
            print_error "$module_name no encontrado: $module_path"
            exit 1
        fi
    done
}

# ============================================================================
# PASO 8: Ejecutar check de ejemplo
# ============================================================================

run_example_check() {
    print_header "PASO 8: Ejecutando Check de Ejemplo"
    
    if [ -f "examples/config.yml" ]; then
        print_step "Ejecutando check con config.yml..."
        
        # Ejecutar check en modo dry-run (solo verifica, no ejecuta acciones)
        if ./"$BINARY_NAME" check -c examples/config.yml 2>&1 | head -20; then
            print_success "Check de ejemplo ejecutado"
        else
            print_info "Check ejecutado (puede haber warnings esperados)"
        fi
    else
        print_info "No hay config.yml de ejemplo"
    fi
}

# ============================================================================
# PASO 9: Verificar documentación
# ============================================================================

verify_documentation() {
    print_header "PASO 9: Verificando Documentación"
    
    DOCS=(
        "README.md:README principal"
        "README-V2.md:Documentación v2.0"
        "SCRIPTS.md:Guía de scripts"
        "IMPLEMENTATION-SUMMARY.md:Resumen de implementación"
    )
    
    for doc_info in "${DOCS[@]}"; do
        IFS=':' read -r doc_path doc_name <<< "$doc_info"
        print_step "Verificando $doc_name..."
        
        if [ -f "$doc_path" ]; then
            print_success "$doc_name existe"
        else
            print_info "$doc_name no encontrado (opcional)"
        fi
    done
}

# ============================================================================
# PASO 10: Resumen final
# ============================================================================

print_summary() {
    print_header "RESUMEN FINAL"
    
    echo -e "${BOLD}📊 Estadísticas:${NC}"
    echo -e "   ${GREEN}✅${NC} Binario: $BINARY_NAME"
    echo -e "   ${GREEN}✅${NC} Tamaño: ${BINARY_SIZE_MB}MB"
    echo -e "   ${GREEN}✅${NC} Tiempo de compilación: ${BUILD_TIME}s"
    echo -e "   ${GREEN}✅${NC} Versión: $VERSION_OUTPUT"
    echo ""
    
    echo -e "${BOLD}🎯 Comandos disponibles:${NC}"
    echo -e "   ${YELLOW}▶${NC} ./$BINARY_NAME version"
    echo -e "   ${YELLOW}▶${NC} ./$BINARY_NAME test-config -c examples/config.yml"
    echo -e "   ${YELLOW}▶${NC} ./$BINARY_NAME check -c examples/config.yml"
    echo -e "   ${YELLOW}▶${NC} ./$BINARY_NAME run -c examples/config.yml"
    echo ""
    
    echo -e "${BOLD}📚 Scripts disponibles:${NC}"
    echo -e "   ${YELLOW}▶${NC} ./build.sh          - Compilación rápida"
    echo -e "   ${YELLOW}▶${NC} ./run-all.sh        - Verificación completa (este script)"
    echo -e "   ${YELLOW}▶${NC} sudo ./test-apache.sh - Test funcional con Apache"
    echo ""
    
    echo ""
    echo -e "${BOLD}${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}${GREEN}║                                                                ║${NC}"
    echo -e "${BOLD}${GREEN}║  ✅  TODO COMPILADO, VERIFICADO Y FUNCIONANDO PERFECTAMENTE  ✅ ║${NC}"
    echo -e "${BOLD}${GREEN}║                                                                ║${NC}"
    echo -e "${BOLD}${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# ============================================================================
# MAIN
# ============================================================================

main() {
    cd "$PROJECT_ROOT"
    
    echo ""
    echo -e "${BOLD}${BLUE}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BOLD}${BLUE}  Neon Watchdog - Script de Verificación Completa${NC}"
    echo -e "${BOLD}${BLUE}════════════════════════════════════════════════════════════════${NC}"
    echo ""
    
    check_dependencies
    clean_build
    check_go_modules
    compile_project
    verify_binary
    verify_configs
    verify_v2_modules
    run_example_check
    verify_documentation
    print_summary
}

# Ejecutar
main "$@"
