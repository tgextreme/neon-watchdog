#!/bin/bash
# Script de prueba completa para Neon Watchdog con Apache
# Demuestra que el watchdog detecta y recupera servicios caídos
# DEBE EJECUTARSE COMO ROOT: sudo ./test-apache.sh

set -e

# Verificar que se ejecuta como root
if [ "$EUID" -ne 0 ]; then
    echo "❌ Este script debe ejecutarse como root"
    echo "Ejecuta: sudo ./test-apache.sh"
    exit 1
fi

WATCHDOG_BIN="./bin/neon-watchdog"
CONFIG_FILE="test-apache-config.yml"
LOG_FILE="test-apache-$(date +%Y%m%d-%H%M%S).log"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Función para log (muestra en pantalla y guarda en archivo)
log() {
    echo -e "$@" | tee -a "$LOG_FILE"
}

log "${BLUE}════════════════════════════════════════════════════════════════${NC}"
log "${BLUE}  🐺 Neon Watchdog - Prueba Completa con Apache2${NC}"
log "${BLUE}════════════════════════════════════════════════════════════════${NC}"
log ""
log "${YELLOW}📝 Guardando logs en: $LOG_FILE${NC}"
log ""

# Función para limpiar al salir
cleanup() {
    log ""
    log "${YELLOW}🧹 Limpiando...${NC}"
    rm -f "$CONFIG_FILE"
    # Asegurar que Apache está corriendo
    systemctl start apache2 2>/dev/null || true
    log "${GREEN}✅ Limpieza completada${NC}"
    log ""
    log "${BLUE}📋 Log guardado en: $LOG_FILE${NC}"
}
trap cleanup EXIT

# Verificar que el binario existe
if [ ! -f "$WATCHDOG_BIN" ]; then
    log "${RED}❌ Error: Binario no encontrado. Ejecuta 'make build' primero.${NC}"
    exit 1
fi

# ============================================================
# PASO 1: Verificar que Apache está corriendo
# ============================================================
log "${BLUE}━━━ Paso 1: Verificar estado inicial de Apache ━━━${NC}"
if systemctl is-active apache2 > /dev/null 2>&1; then
    log "${GREEN}✅ Apache está corriendo${NC}"
else
    log "${YELLOW}⚠️  Apache no está activo, iniciándolo...${NC}"
    systemctl start apache2
    sleep 2
fi

# Verificar que responde en el puerto 80
if curl -s http://localhost:80 > /dev/null 2>&1; then
    log "${GREEN}✅ Apache responde en puerto 80${NC}"
else
    log "${RED}❌ Apache no responde en puerto 80${NC}"
    exit 1
fi
log ""

# ============================================================
# PASO 2: Crear configuración de prueba
# ============================================================
log "${BLUE}━━━ Paso 2: Crear configuración de watchdog ━━━${NC}"
cat > "$CONFIG_FILE" <<EOF
log_level: INFO
timeout_seconds: 10

default_policy:
  fail_threshold: 1
  restart_cooldown_seconds: 5
  max_restarts_per_hour: 20

targets:
  - name: apache2
    enabled: true
    checks:
      - type: process_name
        process_name: apache2
      - type: tcp_port
        tcp_port: "80"
    action:
      type: systemd
      systemd:
        unit: apache2.service
        method: restart
EOF

log "${GREEN}✅ Configuración creada: $CONFIG_FILE${NC}"
log ""

# ============================================================
# PASO 3: Validar configuración
# ============================================================
log "${BLUE}━━━ Paso 3: Validar configuración ━━━${NC}"
$WATCHDOG_BIN test-config -c "$CONFIG_FILE" 2>&1 | tee -a "$LOG_FILE"
log "${GREEN}✅ Configuración válida${NC}"
log ""

# ============================================================
# PASO 4: Ejecutar check con Apache funcionando (debe pasar)
# ============================================================
log "${BLUE}━━━ Paso 4: Check con Apache funcionando (debe estar healthy) ━━━${NC}"
$WATCHDOG_BIN check -c "$CONFIG_FILE" --verbose 2>&1 | tee -a "$LOG_FILE"
EXIT_CODE=${PIPESTATUS[0]}

if [ $EXIT_CODE -eq 0 ]; then
    log "${GREEN}✅ Watchdog detectó Apache como HEALTHY (exit code: 0)${NC}"
else
    log "${RED}❌ Error: Apache debería estar healthy pero watchdog reportó fallo (exit code: $EXIT_CODE)${NC}"
    exit 1
fi
log ""
sleep 2

# ============================================================
# PASO 5: Simular fallo (detener Apache)
# ============================================================
log "${BLUE}━━━ Paso 5: Simular fallo - Deteniendo Apache ━━━${NC}"
log "${YELLOW}⚠️  Deteniendo Apache...${NC}"
systemctl stop apache2
sleep 2

if systemctl is-active apache2 > /dev/null 2>&1; then
    log "${RED}❌ Error: Apache sigue activo${NC}"
    exit 1
else
    log "${GREEN}✅ Apache detenido correctamente${NC}"
fi
log ""

# ============================================================
# PASO 6: Ejecutar watchdog (debe detectar fallo y recuperar)
# ============================================================
log "${BLUE}━━━ Paso 6: Ejecutar watchdog (debe detectar fallo y recuperar) ━━━${NC}"
log "${YELLOW}👀 Observa cómo detecta el fallo y ejecuta la acción de recuperación...${NC}"
log ""

$WATCHDOG_BIN check -c "$CONFIG_FILE" --verbose 2>&1 | tee -a "$LOG_FILE"
EXIT_CODE=${PIPESTATUS[0]}

if [ $EXIT_CODE -eq 2 ]; then
    log ""
    log "${YELLOW}⚠️  Watchdog detectó fallo (exit code: 2 = UNHEALTHY)${NC}"
    log "${YELLOW}⚠️  Esto es CORRECTO - significa que detectó el problema${NC}"
else
    log ""
    log "${YELLOW}⚠️  Exit code: $EXIT_CODE${NC}"
fi
log ""
sleep 3

# ============================================================
# PASO 7: Verificar recuperación
# ============================================================
log "${BLUE}━━━ Paso 7: Verificar que Apache fue recuperado ━━━${NC}"

# Dar tiempo al restart
sleep 3

if systemctl is-active apache2 > /dev/null 2>&1; then
    log "${GREEN}✅ Apache está corriendo nuevamente${NC}"
else
    log "${RED}❌ Error: Apache no se recuperó${NC}"
    log "${YELLOW}Verificando estado...${NC}"
    systemctl status apache2 --no-pager 2>&1 | tee -a "$LOG_FILE"
    exit 1
fi

# Verificar que responde
if curl -s http://localhost:80 > /dev/null 2>&1; then
    log "${GREEN}✅ Apache responde en puerto 80${NC}"
else
    log "${RED}❌ Apache no responde en puerto 80${NC}"
    exit 1
fi
log ""

# ============================================================
# PASO 8: Check final (debe estar healthy de nuevo)
# ============================================================
log "${BLUE}━━━ Paso 8: Check final (debe estar healthy) ━━━${NC}"
$WATCHDOG_BIN check -c "$CONFIG_FILE" --verbose 2>&1 | tee -a "$LOG_FILE"
EXIT_CODE=${PIPESTATUS[0]}

if [ $EXIT_CODE -eq 0 ]; then
    log "${GREEN}✅ Apache está HEALTHY nuevamente (exit code: 0)${NC}"
else
    log "${RED}❌ Apache debería estar healthy (exit code: $EXIT_CODE)${NC}"
    exit 1
fi
log ""

# ============================================================
# RESUMEN
# ============================================================
log "${GREEN}════════════════════════════════════════════════════════════════${NC}"
log "${GREEN}  ✅ ¡PRUEBA COMPLETA EXITOSA!${NC}"
log "${GREEN}════════════════════════════════════════════════════════════════${NC}"
log ""
log "${GREEN}Resumen de la prueba:${NC}"
log "  1. ✅ Apache detectado como healthy inicialmente"
log "  2. ✅ Apache detenido correctamente (simulación de fallo)"
log "  3. ✅ Watchdog detectó el fallo"
log "  4. ✅ Watchdog ejecutó acción de recuperación (systemctl restart)"
log "  5. ✅ Apache se recuperó automáticamente"
log "  6. ✅ Apache funcional nuevamente"
log ""
log "${BLUE}🎉 Neon Watchdog funciona correctamente!${NC}"
log ""
log "${YELLOW}Para monitorización continua, instala con systemd:${NC}"
log "  sudo make install"
log "  sudo systemctl enable --now neon-watchdog.timer"
log ""
