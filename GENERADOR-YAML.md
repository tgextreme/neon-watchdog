# 📝 Generador de Servicios YAML - Neon Watchdog

## Descripción

Sistema de generación de configuraciones YAML para el daemon de Neon Watchdog a través de una interfaz web intuitiva.

## 🎯 Características

✅ **Formulario web interactivo** para crear servicios  
✅ **Vista previa del YAML** antes de generar  
✅ **Múltiples tipos de verificación** (check):
- Nombre de proceso
- Puerto TCP
- Archivo PID  
- Comando personalizado
- HTTP Health Check

✅ **Múltiples tipos de acción** (restart):
- Systemd (systemctl)
- Comando personalizado (exec)
- Docker containers

✅ **Políticas de reinicio configurables**
✅ **Auto-completado inteligente** de campos
✅ **Validación de formularios**

## 📂 Ubicación de Archivos

### Archivos generados
```
/etc/neon-watchdog/services.d/
├── [nombre-servicio].yml
├── [nombre-servicio2].yml
└── ...
```

### Backend
```
/var/www/html/app-gestion-neon-watchdogs/
├── services.php              # Página principal con formulario
└── api/
    └── generate_service_yaml.php   # API para generar YAML
```

## 🚀 Uso

### 1. Acceder al formulario

Navega a: `http://localhost/app-gestion-neon-watchdogs/services.php`

Haz clic en el botón **"Nuevo Servicio"**

### 2. Completar el formulario

#### Información Básica
- **Nombre del Servicio**: Identificador único (solo minúsculas, números, guiones)
- **Nombre para Mostrar**: Nombre descriptivo
- **Habilitar**: Si el servicio debe estar activo inmediatamente

#### Configuración de Verificación (Check)
Selecciona el tipo de check y completa los campos correspondientes:

**Ejemplo: Nginx con proceso + puerto**
```
Tipo: Nombre de Proceso
Nombre del Proceso: nginx
Puerto TCP Adicional: 80
```

#### Configuración de Acción (Reinicio)
Define cómo se reiniciará el servicio:

**Ejemplo: Systemd**
```
Tipo de Acción: Systemd
Unidad Systemd: nginx.service
Método: restart
```

#### Política de Reinicio (Opcional)
- **Umbral de Fallos**: Número de fallos antes de reiniciar
- **Cooldown**: Tiempo mínimo entre reinicios (segundos)
- **Máx. Reinicios/Hora**: Límite de reinicios por hora

### 3. Vista Previa
Haz clic en **"Vista Previa"** para ver el YAML que se generará

### 4. Generar YAML
Haz clic en **"Generar YAML"** para crear el archivo

### 5. Aplicar cambios
```bash
sudo systemctl restart neon-watchdog
```

## 📋 Ejemplos de Configuración

### Ejemplo 1: Nginx Web Server
```yaml
targets:
  - name: nginx
    enabled: true
    checks:
      - type: process_name
        process_name: nginx
      - type: tcp_port
        tcp_port: "80"
    action:
      type: systemd
      systemd:
        unit: nginx.service
        method: restart
    policy:
      fail_threshold: 1
      restart_cooldown_seconds: 60
      max_restarts_per_hour: 10
```

### Ejemplo 2: Aplicación con HTTP Health Check
```yaml
targets:
  - name: myapi
    enabled: true
    checks:
      - type: command
        command:
          - /usr/bin/curl
          - -fsS
          - --max-time
          - "5"
          - http://127.0.0.1:8080/health
    action:
      type: systemd
      systemd:
        unit: myapi.service
        method: restart
```

### Ejemplo 3: Contenedor Docker
```yaml
targets:
  - name: redis
    enabled: true
    checks:
      - type: tcp_port
        tcp_port: "6379"
    action:
      type: exec
      exec:
        restart:
          - /usr/bin/docker
          - restart
          - redis
```

## 🔧 API Endpoint

### POST `/api/generate_service_yaml.php`

**Request Body (JSON):**
```json
{
  "name": "nginx",
  "display_name": "Nginx Web Server",
  "enabled": true,
  "check_type": "process_name",
  "process_name": "nginx",
  "additional_tcp_port": "80",
  "action_type": "systemd",
  "systemd_unit": "nginx.service",
  "systemd_method": "restart",
  "fail_threshold": 1,
  "restart_cooldown": 60,
  "max_restarts": 10
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Archivo YAML generado exitosamente",
  "filepath": "/etc/neon-watchdog/services.d/nginx.yml",
  "yaml": "# Configuración para: Nginx Web Server\n..."
}
```

**Response (Error):**
```json
{
  "success": false,
  "message": "Error al generar YAML: [detalle del error]"
}
```

## 🛡️ Seguridad

- ✅ Requiere autenticación de sesión
- ✅ Validación de campos obligatorios
- ✅ Sanitización de nombres de archivo
- ✅ Permisos correctos en archivos generados (0644)
- ✅ Registro en audit log de todas las operaciones

## 📝 Audit Log

Todas las generaciones de YAML se registran:
```sql
SELECT * FROM audit_logs 
WHERE action = 'service_yaml_created';
```

Detalles incluidos:
- Usuario que generó el archivo
- Nombre del servicio
- Ruta del archivo
- Timestamp
- IP y User Agent

## 🔄 Integración con Neon Watchdog Daemon

### Configuración del Daemon

Edita `/etc/neon-watchdog/config.yml` para incluir el directorio de servicios:

```yaml
# Incluir todos los archivos YAML del directorio services.d
include:
  - /etc/neon-watchdog/services.d/*.yml

# O cargar el directorio completo
config_dir: /etc/neon-watchdog/services.d
```

### Recargar configuración

Después de generar nuevos servicios:

```bash
# Reiniciar el daemon
sudo systemctl restart neon-watchdog

# O si soporta reload
sudo systemctl reload neon-watchdog

# Verificar el estado
sudo systemctl status neon-watchdog
```

## 🐛 Troubleshooting

### El archivo no se genera
```bash
# Verificar permisos del directorio
ls -la /etc/neon-watchdog/services.d/

# Verificar logs de Apache/PHP
sudo tail -f /var/log/apache2/error.log

# Verificar que www-data tenga permisos de escritura
sudo chown -R www-data:www-data /etc/neon-watchdog
```

### Error de sesión
```bash
# Verificar que estás autenticado
# Refresca la página de login: http://localhost/app-gestion-neon-watchdogs/login.php
```

### El daemon no carga la configuración
```bash
# Verificar sintaxis YAML
yamllint /etc/neon-watchdog/services.d/*.yml

# Ver logs del daemon
sudo journalctl -u neon-watchdog -f
```

## 📚 Referencias

- [API-REST.md](API-REST.md) - Documentación completa de la API REST
- [config.yml](examples/config.yml) - Ejemplo de configuración completa
- [AUTENTICACION.md](AUTENTICACION.md) - Sistema de autenticación

## 🎨 Auto-completado de Campos

El formulario incluye auto-completado inteligente:

- **Nombre del Servicio → Display Name**: `nginx` → `Nginx Service`
- **Nombre del Servicio → Proceso**: `nginx` → `nginx`
- **Nombre del Servicio → Systemd Unit**: `nginx` → `nginx.service`
- **Nombre del Servicio → Docker Container**: `nginx` → `nginx`

Esto acelera la configuración de servicios comunes.

## ✨ Próximas Mejoras

- [ ] Soporte para múltiples checks del mismo tipo
- [ ] Importar/exportar configuraciones
- [ ] Templates predefinidos (NGINX, Apache, MySQL, etc.)
- [ ] Validación de configuración antes de generar
- [ ] Editor visual de YAML
- [ ] Clonar servicios existentes

---

**Autor**: Neon Watchdog Team  
**Fecha**: Enero 2026  
**Versión**: 1.0
