#!/bin/bash
# Script para ejecutar Neon Watchdog en modo local
# Configuración completa en config-local.yml

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colores para output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Neon Watchdog - Ejecución Local     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# Verificar que existe el binario
if [ ! -f "./neon-watchdog" ]; then
    echo -e "${RED}✗ Error: No se encuentra el binario 'neon-watchdog'${NC}"
    echo -e "${YELLOW}Compila el proyecto con: make build${NC}"
    exit 1
fi

# Verificar que existe la configuración local
if [ ! -f "./config-local.yml" ]; then
    echo -e "${RED}✗ Error: No se encuentra 'config-local.yml'${NC}"
    exit 1
fi

# Verificar que existe users.txt
if [ ! -f "./users.txt" ]; then
    echo -e "${YELLOW}⚠ Advertencia: No se encuentra 'users.txt'${NC}"
    echo -e "${YELLOW}Se usarán credenciales por defecto: admin:admin123${NC}"
    echo ""
fi

# Matar instancias previas
echo -e "${YELLOW}→ Deteniendo instancias previas...${NC}"
pkill -f "neon-watchdog run" 2>/dev/null
sleep 1

# Ejecutar en background
echo -e "${GREEN}✓ Iniciando Neon Watchdog...${NC}"
echo -e "${BLUE}  - Configuración: config-local.yml${NC}"
echo -e "${BLUE}  - Dashboard: http://localhost:8080${NC}"
echo -e "${BLUE}  - Usuario: admin${NC}"
echo -e "${BLUE}  - Contraseña: admin123${NC}"
echo ""

./neon-watchdog run -c config-local.yml > neon-local.log 2>&1 &
PID=$!

# Esperar un momento para verificar que arrancó
sleep 2

if ps -p $PID > /dev/null; then
    echo -e "${GREEN}✓ Neon Watchdog iniciado correctamente (PID: $PID)${NC}"
    echo -e "${BLUE}  - Logs: tail -f neon-local.log${NC}"
    echo ""
    
    # Esperar a que el servidor esté listo
    echo -e "${YELLOW}→ Esperando a que el servidor esté listo...${NC}"
    for i in {1..10}; do
        if curl -s -u admin:admin123 http://localhost:8080/api/health > /dev/null 2>&1; then
            echo -e "${GREEN}✓ API REST disponible en http://localhost:8080${NC}"
            echo ""
            
            # Mostrar ejemplos de uso
            echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
            echo -e "${BLUE}║         Ejemplos de Uso API            ║${NC}"
            echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
            echo ""
            echo -e "${GREEN}# Ver estado:${NC}"
            echo -e "curl -u admin:admin123 http://localhost:8080/api/status"
            echo ""
            echo -e "${GREEN}# Listar servicios:${NC}"
            echo -e "curl -u admin:admin123 http://localhost:8080/api/targets"
            echo ""
            echo -e "${GREEN}# Crear servicio:${NC}"
            echo -e "curl -u admin:admin123 -X POST http://localhost:8080/api/targets \\"
            echo -e "  -H 'Content-Type: application/json' \\"
            echo -e "  -d '{"
            echo -e "    \"name\": \"mi-servicio\","
            echo -e "    \"enabled\": true,"
            echo -e "    \"checks\": [{\"type\": \"tcp_port\", \"tcp_port\": \"22\"}],"
            echo -e "    \"action\": {"
            echo -e "      \"type\": \"exec\","
            echo -e "      \"exec\": {\"restart\": [\"/bin/echo\", \"test\"]}"
            echo -e "    },"
            echo -e "    \"policy\": {"
            echo -e "      \"fail_threshold\": 3,"
            echo -e "      \"restart_cooldown_seconds\": 60,"
            echo -e "      \"max_restarts_per_hour\": 10"
            echo -e "    }"
            echo -e "  }'"
            echo ""
            echo -e "${GREEN}# Eliminar servicio:${NC}"
            echo -e "curl -u admin:admin123 -X DELETE http://localhost:8080/api/targets/mi-servicio"
            echo ""
            echo -e "${YELLOW}Para detener: pkill -f 'neon-watchdog run' o kill $PID${NC}"
            echo -e "${BLUE}Documentación completa: API-REST.md${NC}"
            echo ""
            exit 0
        fi
        sleep 1
    done
    
    echo -e "${YELLOW}⚠ El servidor tardó en iniciar, revisa los logs:${NC}"
    echo -e "${YELLOW}  tail -f neon-local.log${NC}"
else
    echo -e "${RED}✗ Error al iniciar Neon Watchdog${NC}"
    echo -e "${YELLOW}Revisa los logs para más detalles:${NC}"
    tail -20 neon-local.log
    exit 1
fi
