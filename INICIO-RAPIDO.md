# 🚀 Neon Watchdog - Configuración Local Lista

## ✅ ¡Todo Configurado y Funcionando!

La API REST está completamente funcional en local. Se corrigió la documentación y se creó todo lo necesario.

---

## 🎯 Inicio Rápido

### 1. Iniciar el Watchdog

```bash
./run-local.sh
```

Esto iniciará:
- ✅ Neon Watchdog en modo local
- ✅ API REST en http://localhost:8080
- ✅ Dashboard web accesible
- ✅ Autenticación: `admin` / `admin123`

### 2. Verificar que Funciona

```bash
# Ver estado
curl -u admin:admin123 http://localhost:8080/api/status

# Listar servicios
curl -u admin:admin123 http://localhost:8080/api/targets
```

### 3. Crear un Servicio

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

### 4. Ver Dashboard Web

Abre en tu navegador: http://localhost:8080

Usuario: `admin`  
Contraseña: `admin123`

---

## 📋 Todos los Endpoints de la API

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/status` | Estado completo del watchdog |
| GET | `/api/health` | Health check rápido |
| GET | `/api/targets` | Listar todos los servicios |
| GET | `/api/targets/{name}` | Obtener servicio específico |
| POST | `/api/targets` | Crear nuevo servicio |
| PUT | `/api/targets/{name}` | Actualizar servicio |
| DELETE | `/api/targets/{name}` | Eliminar servicio |
| GET | `/api/config` | Ver configuración completa |

---

## 📦 Estructura de un Target (Servicio)

```json
{
  "name": "nombre-unico",
  "enabled": true,
  "checks": [
    {
      "type": "process_name | tcp_port | http | command | script",
      "process_name": "nginx",
      "tcp_port": "80",
      "http": {
        "url": "http://localhost/health",
        "method": "GET",
        "expected_status": 200
      },
      "command": ["pgrep", "nginx"],
      "script": {
        "path": "/path/to/script.sh"
      }
    }
  ],
  "action": {
    "type": "exec | systemd",
    "exec": {
      "restart": ["/bin/systemctl", "restart", "nginx"]
    },
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
}
```

---

## 🔧 Tipos de Checks Disponibles

### 1. process_name - Verificar proceso por nombre

```json
{
  "type": "process_name",
  "process_name": "nginx"
}
```

### 2. tcp_port - Verificar puerto TCP

```json
{
  "type": "tcp_port",
  "tcp_port": "80"
}
```

### 3. http - Health check HTTP

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

### 4. command - Ejecutar comando

```json
{
  "type": "command",
  "command": ["pgrep", "-f", "myapp"]
}
```

### 5. script - Ejecutar script personalizado

```json
{
  "type": "script",
  "script": {
    "path": "/opt/scripts/check-health.sh",
    "args": ["--verbose"],
    "success_exit_codes": [0]
  }
}
```

---

## 🎯 Tipos de Actions Disponibles

### 1. systemd - Gestionar servicio systemd

```json
{
  "type": "systemd",
  "systemd": {
    "unit": "nginx.service",
    "method": "restart"
  }
}
```

### 2. exec - Ejecutar comando personalizado

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

## 📝 Ejemplos Completos

### Monitorear Apache

```bash
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "apache",
    "enabled": true,
    "checks": [
      {"type": "tcp_port", "tcp_port": "80"}
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
  }'
```

### Monitorear MySQL

```bash
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "mysql",
    "enabled": true,
    "checks": [
      {"type": "tcp_port", "tcp_port": "3306"}
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

### Monitorear API con Health Check HTTP

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
        "restart": ["/usr/bin/systemctl", "restart", "api-backend.service"]
      }
    },
    "policy": {
      "fail_threshold": 2,
      "restart_cooldown_seconds": 60,
      "max_restarts_per_hour": 8
    }
  }'
```

---

## 🛠️ Comandos Útiles

### Ver Logs en Tiempo Real

```bash
tail -f neon-local.log
```

### Detener el Watchdog

```bash
pkill -f "neon-watchdog run"
```

### Reiniciar el Watchdog

```bash
./run-local.sh
```

### Ver Configuración Actual

```bash
curl -u admin:admin123 http://localhost:8080/api/config
```

---

## 📚 Archivos del Proyecto

| Archivo | Descripción |
|---------|-------------|
| `run-local.sh` | Script para iniciar en local |
| `config-local.yml` | Configuración local |
| `neon-local.log` | Logs de la aplicación |
| `state-local.json` | Estado persistente |
| `API-REST.md` | Documentación completa de la API |
| `users.txt` | Usuarios para autenticación |

---

## ✅ Prueba Rápida de la API

```bash
# 1. Crear servicio
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-ssh",
    "enabled": true,
    "checks": [{"type": "tcp_port", "tcp_port": "22"}],
    "action": {"type": "exec", "exec": {"restart": ["/bin/echo", "test"]}},
    "policy": {"fail_threshold": 3, "restart_cooldown_seconds": 60, "max_restarts_per_hour": 5}
  }'

# 2. Ver el servicio creado
curl -u admin:admin123 http://localhost:8080/api/targets/test-ssh

# 3. Listar todos
curl -u admin:admin123 http://localhost:8080/api/targets

# 4. Eliminar
curl -u admin:admin123 -X DELETE http://localhost:8080/api/targets/test-ssh
```

---

## 🎉 Resultado

✅ **API REST 100% funcional**  
✅ **Todos los endpoints operativos**  
✅ **CRUD completo (Create, Read, Update, Delete)**  
✅ **Configuración local con permisos correctos**  
✅ **Dashboard web accesible**

**¡Todo funciona perfectamente!** 🚀
