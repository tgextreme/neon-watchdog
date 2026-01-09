#!/bin/bash

# 🚀 Neon Watchdog - Quick Build and Test
# Script simplificado para compilar y verificar el proyecto

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════════════╗"
echo "║     🐺 Neon Watchdog v2.0 - Build & Test Script      ║"
echo "╚════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo ""

PASS=0
FAIL=0

test_pass() {
    echo -e "${GREEN}✅ $1${NC}"
    ((PASS++))
}

test_fail() {
    echo -e "${RED}❌ $1${NC}"
    ((FAIL++))
}

# 1. Verificar Go
echo -e "${BLUE}━━━ PASO 1: Verificando Go ━━━${NC}"
if command -v go &> /dev/null; then
    test_pass "Go instalado: $(go version | awk '{print $3}')"
else
    test_fail "Go no está instalado"
    exit 1
fi
echo ""

# 2. Limpiar build anterior
echo -e "${BLUE}━━━ PASO 2: Limpiando ━━━${NC}"
rm -f neon-watchdog
test_pass "Limpieza completada"
echo ""

# 3. Actualizar dependencias
echo -e "${BLUE}━━━ PASO 3: Actualizando dependencias ━━━${NC}"
if go mod tidy; then
    test_pass "go mod tidy"
else
    test_fail "go mod tidy"
fi
echo ""

# 4. Compilar
echo -e "${BLUE}━━━ PASO 4: Compilando proyecto ━━━${NC}"
START=$(date +%s)
if go build -o neon-watchdog ./cmd/neon-watchdog; then
    END=$(date +%s)
    DURATION=$((END - START))
    SIZE=$(du -h neon-watchdog | cut -f1)
    test_pass "Compilación exitosa en ${DURATION}s → $SIZE"
else
    test_fail "Error en compilación"
    exit 1
fi
echo ""

# 5. Verificar binario
echo -e "${BLUE}━━━ PASO 5: Verificando binario ━━━${NC}"
if [ -f "neon-watchdog" ] && [ -x "neon-watchdog" ]; then
    test_pass "Binario ejecutable creado"
else
    test_fail "Binario no creado o no ejecutable"
    exit 1
fi
echo ""

# 6. Test de comandos
echo -e "${BLUE}━━━ PASO 6: Probando comandos ━━━${NC}"

if ./neon-watchdog version &>/dev/null; then
    VERSION=$(./neon-watchdog version 2>&1 | head -1)
    test_pass "Comando version: $VERSION"
else
    test_fail "Comando version"
fi

if ./neon-watchdog help &>/dev/null; then
    test_pass "Comando help"
else
    test_fail "Comando help"
fi

# 7. Validar configuraciones
echo ""
echo -e "${BLUE}━━━ PASO 7: Validando configuraciones ━━━${NC}"

if [ -f "examples/config.yml" ]; then
    if ./neon-watchdog test-config -c examples/config.yml &>/dev/null; then
        test_pass "Validación de examples/config.yml"
    else
        test_fail "Validación de examples/config.yml"
    fi
fi

if [ -f "examples/config-v2-full.yml" ]; then
    if ./neon-watchdog test-config -c examples/config-v2-full.yml &>/dev/null; then
        test_pass "Validación de examples/config-v2-full.yml"
    else
        test_fail "Validación de examples/config-v2-full.yml"
    fi
fi

# 8. Verificar módulos v2
echo ""
echo -e "${BLUE}━━━ PASO 8: Verificando módulos v2.0 ━━━${NC}"

check_module() {
    if [ -f "$1" ]; then
        LINES=$(wc -l < "$1")
        test_pass "$(basename $1) ($LINES líneas)"
    else
        test_fail "$(basename $1) no encontrado"
    fi
}

check_module "internal/notifications/notifications.go"
check_module "internal/metrics/metrics.go"
check_module "internal/dashboard/dashboard.go"
check_module "internal/history/history.go"

# Resumen final
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}📊 RESUMEN:${NC}"
echo -e "   ${GREEN}✅ Pasados: $PASS${NC}"
echo -e "   ${RED}❌ Fallidos: $FAIL${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   ✅ BUILD COMPLETADO EXITOSAMENTE            ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${YELLOW}🚀 Comandos disponibles:${NC}"
    echo -e "   ${CYAN}./neon-watchdog version${NC}"
    echo -e "   ${CYAN}./neon-watchdog test-config -c examples/config.yml${NC}"
    echo -e "   ${CYAN}./neon-watchdog check -c examples/config.yml${NC}"
    echo -e "   ${CYAN}sudo ./test-apache.sh${NC} ${BLUE}(test completo con Apache)${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║   ⚠️  ALGUNOS TESTS FALLARON                  ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════╝${NC}"
    echo ""
    exit 1
fi
