#!/bin/bash
# Prueba rápida de la API REST

echo "🧪 Prueba Rápida de API REST"
echo "============================"
echo ""

# Verificar que el servidor está corriendo
if ! curl -s -u admin:admin123 http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "❌ Error: El servidor no está corriendo"
    echo "Ejecuta primero: ./run-local.sh"
    exit 1
fi

echo "✅ Servidor activo"
echo ""

# 1. Listar servicios actuales
echo "📋 1. Servicios actuales:"
curl -s -u admin:admin123 http://localhost:8080/api/targets | grep -o '"name":"[^"]*"'
echo ""

# 2. Crear servicio de prueba
echo "➕ 2. Creando servicio 'test-api'..."
RESPONSE=$(curl -s -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-api",
    "enabled": true,
    "checks": [{"type": "tcp_port", "tcp_port": "22"}],
    "action": {"type": "exec", "exec": {"restart": ["/bin/echo", "Test OK"]}},
    "policy": {"fail_threshold": 3, "restart_cooldown_seconds": 60, "max_restarts_per_hour": 5}
  }')

if echo "$RESPONSE" | grep -q "test-api"; then
    echo "✅ Servicio creado exitosamente"
else
    echo "❌ Error al crear servicio"
    echo "$RESPONSE"
fi
echo ""

# 3. Verificar que se creó
echo "🔍 3. Verificando creación..."
curl -s -u admin:admin123 http://localhost:8080/api/targets | grep -o '"name":"[^"]*"'
echo ""

# 4. Actualizar servicio
echo "✏️  4. Actualizando servicio..."
RESPONSE=$(curl -s -u admin:admin123 -X PUT http://localhost:8080/api/targets/test-api \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-api",
    "enabled": false,
    "checks": [{"type": "tcp_port", "tcp_port": "22"}],
    "action": {"type": "exec", "exec": {"restart": ["/bin/echo", "Test Updated"]}},
    "policy": {"fail_threshold": 5, "restart_cooldown_seconds": 120, "max_restarts_per_hour": 3}
  }')

if echo "$RESPONSE" | grep -q "test-api"; then
    echo "✅ Servicio actualizado exitosamente"
else
    echo "❌ Error al actualizar servicio"
fi
echo ""

# 5. Eliminar servicio
echo "🗑️  5. Eliminando servicio..."
curl -s -u admin:admin123 -X DELETE http://localhost:8080/api/targets/test-api
echo "✅ Servicio eliminado"
echo ""

# 6. Verificar eliminación
echo "🔍 6. Verificando eliminación..."
curl -s -u admin:admin123 http://localhost:8080/api/targets | grep -o '"name":"[^"]*"'
echo ""

echo "================================"
echo "✅ Todas las pruebas completadas"
echo ""
echo "📚 Documentación: API-REST.md"
echo "🚀 Guía rápida: INICIO-RAPIDO.md"
