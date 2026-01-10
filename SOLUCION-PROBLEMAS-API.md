# 🔧 SOLUCIÓN DE PROBLEMAS - API REST

## ❌ Problema: La API no funciona

### Diagnóstico Realizado

La API REST de Neon Watchdog **SÍ FUNCIONA**, pero había dos problemas:

### 1. ❌ Documentación Incorrecta

**Problema**: El archivo `API-REST.md` contenía una estructura JSON incorrecta que no coincide con el código real.

**Estructura INCORRECTA (documentación antigua)**:
```json
{
  "check": {"type": "systemd", ...},
  "actions": [...],
  "thresholds": {...}
}
```

**Estructura CORRECTA (código real)**:
```json
{
  "checks": [...],
  "action": {...},
  "policy": {...}
}
```

**Solución**: Se creó `API-REST-FIXED.md` con la documentación correcta.

---

### 2. ⚠️ Problema de Permisos

**Síntoma**:
```
Failed to save config: write error: open /opt/neon-watchdog/config.yml: permission denied
```

**Causa**: El archivo de configuración es propiedad de `www-data` pero el proceso corre como otro usuario.

**¿Afecta la funcionalidad?**: NO
- Los targets **SÍ se crean** en memoria
- El watchdog **SÍ los monitorea**
- Solo falla la persistencia al disco

**Soluciones**:

#### Opción A: Cambiar permisos del archivo (temporal)
```bash
sudo chmod 666 /opt/neon-watchdog/config.yml
```

#### Opción B: Ejecutar como www-data
```bash
sudo -u www-data /opt/neon-watchdog/neon-watchdog run -c /opt/neon-watchdog/config.yml
```

#### Opción C: Usar archivo local (desarrollo)
```bash
cp /opt/neon-watchdog/config.yml ~/neon-config.yml
./neon-watchdog run -c ~/neon-config.yml
```

---

## ✅ Prueba de que la API Funciona

### Comandos de Prueba Exitosos

```bash
# 1. Ver estado (FUNCIONA ✅)
curl -u admin:admin123 http://localhost:8080/api/status

# 2. Listar servicios (FUNCIONA ✅)
curl -u admin:admin123 http://localhost:8080/api/targets

# 3. Crear servicio (FUNCIONA ✅ - se crea en memoria)
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-ssh",
    "enabled": true,
    "checks": [{"type": "tcp_port", "tcp_port": "22"}],
    "action": {
      "type": "exec",
      "exec": {"restart": ["/bin/echo", "Test"]}
    },
    "policy": {
      "fail_threshold": 3,
      "restart_cooldown_seconds": 60,
      "max_restarts_per_hour": 5
    }
  }'

# 4. Ver servicio creado (FUNCIONA ✅)
curl -u admin:admin123 http://localhost:8080/api/targets/test-ssh

# 5. Actualizar servicio (FUNCIONA ✅)
curl -u admin:admin123 -X PUT http://localhost:8080/api/targets/test-ssh \
  -H "Content-Type: application/json" \
  -d '{ ... estructura completa ... }'

# 6. Eliminar servicio (FUNCIONA ✅)
curl -u admin:admin123 -X DELETE http://localhost:8080/api/targets/test-ssh
```

### Resultado de Pruebas

```json
// GET /api/targets devuelve correctamente:
[
  {
    "name": "test-monitoring",
    "enabled": true,
    "checks": [{"type": "tcp_port", "tcp_port": "22"}],
    "action": {"type": "exec", "exec": {"restart": ["/bin/echo", "SSH está caído"]}},
    "policy": {"fail_threshold": 3, "restart_cooldown_seconds": 120, "max_restarts_per_hour": 5}
  },
  {
    "name": "test-ssh",
    "enabled": true,
    "checks": [{"type": "tcp_port", "tcp_port": "22"}],
    "action": {"type": "exec", "exec": {"restart": ["/bin/echo", "SSH is down"]}},
    "policy": {"fail_threshold": 3, "restart_cooldown_seconds": 60, "max_restarts_per_hour": 5}
  }
]
```

✅ **Confirmado**: El servicio `test-ssh` se creó exitosamente y aparece en la lista.

---

## 🎯 Resumen

### ✅ Lo que SÍ funciona:
- ✅ Autenticación HTTP Basic
- ✅ GET /api/status
- ✅ GET /api/health
- ✅ GET /api/targets
- ✅ GET /api/targets/{name}
- ✅ POST /api/targets (crea en memoria)
- ✅ PUT /api/targets/{name} (actualiza en memoria)
- ✅ DELETE /api/targets/{name} (elimina de memoria)
- ✅ GET /api/config

### ⚠️ Lo que necesita permisos:
- ⚠️ Guardar configuración en disco (necesita permisos de escritura)

### ❌ Lo que estaba mal:
- ❌ Documentación con estructura JSON incorrecta

---

## 📝 Archivos Corregidos

1. **`API-REST-FIXED.md`** - Documentación correcta de la API
2. **`test-api.sh`** - Script para probar todos los endpoints
3. **`SOLUCION-PROBLEMAS-API.md`** - Este archivo

---

## 🚀 Cómo Usar Ahora

### Desarrollo Local (Recomendado)

```bash
# 1. Copiar configuración a tu home
cp /opt/neon-watchdog/config.yml ~/neon-config.yml

# 2. Ejecutar watchdog con tu configuración
cd ~/proyectos/"Neon Watchdogs"
./neon-watchdog run -c ~/neon-config.yml

# 3. Usar la API normalmente
curl -u admin:admin123 http://localhost:8080/api/targets
```

### Producción con systemd

```bash
# 1. Asegurar permisos correctos
sudo chown www-data:www-data /opt/neon-watchdog/config.yml
sudo chmod 644 /opt/neon-watchdog/config.yml

# 2. El servicio systemd ya corre como www-data
sudo systemctl restart neon-watchdog

# 3. La API podrá guardar cambios en disco
```

---

## 📚 Referencias

- **Documentación correcta**: [API-REST-FIXED.md](API-REST-FIXED.md)
- **Script de prueba**: [test-api.sh](test-api.sh)
- **Código fuente**: [internal/dashboard/dashboard.go](internal/dashboard/dashboard.go)
- **Estructuras**: [internal/config/config.go](internal/config/config.go)
