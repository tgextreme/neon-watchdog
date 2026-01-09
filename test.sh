#!/bin/bash
# Script de ejemplo para testing de neon-watchdog
# Este script levanta un servidor HTTP de prueba y demuestra
# cómo neon-watchdog puede detectar su caída y recuperarlo

set -e

PORT=8888
CONFIG_FILE="test-config.yml"
WATCHDOG_BIN="./bin/neon-watchdog"

echo "🧪 Neon Watchdog - Script de Testing"
echo "======================================"
echo ""

# Verificar que el binario existe
if [ ! -f "$WATCHDOG_BIN" ]; then
    echo "❌ Error: Binario no encontrado. Ejecuta 'make build' primero."
    exit 1
fi

# Crear configuración de prueba temporal
cat > "$CONFIG_FILE" <<EOF
log_level: DEBUG
timeout_seconds: 5

default_policy:
  fail_threshold: 1
  restart_cooldown_seconds: 5
  max_restarts_per_hour: 20

targets:
  - name: test-http-server
    enabled: true
    checks:
      - type: tcp_port
        tcp_port: "$PORT"
    action:
      type: exec
      exec:
        restart:
          - /usr/bin/python3
          - -m
          - http.server
          - "$PORT"
EOF

echo "✅ Configuración de prueba creada: $CONFIG_FILE"
echo ""

# Función para limpiar al salir
cleanup() {
    echo ""
    echo "🧹 Limpiando..."
    pkill -f "http.server $PORT" 2>/dev/null || true
    rm -f "$CONFIG_FILE"
    echo "✅ Limpieza completada"
}
trap cleanup EXIT

# Paso 1: Validar configuración
echo "📋 Paso 1: Validando configuración..."
$WATCHDOG_BIN test-config -c "$CONFIG_FILE"
echo ""

# Paso 2: Levantar servidor de prueba
echo "🚀 Paso 2: Levantando servidor HTTP en puerto $PORT..."
python3 -m http.server $PORT > /dev/null 2>&1 &
SERVER_PID=$!
sleep 2

if kill -0 $SERVER_PID 2>/dev/null; then
    echo "✅ Servidor corriendo (PID: $SERVER_PID)"
else
    echo "❌ Error: No se pudo levantar el servidor"
    exit 1
fi
echo ""

# Paso 3: Verificar que está saludable
echo "🔍 Paso 3: Verificando que el watchdog detecta el servidor..."
$WATCHDOG_BIN check -c "$CONFIG_FILE" --verbose
RESULT=$?
if [ $RESULT -eq 0 ]; then
    echo "✅ Servidor detectado como saludable"
else
    echo "❌ Error: Servidor no detectado"
    exit 1
fi
echo ""

# Paso 4: Simular fallo (matar el servidor)
echo "💀 Paso 4: Simulando fallo (matando el servidor)..."
kill $SERVER_PID
sleep 1
echo "✅ Servidor detenido"
echo ""

# Paso 5: Verificar que detecta el fallo
echo "🔍 Paso 5: Verificando que el watchdog detecta el fallo..."
$WATCHDOG_BIN check -c "$CONFIG_FILE" --verbose
RESULT=$?
if [ $RESULT -eq 2 ]; then
    echo "✅ Fallo detectado correctamente (exit code: 2)"
else
    echo "⚠️  Exit code inesperado: $RESULT"
fi
echo ""

# Paso 6: Verificar que intenta recuperar
echo "🔧 Paso 6: El watchdog debería haber intentado reiniciar el servidor"
echo "    Verifica en los logs anteriores si ves:"
echo "    - 'executing recovery action'"
echo "    - 'recovery action succeeded'"
echo ""

# Dar tiempo a que el servidor arranque
sleep 2

# Verificar si el servidor está corriendo de nuevo
if curl -s http://localhost:$PORT > /dev/null 2>&1; then
    echo "✅ ¡Éxito! El servidor fue recuperado automáticamente"
else
    echo "❌ El servidor no fue recuperado"
    echo "    Esto puede deberse a permisos o al método de restart"
fi
echo ""

echo "🎉 Testing completado!"
echo ""
echo "Para ver un ejemplo en modo daemon, ejecuta:"
echo "  $WATCHDOG_BIN run -c $CONFIG_FILE --verbose"
echo ""
echo "Y en otra terminal, mata el servidor para ver la recuperación automática:"
echo "  pkill -f 'http.server $PORT'"
