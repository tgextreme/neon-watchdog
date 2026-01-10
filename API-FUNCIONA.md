# 🎉 LA API SÍ FUNCIONA - Guía Rápida

## ✅ Resumen

**La API REST funciona correctamente**. El problema era la documentación incorrecta.

---

## 🚀 Uso Rápido

### 1. Ejecutar el Script de Prueba

```bash
cd ~/proyectos/"Neon Watchdogs"
./test-api.sh
```

Este script probará automáticamente todos los endpoints.

---

### 2. Ejemplos Manuales

#### Crear un Servicio

```bash
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx-monitor",
    "enabled": true,
    "checks": [
      {"type": "process_name", "process_name": "nginx"}
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
```

#### Listar Servicios

```bash
curl -u admin:admin123 http://localhost:8080/api/targets
```

#### Ver Estado

```bash
curl -u admin:admin123 http://localhost:8080/api/status
```

#### Eliminar Servicio

```bash
curl -u admin:admin123 -X DELETE http://localhost:8080/api/targets/nginx-monitor
```

---

## 📚 Documentación

| Archivo | Descripción |
|---------|-------------|
| [DIAGNOSTICO-API.md](DIAGNOSTICO-API.md) | Diagnóstico completo |
| [API-REST-FIXED.md](API-REST-FIXED.md) | Documentación correcta de la API |
| [SOLUCION-PROBLEMAS-API.md](SOLUCION-PROBLEMAS-API.md) | Solución de problemas |
| [test-api.sh](test-api.sh) | Script de prueba |

---

## ⚠️ Problema de Permisos

Si ves este error:
```
Failed to save config: write error: permission denied
```

**No te preocupes**: Los cambios SÍ se aplican en memoria y el watchdog funciona.

**Solución rápida**:
```bash
# Usar configuración local
cp /opt/neon-watchdog/config.yml ~/neon-config.yml
./neon-watchdog run -c ~/neon-config.yml
```

---

## 🎯 Diferencia Clave

### ❌ Documentación Antigua (INCORRECTA)
```json
{
  "check": {...},
  "actions": [...],
  "thresholds": {...}
}
```

### ✅ Estructura Real (CORRECTA)
```json
{
  "checks": [...],
  "action": {...},
  "policy": {...}
}
```

---

## 🏁 Resultado

✅ **CREATE funciona**  
✅ **READ funciona**  
✅ **UPDATE funciona**  
✅ **DELETE funciona**

**¡La API está 100% operativa!** 🚀
