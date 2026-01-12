# API REST - Neon Watchdog Dashboard

Documentación completa de la API REST y Dashboard Web de Neon Watchdog.

## 📋 Descripción

La API REST permite gestionar servicios monitoreados, ver el estado del sistema y realizar operaciones sin necesidad de editar archivos YAML. Todas las operaciones guardan automáticamente en el archivo de configuración.

## 🔐 Autenticación

Todas las peticiones requieren **HTTP Basic Authentication**:

```bash
curl -u usuario:password http://localhost:8080/api/...
```

### Crear Usuarios

```bash
# Usando htpasswd
htpasswd -B -c users.txt admin

# O manualmente con bcrypt
# El archivo users.txt debe estar en el directorio de ejecución
# Formato: usuario:hash_bcrypt
```

**Archivo users.txt de ejemplo:**
```
admin:$2a$10$abcdefghijklmnopqrstuvwxyz1234567890
user2:$2a$10$zyxwvutsrqponmlkjihgfedcba0987654321
```

---

## 🌐 Habilitar Dashboard

Añade a tu `config.yml`:

```yaml
dashboard:
  enabled: true
  port: 8080
  path: "/"
```

Reinicia el servicio:

```bash
sudo systemctl restart neon-watchdog.timer
# o
sudo systemctl restart neon-watchdog-daemon.service
```

Acceder:
- Dashboard Web: `http://localhost:8080/`
- API REST: `http://localhost:8080/api/`

---

## 📡 Endpoints de la API

### 1. Ver Estado del Sistema

**GET** `/api/status`

Retorna el estado completo del watchdog con todos los servicios monitoreados.

**Ejemplo:**
```bash
curl -u admin:password http://localhost:8080/api/status
```

**Respuesta:**
```json
{
  "uptime": 3600000000000,
  "start_time": "2026-01-09T20:00:00Z",
  "targets": {
    "nginx": {
      "name": "nginx",
      "healthy": true,
      "enabled": true,
      "last_check": "2026-01-09T21:00:00Z",
      "consecutive_failures": 0,
      "total_restarts": 2,
      "last_restart": "2026-01-09T20:30:00Z",
      "message": "Service is running"
    },
    "apache": {
      "name": "apache",
      "healthy": false,
      "enabled": true,
      "last_check": "2026-01-09T21:00:05Z",
      "consecutive_failures": 3,
      "total_restarts": 5,
      "last_restart": "2026-01-09T20:58:00Z",
      "message": "Connection refused"
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
curl -u admin:password http://localhost:8080/api/health
```

**Respuesta (200 OK) - Todos saludables:**
```json
{
  "status": "healthy",
  "targets": 3
}
```

**Respuesta (503 Service Unavailable) - Alguno no saludable:**
```json
{
  "status": "unhealthy",
  "targets": 3,
  "unhealthy_count": 1
}
```

---

### 3. Listar Todos los Servicios

**GET** `/api/targets`

Obtiene la lista completa de servicios configurados.

**Ejemplo:**
```bash
curl -u admin:password http://localhost:8080/api/targets
```

**Respuesta:**
```json
{
  "targets": [
    {
      "name": "nginx",
      "enabled": true,
      "depends_on": [],
      "checks": [
        {
          "type": "process_name",
          "process_name": "nginx"
        },
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
        "fail_threshold": 1,
        "restart_cooldown_seconds": 60,
        "max_restarts_per_hour": 10
      }
    }
  ]
}
```

---

### 4. Obtener Servicio Específico

**GET** `/api/targets/{name}`

Obtiene la configuración completa de un servicio específico.

**Ejemplo:**
```bash
curl -u admin:password http://localhost:8080/api/targets/nginx
```

**Respuesta:**
```json
{
  "name": "nginx",
  "enabled": true,
  "depends_on": [],
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
    "fail_threshold": 1,
    "restart_cooldown_seconds": 60,
    "max_restarts_per_hour": 10
  }
}
```

**Errores:**
- `404 Not Found` - El servicio no existe

---

### 5. Crear Nuevo Servicio

**POST** `/api/targets`

Añade un nuevo servicio al watchdog. Se guarda automáticamente en el YAML.

**Content-Type:** `application/json`

#### Ejemplo 1: Monitorear Servicio Systemd

```bash
curl -u admin:password -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx",
    "enabled": true,
    "checks": [
      {
        "type": "process_name",
        "process_name": "nginx"
      },
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
```

#### Ejemplo 2: Monitorear Puerto TCP

```bash
curl -u admin:password -X POST http://localhost:8080/api/targets \
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

#### Ejemplo 3: Monitorear Health Check HTTP

```bash
curl -u admin:password -X POST http://localhost:8080/api/targets \
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
      "type": "systemd",
      "systemd": {
        "unit": "api-backend.service",
        "method": "restart"
      }
    },
    "policy": {
      "fail_threshold": 2,
      "restart_cooldown_seconds": 60,
      "max_restarts_per_hour": 8
    }
  }'
```

#### Ejemplo 4: Comando Personalizado

```bash
curl -u admin:password -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "custom-app",
    "enabled": true,
    "checks": [
      {
        "type": "command",
        "command": ["/usr/local/bin/check-app.sh"]
      }
    ],
    "action": {
      "type": "exec",
      "exec": {
        "restart": ["/opt/app/restart.sh"]
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
curl -u admin:password -X PUT http://localhost:8080/api/targets/nginx \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx",
    "enabled": true,
    "checks": [
      {
        "type": "process_name",
        "process_name": "nginx"
      },
      {
        "type": "tcp_port",
        "tcp_port": "80"
      },
      {
        "type": "tcp_port",
        "tcp_port": "443"
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
```

**Respuesta (200 OK):**
```json
{
  "message": "target updated successfully",
  "name": "nginx"
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
curl -u admin:password -X DELETE http://localhost:8080/api/targets/nginx
```

**Respuesta (200 OK):**
```json
{
  "message": "target deleted successfully",
  "name": "nginx"
}
```

**Errores:**
- `404 Not Found` - El servicio no existe

---

### 8. Habilitar/Deshabilitar Servicio

**PUT** `/api/targets/{name}/toggle`

Activa o desactiva el monitoreo de un servicio sin eliminarlo.

**Ejemplo:**
```bash
curl -u admin:password -X PUT http://localhost:8080/api/targets/nginx/toggle
```

**Respuesta (200 OK):**
```json
{
  "message": "target toggled successfully",
  "name": "nginx",
  "enabled": false
}
```

---

### 9. Actualizar Solo Estado (Enable/Disable)

**PATCH** `/api/targets/{name}`

Actualiza solo el estado enabled de un servicio.

**Content-Type:** `application/json`

**Ejemplo:**
```bash
curl -u admin:password -X PATCH http://localhost:8080/api/targets/nginx \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": false
  }'
```

**Respuesta (200 OK):**
```json
{
  "message": "target state updated",
  "name": "nginx",
  "enabled": false
}
```

---

### 10. Ver Configuración Completa

**GET** `/api/config`

Obtiene la configuración completa del watchdog (incluyendo global y targets).

**Ejemplo:**
```bash
curl -u admin:password http://localhost:8080/api/config
```

**Respuesta:**
```json
{
  "interval_seconds": 30,
  "timeout_seconds": 10,
  "log_level": "INFO",
  "state_file": "/var/lib/neon-watchdog/state.json",
  "default_policy": {
    "fail_threshold": 1,
    "restart_cooldown_seconds": 60,
    "max_restarts_per_hour": 10
  },
  "targets": [...],
  "notifications": [...],
  "metrics": {...},
  "dashboard": {...}
}
```

---

### 11. Actualizar Configuración Global

**PUT** `/api/config`

Actualiza la configuración global (no incluye targets).

**Content-Type:** `application/json`

**Ejemplo:**
```bash
curl -u admin:password -X PUT http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{
    "log_level": "DEBUG",
    "timeout_seconds": 15,
    "default_policy": {
      "fail_threshold": 2,
      "restart_cooldown_seconds": 90,
      "max_restarts_per_hour": 8
    }
  }'
```

**Respuesta (200 OK):**
```json
{
  "message": "configuration updated successfully"
}
```

---

### 12. Validar Configuración

**POST** `/api/config/validate`

Valida la configuración actual sin guardarla.

**Content-Type:** `application/json`

**Ejemplo:**
```bash
curl -u admin:password -X POST http://localhost:8080/api/config/validate \
  -H "Content-Type: application/json" \
  -d '{
    "targets": [
      {
        "name": "test",
        "enabled": true,
        "checks": [{"type": "tcp_port", "tcp_port": "80"}],
        "action": {"type": "systemd", "systemd": {"unit": "test.service", "method": "restart"}}
      }
    ]
  }'
```

**Respuesta (200 OK):**
```json
{
  "valid": true,
  "message": "configuration is valid"
}
```

**Respuesta (400 Bad Request):**
```json
{
  "valid": false,
  "error": "target 'test': check type 'invalid_type' is not supported"
}
```

---

## 🖥️ Dashboard Web

### Interfaz Web

El dashboard web proporciona una interfaz visual para:

- Ver estado de todos los servicios en tiempo real
- Ver historial de checks y restarts
- Habilitar/deshabilitar servicios con un click
- Añadir nuevos servicios mediante formulario
- Editar configuración de servicios existentes
- Ver logs y métricas

### Acceso

```
http://localhost:8080/
```

### Características

✅ **Vista de Estado en Tiempo Real**
- Indicadores visuales (verde/rojo) para cada servicio
- Tiempo desde último check
- Número de fallos consecutivos
- Total de restarts

✅ **Gestión de Servicios**
- Añadir/editar/eliminar servicios
- Habilitar/deshabilitar con toggle
- Validación de formularios

✅ **Historial**
- Últimos eventos de cada servicio
- Timeline de restarts
- Logs filtrados por servicio

✅ **Configuración**
- Editor YAML integrado
- Validación en tiempo real
- Vista previa de cambios

---

## 🔒 Seguridad

### HTTPS (Recomendado para Producción)

Usa un reverse proxy como nginx:

```nginx
server {
    listen 443 ssl;
    server_name watchdog.example.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Authorization $http_authorization;
    }
}
```

### Restricciones de IP

```nginx
location / {
    allow 192.168.1.0/24;
    allow 10.0.0.0/8;
    deny all;
    
    proxy_pass http://localhost:8080;
}
```

### Rate Limiting

```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

location /api/ {
    limit_req zone=api burst=20 nodelay;
    proxy_pass http://localhost:8080;
}
```

---

## 🧪 Ejemplos de Integración

### Script de Monitoreo

```bash
#!/bin/bash
# check-services.sh - Monitorear servicios via API

API_URL="http://localhost:8080"
USER="admin"
PASS="password"

# Obtener estado
status=$(curl -s -u $USER:$PASS "$API_URL/api/health")
health=$(echo $status | jq -r '.status')

if [ "$health" != "healthy" ]; then
    echo "ALERT: System unhealthy!"
    curl -s -u $USER:$PASS "$API_URL/api/status" | jq '.targets'
    
    # Enviar alerta
    # ...
fi
```

### Python Client

```python
import requests
from requests.auth import HTTPBasicAuth

class NeonWatchdogClient:
    def __init__(self, base_url, username, password):
        self.base_url = base_url
        self.auth = HTTPBasicAuth(username, password)
    
    def get_status(self):
        r = requests.get(f"{self.base_url}/api/status", auth=self.auth)
        return r.json()
    
    def add_target(self, target_config):
        r = requests.post(
            f"{self.base_url}/api/targets",
            json=target_config,
            auth=self.auth
        )
        return r.json()
    
    def delete_target(self, name):
        r = requests.delete(
            f"{self.base_url}/api/targets/{name}",
            auth=self.auth
        )
        return r.json()

# Uso
client = NeonWatchdogClient("http://localhost:8080", "admin", "password")
status = client.get_status()
print(f"Uptime: {status['uptime']}")

for name, target in status['targets'].items():
    print(f"{name}: {'✓' if target['healthy'] else '✗'}")
```

### Terraform Provider (Ejemplo)

```hcl
resource "neon_watchdog_target" "nginx" {
  name    = "nginx"
  enabled = true
  
  checks = [
    {
      type         = "process_name"
      process_name = "nginx"
    },
    {
      type     = "tcp_port"
      tcp_port = "80"
    }
  ]
  
  action = {
    type = "systemd"
    systemd = {
      unit   = "nginx.service"
      method = "restart"
    }
  }
  
  policy = {
    fail_threshold            = 3
    restart_cooldown_seconds  = 60
    max_restarts_per_hour     = 10
  }
}
```

---

## 📊 Códigos de Estado HTTP

- **200 OK** - Operación exitosa
- **201 Created** - Recurso creado
- **400 Bad Request** - Datos inválidos o JSON malformado
- **401 Unauthorized** - Autenticación fallida
- **404 Not Found** - Recurso no encontrado
- **409 Conflict** - Conflicto (ej: target ya existe)
- **500 Internal Server Error** - Error del servidor
- **503 Service Unavailable** - Sistema no saludable (en /health)

---

## 🔍 Troubleshooting API

### Dashboard no accesible

```bash
# Verificar que está habilitado en config
grep -A3 "dashboard:" /etc/neon-watchdog/config.yml

# Verificar puerto abierto
netstat -tlnp | grep :8080

# Verificar logs
journalctl -u neon-watchdog.service | grep dashboard
```

### Error 401 Unauthorized

```bash
# Verificar usuarios en users.txt
cat users.txt

# Crear usuario nuevo
htpasswd -B users.txt username

# Test con curl
curl -v -u username:password http://localhost:8080/api/status
```

### Cambios no se guardan

```bash
# Verificar permisos del archivo config
ls -l /etc/neon-watchdog/config.yml

# Ver logs de escritura
journalctl -u neon-watchdog.service | grep "config saved"
```

---

## 📚 Recursos Adicionales

- **[README.md](README.md)** - Documentación general
- **[INSTALL.md](INSTALL.md)** - Guía de instalación
- **Ejemplos**: Ver carpeta `examples/` en el repositorio

---

**Desarrollado con ❤️ para mantener tus servicios siempre activos**
