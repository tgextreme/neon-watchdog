#!/bin/bash
# Script para probar la API REST de Neon Watchdog
# Este script demuestra cómo usar correctamente la API

BASE_URL="http://localhost:8080"
USER="admin"
PASS="admin123"

echo "=== PROBANDO API REST DE NEON WATCHDOG ==="
echo ""

# 1. Ver estado actual
echo "1. Estado del sistema:"
curl -s -u $USER:$PASS $BASE_URL/api/status | head -5
echo ""
echo ""

# 2. Listar servicios
echo "2. Listar todos los servicios:"
curl -s -u $USER:$PASS $BASE_URL/api/targets
echo ""
echo ""

# 3. Crear servicio de prueba
echo "3. Crear servicio de prueba (test-nginx):"
curl -s -u $USER:$PASS -X POST $BASE_URL/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-nginx",
    "enabled": true,
    "checks": [
      {
        "type": "tcp_port",
        "tcp_port": "80"
      }
    ],
    "action": {
      "type": "systemd",
      "systemd": {
        "unit": "nginx.service",
        "method": "restart"
      }
    },
    "policy": {
      "fail_threshold": 3,
      "restart_cooldown_seconds": 60,
      "max_restarts_per_hour": 10
    }
  }'
echo ""
echo ""

# 4. Ver servicio específico
echo "4. Ver servicio creado:"
curl -s -u $USER:$PASS $BASE_URL/api/targets/test-nginx
echo ""
echo ""

# 5. Actualizar servicio
echo "5. Actualizar servicio (cambiar threshold a 5):"
curl -s -u $USER:$PASS -X PUT $BASE_URL/api/targets/test-nginx \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-nginx",
    "enabled": true,
    "checks": [
      {
        "type": "tcp_port",
        "tcp_port": "80"
      }
    ],
    "action": {
      "type": "systemd",
      "systemd": {
        "unit": "nginx.service",
        "method": "restart"
      }
    },
    "policy": {
      "fail_threshold": 5,
      "restart_cooldown_seconds": 120,
      "max_restarts_per_hour": 8
    }
  }'
echo ""
echo ""

# 6. Eliminar servicio
echo "6. Eliminar servicio de prueba:"
curl -s -u $USER:$PASS -X DELETE $BASE_URL/api/targets/test-nginx
echo ""
echo ""

echo "=== PRUEBA COMPLETADA ==="
