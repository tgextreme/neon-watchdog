# 🚀 API REST - Neon Watchdog Dashboard (CORREGIDA)

## 📋 Descripción

API REST completa para gestionar servicios monitoreados **sin editar archivos YAML**. Todas las operaciones guardan automáticamente en el archivo de configuración.

## 🔐 Autenticación

Todas las peticiones requieren **HTTP Basic Authentication**:

```bash
curl -u admin:admin123 http://localhost:8080/api/...
```

**Usuarios disponibles**: Ver [USUARIOS-LOGIN.txt](USUARIOS-LOGIN.txt)

---

## ⚠️ IMPORTANTE: Estructura Correcta de los Targets

La API usa la siguiente estructura (**diferente a la documentación anterior**):

```json
{
  "name": "nombre-servicio",
  "enabled": true,
  "checks": [
    {
      "type": "process_name | tcp_port | http | command | script",
      "process_name": "nombre-proceso",
      "tcp_port": "puerto",
      "http": {...},
      "command": [...],
      "script": {...}
    }
  ],
  "action": {
    "type": "exec | systemd",
    "exec": {
      "restart": ["/comando", "arg1", "arg2"]
    },
    "systemd": {
      "unit": "nombre.service",
      "method": "restart | start | stop"
    }
  },
  "policy": {
    "fail_threshold": 3,
    "restart_cooldown_seconds": 60,
    "max_restarts_per_hour": 10
  }
}
```

---

## 📡 Endpoints Disponibles

### 1. Ver Estado del Sistema

**GET** `/api/status`

Retorna el estado completo del watchdog con todos los servicios monitoreados.

**Ejemplo:**
```bash
curl -u admin:admin123 http://localhost:8080/api/status
```

**Respuesta:**
```json
{
  "uptime": 3600000000000,
  "start_time": "2026-01-09T20:00:00Z",
  "targets": {
    "apache": {
      "name": "apache",
      "healthy": true,
      "enabled": true,
      "last_check": "2026-01-09T21:00:00Z",
      "consecutive_failures": 0,
      "total_restarts": 2,
      "last_restart": "2026-01-09T20:30:00Z",
      "message": "Service is running"
    }
  }
}
```

---

### 2. Health Check Simple

**GET** `/api/health`

Endpoint rápido para verificar si todos los servicios están saludables.

**Ejemplo:**
```bash
curl -u admin:admin123 http://localhost:8080/api/health
```

**Respuesta (200 OK):**
```json
{
  "status": "healthy",
  "targets": 3
}
```

**Respuesta (503 Service Unavailable):**
```json
{
  "status": "unhealthy",
  "targets": 3
}
```

---

### 3. Listar Todos los Servicios

**GET** `/api/targets`

Obtiene la lista completa de servicios configurados.

**Ejemplo:**
```bash
curl -u admin:admin123 http://localhost:8080/api/targets
```

**Respuesta:**
```json
[
  {
    "name": "apache",
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
        "unit": "apache2.service",
        "method": "restart"
      }
    },
    "policy": {
      "fail_threshold": 3,
      "restart_cooldown_seconds": 60,
      "max_restarts_per_hour": 10
    }
  }
]
```

---

### 4. Obtener Servicio Específico

**GET** `/api/targets/{name}`

Obtiene la configuración completa de un servicio específico.

**Ejemplo:**
```bash
curl -u admin:admin123 http://localhost:8080/api/targets/apache
```

---

### 5. Crear Nuevo Servicio

**POST** `/api/targets`

Añade un nuevo servicio al watchdog. Se guarda automáticamente en el YAML.

**Content-Type:** `application/json`

#### Ejemplo 1: Monitorear Servicio Systemd

```bash
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx",
    "enabled": true,
    "checks": [
      {
        "type": "process_name",
        "process_name": "nginx"
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
```

#### Ejemplo 2: Monitorear Puerto TCP

```bash
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "database",
    "enabled": true,
    "checks": [
      {
        "type": "tcp_port",
        "tcp_port": "3306"
      }
    ],
    "action": {
      "type": "systemd",
      "systemd": {
        "unit": "mysql.service",
        "method": "restart"
      }
    },
    "policy": {
      "fail_threshold": 5,
      "restart_cooldown_seconds": 120,
      "max_restarts_per_hour": 5
    }
  }'
```

#### Ejemplo 3: Monitorear URL HTTP

```bash
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "api-backend",
    "enabled": true,
    "checks": [
      {
        "type": "http",
        "http": {
          "url": "http://localhost:8000/health",
          "method": "GET",
          "expected_status": 200,
          "timeout_seconds": 10
        }
      }
    ],
    "action": {
      "type": "exec",
      "exec": {
        "restart": ["systemctl", "restart", "api-backend.service"]
      }
    },
    "policy": {
      "fail_threshold": 2,
      "restart_cooldown_seconds": 60,
      "max_restarts_per_hour": 8
    }
  }'
```

#### Ejemplo 4: Ejecutar Comando Personalizado

```bash
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "custom-app",
    "enabled": true,
    "checks": [
      {
        "type": "command",
        "command": ["pgrep", "-f", "my-app"]
      }
    ],
    "action": {
      "type": "exec",
      "exec": {
        "restart": ["/opt/app/start.sh"]
      }
    },
    "policy": {
      "fail_threshold": 3,
      "restart_cooldown_seconds": 60,
      "max_restarts_per_hour": 10
    }
  }'
```

**Respuesta (201 Created):**
```json
{
  "name": "nginx",
  "enabled": true,
  "checks": [...],
  "action": {...},
  "policy": {...}
}
```

**Errores:**
- `400 Bad Request` - JSON inválido o campos requeridos faltantes
- `409 Conflict` - Ya existe un servicio con ese nombre
- `401 Unauthorized` - Credenciales incorrectas

---

### 6. Actualizar Servicio

**PUT** `/api/targets/{name}`

Actualiza la configuración completa de un servicio existente.

**Content-Type:** `application/json`

**Ejemplo:**
```bash
curl -u admin:admin123 -X PUT http://localhost:8080/api/targets/apache \
  -H "Content-Type: application/json" \
  -d '{
    "name": "apache",
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
        "unit": "apache2.service",
        "method": "restart"
      }
    },
    "policy": {
      "fail_threshold": 5,
      "restart_cooldown_seconds": 120,
      "max_restarts_per_hour": 8
    }
  }'
```

**Respuesta (200 OK):**
```json
{
  "name": "apache",
  "enabled": true,
  ...
}
```

**Errores:**
- `404 Not Found` - El servicio no existe
- `400 Bad Request` - JSON inválido

---

### 7. Eliminar Servicio

**DELETE** `/api/targets/{name}`

Elimina un servicio del watchdog.

**Ejemplo:**
```bash
curl -u admin:admin123 -X DELETE http://localhost:8080/api/targets/apache
```

**Respuesta (204 No Content)**

**Errores:**
- `404 Not Found` - El servicio no existe

---

### 8. Obtener Configuración Completa

**GET** `/api/config`

Descarga la configuración completa del watchdog en formato JSON.

**Ejemplo:**
```bash
curl -u admin:admin123 http://localhost:8080/api/config
```

---

## 📦 Tipos de Checks Disponibles

### 1. process_name
Verifica si un proceso está corriendo por nombre.

```json
{
  "type": "process_name",
  "process_name": "nginx"
}
```

### 2. tcp_port
Verifica si un puerto TCP está escuchando.

```json
{
  "type": "tcp_port",
  "tcp_port": "3306"
}
```

### 3. http
Realiza un health check HTTP.

```json
{
  "type": "http",
  "http": {
    "url": "http://localhost:8000/health",
    "method": "GET",
    "expected_status": 200,
    "timeout_seconds": 10
  }
}
```

### 4. command
Ejecuta un comando y verifica su código de salida.

```json
{
  "type": "command",
  "command": ["pgrep", "-f", "myapp"]
}
```

### 5. script
Ejecuta un script personalizado.

```json
{
  "type": "script",
  "script": {
    "path": "/opt/scripts/check-health.sh",
    "args": ["--verbose"],
    "success_exit_codes": [0],
    "warning_exit_codes": [1]
  }
}
```

---

## 📦 Tipos de Actions Disponibles

### 1. systemd
Gestiona servicios systemd.

```json
{
  "type": "systemd",
  "systemd": {
    "unit": "nginx.service",
    "method": "restart"
  }
}
```

### 2. exec
Ejecuta comandos arbitrarios.

```json
{
  "type": "exec",
  "exec": {
    "restart": ["/opt/app/restart.sh", "--force"],
    "start": ["/opt/app/start.sh"],
    "stop": ["/opt/app/stop.sh"]
  }
}
```

---

## 🔧 Ejemplos Prácticos Completos

### Probar la API Rápidamente

```bash
# Ver estado actual
curl -u admin:admin123 http://localhost:8080/api/status

# Listar todos los servicios
curl -u admin:admin123 http://localhost:8080/api/targets

# Crear servicio de prueba
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-ssh",
    "enabled": true,
    "checks": [{"type": "tcp_port", "tcp_port": "22"}],
    "action": {
      "type": "exec",
      "exec": {"restart": ["/bin/echo", "SSH is down"]}
    },
    "policy": {
      "fail_threshold": 3,
      "restart_cooldown_seconds": 60,
      "max_restarts_per_hour": 5
    }
  }'

# Ver el servicio creado
curl -u admin:admin123 http://localhost:8080/api/targets/test-ssh

# Eliminar el servicio
curl -u admin:admin123 -X DELETE http://localhost:8080/api/targets/test-ssh
```

---

## 🐛 Códigos de Error

| Código | Descripción |
|--------|-------------|
| 200 | OK - Operación exitosa |
| 201 | Created - Recurso creado |
| 204 | No Content - Recurso eliminado |
| 400 | Bad Request - JSON inválido o campos faltantes |
| 401 | Unauthorized - Credenciales incorrectas |
| 404 | Not Found - Servicio no existe |
| 409 | Conflict - Ya existe (creación duplicada) |
| 500 | Internal Server Error - Error del servidor |
| 503 | Service Unavailable - Configuración no disponible |

---

## 📚 Recursos Adicionales

- **Dashboard Web**: http://localhost:8080 (interfaz visual)
- **Usuarios**: Ver [USUARIOS-LOGIN.txt](USUARIOS-LOGIN.txt)
- **Logs**: `tail -f neon-watchdog.log`

---

## 🎯 Diferencias con la Documentación Anterior

### ❌ INCORRECTO (documentación antigua):
```json
{
  "check": {"type": "systemd", "systemd": {...}},
  "actions": [{"type": "systemd_restart", ...}],
  "thresholds": {...}
}
```

### ✅ CORRECTO (estructura real):
```json
{
  "checks": [{"type": "tcp_port", "tcp_port": "80"}],
  "action": {"type": "systemd", "systemd": {...}},
  "policy": {...}
}
```

**Cambios principales:**
1. `check` → `checks` (array)
2. `actions` → `action` (singular)
3. `thresholds` → `policy`
4. Tipos de checks diferentes (sin prefijo "systemd_")
5. Estructura de action diferente
