# 🚀 API REST - Neon Watchdog Dashboard

## 📋 Descripción

API REST completa para gestionar servicios monitoreados **sin editar archivos YAML**. Todas las operaciones guardan automáticamente en el archivo de configuración.

## 🔐 Autenticación

Todas las peticiones requieren **HTTP Basic Authentication**:

```bash
curl -u admin:admin123 http://localhost:8080/api/...
```

**Usuarios disponibles**: Ver [USUARIOS-LOGIN.txt](USUARIOS-LOGIN.txt)

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

Obtiene la lista completa de servicios configurados con toda su configuración.

**Ejemplo:**
```bash
curl -u admin:admin123 http://localhost:8080/api/targets
```

**Respuesta:**
```json
{
  "targets": [
    {
      "name": "apache",
      "enabled": true,
      "check": {
        "type": "systemd",
        "systemd": {
          "service": "apache2.service"
        }
      },
      "actions": [
        {
          "type": "systemd_restart",
          "systemd_restart": {
            "service": "apache2.service"
          }
        }
      ],
      "thresholds": {
        "max_failures": 3,
        "check_interval_seconds": 30,
        "restart_delay_seconds": 5
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
curl -u admin:admin123 http://localhost:8080/api/targets/apache
```

**Respuesta:**
```json
{
  "name": "apache",
  "enabled": true,
  "check": {
    "type": "systemd",
    "systemd": {
      "service": "apache2.service"
    }
  },
  "actions": [
    {
      "type": "systemd_restart",
      "systemd_restart": {
        "service": "apache2.service"
      }
    }
  ],
  "thresholds": {
    "max_failures": 3,
    "check_interval_seconds": 30,
    "restart_delay_seconds": 5
  }
}
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
    "check": {
      "type": "systemd",
      "systemd": {
        "service": "nginx.service"
      }
    },
    "actions": [
      {
        "type": "systemd_restart",
        "systemd_restart": {
          "service": "nginx.service"
        }
      }
    ],
    "thresholds": {
      "max_failures": 3,
      "check_interval_seconds": 30,
      "restart_delay_seconds": 5
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
    "check": {
      "type": "tcp_port",
      "tcp_port": {
        "host": "localhost",
        "port": 3306,
        "timeout_seconds": 5
      }
    },
    "actions": [
      {
        "type": "systemd_restart",
        "systemd_restart": {
          "service": "mysql.service"
        }
      }
    ],
    "thresholds": {
      "max_failures": 5,
      "check_interval_seconds": 60,
      "restart_delay_seconds": 10
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
    "check": {
      "type": "http",
      "http": {
        "url": "http://localhost:8000/health",
        "method": "GET",
        "expected_status": 200,
        "timeout_seconds": 10
      }
    },
    "actions": [
      {
        "type": "command",
        "command": {
          "cmd": "systemctl restart api-backend.service"
        }
      }
    ],
    "thresholds": {
      "max_failures": 2,
      "check_interval_seconds": 30,
      "restart_delay_seconds": 3
    }
  }'
```

#### Ejemplo 4: Monitorear Proceso

```bash
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "python-app",
    "enabled": true,
    "check": {
      "type": "process",
      "process": {
        "name": "python3",
        "cmdline_contains": "app.py"
      }
    },
    "actions": [
      {
        "type": "command",
        "command": {
          "cmd": "/opt/app/start.sh"
        }
      }
    ],
    "thresholds": {
      "max_failures": 3,
      "check_interval_seconds": 30,
      "restart_delay_seconds": 5
    }
  }'
```

**Respuesta (201 Created):**
```json
{
  "message": "target created successfully",
  "name": "nginx"
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
    "check": {
      "type": "systemd",
      "systemd": {
        "service": "apache2.service"
      }
    },
    "actions": [
      {
        "type": "systemd_restart",
        "systemd_restart": {
          "service": "apache2.service"
        }
      }
    ],
    "thresholds": {
      "max_failures": 5,
      "check_interval_seconds": 60,
      "restart_delay_seconds": 10
    }
  }'
```

**Respuesta (200 OK):**
```json
{
  "message": "target updated successfully",
  "name": "apache"
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

**Respuesta (200 OK):**
```json
{
  "message": "target deleted successfully",
  "name": "apache"
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
curl -u admin:admin123 -X PUT http://localhost:8080/api/targets/apache/toggle
```

**Respuesta (200 OK):**
```json
{
  "message": "target toggled successfully",
  "name": "apache",
  "enabled": false
}
```

---

### 9. Obtener Configuración Completa

**GET** `/api/config`

Descarga la configuración completa del watchdog en formato YAML.

**Ejemplo:**
```bash
curl -u admin:admin123 http://localhost:8080/api/config
```

**Respuesta (Content-Type: text/yaml):**
```yaml
dashboard:
  enabled: true
  port: 8080
  path: /

targets:
  - name: apache
    enabled: true
    check:
      type: systemd
      systemd:
        service: apache2.service
    actions:
      - type: systemd_restart
        systemd_restart:
          service: apache2.service
    thresholds:
      max_failures: 3
      check_interval_seconds: 30
      restart_delay_seconds: 5
```

---

## 📦 Estructura de Datos

### Target (Servicio)

```json
{
  "name": "string",           // Nombre único del servicio
  "enabled": boolean,         // Si está activo el monitoreo
  "check": {                  // Configuración del check
    "type": "string",         // Tipo: systemd|tcp_port|http|process|command
    ...                       // Configuración específica del tipo
  },
  "actions": [                // Acciones a ejecutar en caso de fallo
    {
      "type": "string",       // Tipo: systemd_restart|command|webhook
      ...                     // Configuración específica
    }
  ],
  "thresholds": {             // Umbrales de fallo
    "max_failures": number,   // Fallos consecutivos antes de actuar
    "check_interval_seconds": number,  // Intervalo entre checks
    "restart_delay_seconds": number    // Espera después de restart
  }
}
```

### Tipos de Check Disponibles

#### 1. systemd
```json
{
  "type": "systemd",
  "systemd": {
    "service": "nombre.service"
  }
}
```

#### 2. tcp_port
```json
{
  "type": "tcp_port",
  "tcp_port": {
    "host": "localhost",
    "port": 3306,
    "timeout_seconds": 5
  }
}
```

#### 3. http
```json
{
  "type": "http",
  "http": {
    "url": "http://localhost:8080/health",
    "method": "GET",
    "expected_status": 200,
    "timeout_seconds": 10,
    "headers": {
      "Authorization": "Bearer token123"
    }
  }
}
```

#### 4. process
```json
{
  "type": "process",
  "process": {
    "name": "nginx",
    "cmdline_contains": "master process"
  }
}
```

#### 5. command
```json
{
  "type": "command",
  "command": {
    "cmd": "/usr/local/bin/check-health.sh",
    "expected_exit_code": 0
  }
}
```

### Tipos de Actions Disponibles

#### 1. systemd_restart
```json
{
  "type": "systemd_restart",
  "systemd_restart": {
    "service": "apache2.service"
  }
}
```

#### 2. command
```json
{
  "type": "command",
  "command": {
    "cmd": "/opt/scripts/restart-app.sh"
  }
}
```

#### 3. webhook
```json
{
  "type": "webhook",
  "webhook": {
    "url": "https://api.ejemplo.com/restart",
    "method": "POST",
    "headers": {
      "Authorization": "Bearer token123"
    },
    "body": "{\"service\": \"apache\"}"
  }
}
```

---

## 🔧 Ejemplos Prácticos Completos

### Monitorear Stack Web Completo

```bash
# 1. Servidor Web Nginx
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx",
    "enabled": true,
    "check": {
      "type": "systemd",
      "systemd": {"service": "nginx.service"}
    },
    "actions": [{
      "type": "systemd_restart",
      "systemd_restart": {"service": "nginx.service"}
    }],
    "thresholds": {
      "max_failures": 3,
      "check_interval_seconds": 30,
      "restart_delay_seconds": 5
    }
  }'

# 2. Base de Datos MySQL
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "mysql",
    "enabled": true,
    "check": {
      "type": "tcp_port",
      "tcp_port": {"host": "localhost", "port": 3306, "timeout_seconds": 5}
    },
    "actions": [{
      "type": "systemd_restart",
      "systemd_restart": {"service": "mysql.service"}
    }],
    "thresholds": {
      "max_failures": 5,
      "check_interval_seconds": 60,
      "restart_delay_seconds": 10
    }
  }'

# 3. API Backend
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "api",
    "enabled": true,
    "check": {
      "type": "http",
      "http": {
        "url": "http://localhost:8000/health",
        "method": "GET",
        "expected_status": 200,
        "timeout_seconds": 10
      }
    },
    "actions": [{
      "type": "command",
      "command": {"cmd": "systemctl restart api.service"}
    }],
    "thresholds": {
      "max_failures": 2,
      "check_interval_seconds": 30,
      "restart_delay_seconds": 3
    }
  }'
```

### Gestionar Servicios

```bash
# Ver todos los servicios
curl -u admin:admin123 http://localhost:8080/api/targets

# Ver estado de uno específico
curl -u admin:admin123 http://localhost:8080/api/status | jq '.targets.nginx'

# Deshabilitar temporalmente
curl -u admin:admin123 -X PUT http://localhost:8080/api/targets/nginx/toggle

# Modificar configuración
curl -u admin:admin123 -X PUT http://localhost:8080/api/targets/nginx \
  -H "Content-Type: application/json" \
  -d @nginx-config.json

# Eliminar servicio
curl -u admin:admin123 -X DELETE http://localhost:8080/api/targets/nginx
```

---

## 🐛 Códigos de Error

| Código | Descripción |
|--------|-------------|
| 200 | OK - Operación exitosa |
| 201 | Created - Recurso creado |
| 400 | Bad Request - JSON inválido o campos faltantes |
| 401 | Unauthorized - Credenciales incorrectas |
| 404 | Not Found - Servicio no existe |
| 409 | Conflict - Ya existe (creación duplicada) |
| 500 | Internal Server Error - Error del servidor |

---

## 📝 Notas Importantes

### ✅ Ventajas

- **Sin reinicio**: Cambios aplicados inmediatamente
- **Persistente**: Guarda automáticamente en YAML
- **Seguro**: Autenticación en todas las peticiones
- **RESTful**: API estándar fácil de integrar
- **JSON**: Formato universal y fácil de usar

### ⚠️ Consideraciones

1. **Autenticación obligatoria**: Todas las peticiones requieren usuario/contraseña
2. **HTTPS en producción**: Usa proxy reverso (nginx) para HTTPS
3. **Validación**: La API valida que el JSON sea correcto antes de guardar
4. **Backup**: Haz backup del YAML antes de cambios masivos
5. **Firewall**: Protege el puerto 8080 con firewall si expones a internet

---

## 🔗 Integración con Herramientas

### Postman

Importa esta colección para probar la API:

1. Crear colección "Neon Watchdog"
2. Configurar Auth: Type: Basic Auth, Username: admin, Password: admin123
3. Añadir requests para cada endpoint

### curl + jq (Terminal)

```bash
# Listar servicios ordenados por salud
curl -s -u admin:admin123 http://localhost:8080/api/status | \
  jq '.targets | to_entries | sort_by(.value.healthy) | from_entries'

# Contar servicios por estado
curl -s -u admin:admin123 http://localhost:8080/api/status | \
  jq '[.targets[] | select(.enabled)] | group_by(.healthy) | 
      map({status: (if .[0].healthy then "healthy" else "unhealthy" end), 
           count: length})'

# Ver solo servicios con problemas
curl -s -u admin:admin123 http://localhost:8080/api/status | \
  jq '.targets | to_entries | map(select(.value.enabled and .value.healthy == false))'
```

### Python Script

```python
import requests
from requests.auth import HTTPBasicAuth

BASE_URL = "http://localhost:8080"
AUTH = HTTPBasicAuth('admin', 'admin123')

# Listar servicios
response = requests.get(f"{BASE_URL}/api/targets", auth=AUTH)
targets = response.json()['targets']

for target in targets:
    print(f"{target['name']}: {'✓' if target['enabled'] else '✗'}")

# Añadir servicio
new_service = {
    "name": "redis",
    "enabled": True,
    "check": {
        "type": "tcp_port",
        "tcp_port": {"host": "localhost", "port": 6379, "timeout_seconds": 5}
    },
    "actions": [{
        "type": "systemd_restart",
        "systemd_restart": {"service": "redis.service"}
    }],
    "thresholds": {
        "max_failures": 3,
        "check_interval_seconds": 30,
        "restart_delay_seconds": 5
    }
}

response = requests.post(f"{BASE_URL}/api/targets", json=new_service, auth=AUTH)
print(response.json())
```

---

## 📚 Recursos Adicionales

- **Dashboard Web**: http://localhost:8080 (interfaz visual)
- **Documentación de Checks**: Ver implementaciones en `internal/checks/`
- **Documentación de Actions**: Ver implementaciones en `internal/actions/`
- **Usuarios**: Ver [USUARIOS-LOGIN.txt](USUARIOS-LOGIN.txt)

---

## 🎯 Próximos Pasos

1. **Prueba la API** con los ejemplos de curl
2. **Crea tus servicios** adaptando los ejemplos
3. **Integra con tus scripts** usando Python/Bash
4. **Monitorea en tiempo real** con el dashboard web

**¿Necesitas ayuda?** Revisa los logs en tiempo real:
```bash
tail -f neon-watchdog.log
```
