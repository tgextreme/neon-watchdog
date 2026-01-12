# Neon Watchdog 🐺

**Linux Process Guardian** - Monitor y recuperación automática de procesos/servicios en Linux.

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Go](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

## 🎯 ¿Qué es Neon Watchdog?

Neon Watchdog es un sistema ligero y robusto para monitorizar servicios críticos en Linux y ejecutar acciones de recuperación automática cuando detecta fallos. Diseñado para integrarse nativamente con **systemd** y ofrecer máxima confiabilidad con mínimo overhead.

### 🌟 Características Principales

✅ **Múltiples Tipos de Checks:**
- Verificación de procesos por nombre o PID
- Healthcheck de puertos TCP
- Validación HTTP/HTTPS con códigos de estado
- Ejecución de scripts personalizados
- Comprobación de servicios systemd
- Checks lógicos (AND/OR) para validaciones complejas

✅ **Acciones de Recuperación Inteligentes:**
- Restart/start/stop de servicios systemd
- Ejecución de comandos personalizados
- Hooks before/after restart
- Hooks on failure

✅ **Políticas Avanzadas:**
- Umbrales de fallos consecutivos
- Cooldown entre reinicios
- Rate limiting (max reinicios por hora)
- Estrategias de backoff (linear/exponential)
- Gestión de dependencias entre servicios

✅ **Dashboard Web y API REST:**
- Interfaz web para visualización del estado
- API REST completa con autenticación
- Gestión de servicios sin editar archivos
- Actualización en tiempo real

✅ **Métricas y Notificaciones:**
- Exportación de métricas Prometheus
- Notificaciones por email, webhook y Telegram
- Historial persistente de eventos
- Logging estructurado compatible con journald

✅ **Integración con systemd:**
- Modo timer (oneshot) - recomendado
- Modo daemon (persistente)
- Logs integrados con journalctl

---

## 📚 Documentación

- **[INSTALL.md](INSTALL.md)** - Guía completa de instalación y configuración
- **[API-REST.md](API-REST.md)** - Documentación de la API REST y dashboard web
- **[LICENSE](LICENSE)** - Licencia MIT

---

## 🚀 Inicio Rápido

### Instalación en 3 Pasos

```bash
# 1. Compilar e instalar
git clone https://github.com/tgextreme/neon-watchdog.git
cd neon-watchdog
make build && sudo make install

# 2. Validar configuración
neon-watchdog test-config -c /etc/neon-watchdog/config.yml

# 3. Activar systemd timer
sudo systemctl enable --now neon-watchdog.timer
```

### Configuración Básica

Edita `/etc/neon-watchdog/config.yml`:

```yaml
log_level: INFO
timeout_seconds: 10
state_file: /var/lib/neon-watchdog/state.json

default_policy:
  fail_threshold: 1
  restart_cooldown_seconds: 60
  max_restarts_per_hour: 10

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
```

### Verificar Estado

```bash
# Ver logs en tiempo real
journalctl -u neon-watchdog.service -f

# Ver estado del timer
systemctl status neon-watchdog.timer

# Ejecutar check manual
neon-watchdog check -c /etc/neon-watchdog/config.yml --verbose
```

---

## 📖 Uso

### Comandos Disponibles

```bash
# Ejecutar checks una vez (para systemd timer)
neon-watchdog check -c /etc/neon-watchdog/config.yml

# Ejecutar como daemon (loop continuo)
neon-watchdog run -c /etc/neon-watchdog/config.yml

# Validar configuración
neon-watchdog test-config -c /etc/neon-watchdog/config.yml

# Versión
neon-watchdog version

# Ayuda
neon-watchdog help
```

### Opciones

- `-c, --config <path>`: Ruta al archivo de configuración (requerido)
- `--verbose`: Activar logging detallado (DEBUG)
- `--dry-run`: No ejecutar acciones de recuperación (solo simular)

---

## 🔧 Tipos de Checks

### 1. Process Name

Verifica si existe un proceso con el nombre especificado:

```yaml
- type: process_name
  process_name: nginx
```

### 2. PID File

Valida que el PID en el archivo existe:

```yaml
- type: pid_file
  pid_file: /var/run/myapp.pid
```

### 3. TCP Port

Intenta conectar a un puerto TCP:

```yaml
- type: tcp_port
  tcp_port: "127.0.0.1:8080"  # o solo "8080"
```

### 4. HTTP Check

Realiza petición HTTP y valida el código de estado:

```yaml
- type: http
  http:
    url: http://localhost:8080/health
    method: GET
    expected_status: 200
    timeout_seconds: 5
```

### 5. Command

Ejecuta un comando y verifica el exit code (0 = success):

```yaml
- type: command
  command:
    - /usr/bin/curl
    - -fsS
    - http://localhost:8080/health
```

### 6. Script

Ejecuta un script personalizado:

```yaml
- type: script
  script:
    path: /usr/local/bin/check-app.sh
    args: ["--verbose"]
    success_exit_codes: [0]
    warning_exit_codes: [1]
```

### 7. Logic Groups

Combina múltiples checks con AND/OR:

```yaml
- type: logic
  logic: AND  # o OR
  checks:
    - type: process_name
      process_name: nginx
    - type: tcp_port
      tcp_port: "80"
```

---

## ⚙️ Tipos de Acciones

### 1. Systemd

Ejecuta `systemctl` sobre una unidad:

```yaml
action:
  type: systemd
  systemd:
    unit: nginx.service
    method: restart  # restart, start, stop
```

### 2. Exec

Ejecuta comandos personalizados:

```yaml
action:
  type: exec
  exec:
    restart:
      - /usr/local/bin/restart-app.sh
      - "--force"
```

### 3. Action Hooks

Ejecuta comandos antes/después de acciones:

```yaml
action:
  type: systemd
  systemd:
    unit: myapp.service
    method: restart
  hooks:
    before_restart:
      - /usr/local/bin/backup-state.sh
    after_restart:
      - /usr/local/bin/verify-startup.sh
    on_failure:
      - /usr/local/bin/alert-admin.sh
```

---

## 📊 Dashboard Web y API REST

### Habilitar Dashboard

Añade a tu `config.yml`:

```yaml
dashboard:
  enabled: true
  port: 8080
  path: "/"
```

### Crear Usuario

```bash
# Crear archivo de usuarios
htpasswd -B -c users.txt admin

# O usar generador incluido
./scripts/create-user.sh admin password123
```

### Acceder

```bash
# Dashboard web
http://localhost:8080/

# API REST
curl -u admin:password http://localhost:8080/api/status
```

Ver [API-REST.md](API-REST.md) para documentación completa de la API.

---

## 📈 Métricas Prometheus

### Habilitar Métricas

```yaml
metrics:
  enabled: true
  port: 9090
  path: /metrics
```

### Métricas Disponibles

- `neon_watchdog_check_total` - Total de checks ejecutados
- `neon_watchdog_check_failures_total` - Total de checks fallidos
- `neon_watchdog_action_total` - Total de acciones ejecutadas
- `neon_watchdog_action_failures_total` - Total de acciones fallidas
- `neon_watchdog_target_healthy` - Estado actual de cada target (1=healthy, 0=unhealthy)
- `neon_watchdog_check_duration_seconds` - Duración de los checks

---

## 🔔 Notificaciones

### Email

```yaml
notifications:
  - type: email
    enabled: true
    email:
      smtp_host: smtp.gmail.com
      smtp_port: 587
      username: alert@example.com
      password: ${SMTP_PASSWORD}
      from: neon-watchdog@example.com
      to:
        - admin@example.com
      use_tls: true
```

### Webhook

```yaml
notifications:
  - type: webhook
    enabled: true
    webhook:
      url: https://hooks.slack.com/services/YOUR/WEBHOOK/URL
      method: POST
      headers:
        Content-Type: application/json
      timeout: 10
```

### Telegram

```yaml
notifications:
  - type: telegram
    enabled: true
    telegram:
      bot_token: ${TELEGRAM_BOT_TOKEN}
      chat_id: "-1001234567890"
```

---

## 📊 Logs

### Ver Logs con journalctl

```bash
# Logs en tiempo real
journalctl -u neon-watchdog.service -f

# Últimas 100 líneas
journalctl -u neon-watchdog.service -n 100

# Filtrar por nivel
journalctl -u neon-watchdog.service -p err

# Filtrar por target específico
journalctl -u neon-watchdog.service | grep 'target=nginx'
```

### Formato de Logs

Logs estructurados en formato `clave=valor`:

```
2026-01-09T10:30:45Z level=INFO msg="target healthy" target=nginx check=process_name latency_ms=2
2026-01-09T10:31:15Z level=WARN msg="check failed" target=api check=tcp_port error="connection refused"
2026-01-09T10:31:15Z level=INFO msg="executing action" target=api action="systemd:restart"
```

---

## 🛠️ Troubleshooting

### El watchdog no detecta el proceso

```bash
# Verificar nombre exacto del proceso
ps aux | grep nombre

# Probar con pgrep
pgrep -x nombre

# Usar --verbose para ver detalles
neon-watchdog check -c config.yml --verbose
```

### Permisos denegados en systemctl

Si ejecutas como usuario no-root, configura sudoers:

```bash
# /etc/sudoers.d/neon-watchdog
neon-watchdog ALL=(root) NOPASSWD: /bin/systemctl restart nginx.service
```

### El timer no se ejecuta

```bash
# Verificar estado del timer
systemctl status neon-watchdog.timer

# Ver cuándo se ejecutará
systemctl list-timers | grep neon

# Habilitar y arrancar
sudo systemctl enable --now neon-watchdog.timer
```

### Ver Debug Completo

```bash
# Ejecutar con verbose
neon-watchdog check -c config.yml --verbose

# O cambiar nivel en config
log_level: DEBUG
```

---

## 🏗️ Arquitectura

```
┌─────────────────────────────────────────┐
│          neon-watchdog CLI              │
│  (run | check | test-config)            │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│      Configuration Loader               │
│      (YAML/JSON + Validation)           │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│             Engine                      │
│  ┌────────────────────────────────┐    │
│  │  For each target:              │    │
│  │   1. Run checks                │    │
│  │   2. Evaluate policy           │    │
│  │   3. Execute recovery action   │    │
│  └────────────────────────────────┘    │
└───┬─────────────────────────────┬───────┘
    │                             │
    ▼                             ▼
┌──────────────┐          ┌──────────────┐
│   Checkers   │          │   Actions    │
│ - Process    │          │ - Systemd    │
│ - PID file   │          │ - Exec       │
│ - TCP port   │          │ - Hooks      │
│ - HTTP       │          └──────────────┘
│ - Script     │
│ - Logic      │
└──────────────┘
```

---

## 📄 Licencia

MIT License - Ver [LICENSE](LICENSE) para más detalles.

---

## 🤝 Contribuir

¡Las contribuciones son bienvenidas! Por favor:

1. Haz fork del proyecto
2. Crea una rama para tu feature (`git checkout -b feature/amazing`)
3. Commit tus cambios (`git commit -am 'Add amazing feature'`)
4. Push a la rama (`git push origin feature/amazing`)
5. Abre un Pull Request

---

## 📮 Contacto

- **GitHub**: [github.com/tgextreme/neon-watchdog](https://github.com/tgextreme/neon-watchdog)
- **Issues**: [github.com/tgextreme/neon-watchdog/issues](https://github.com/tgextreme/neon-watchdog/issues)

---

**Hecho con ❤️ para mantener tus servicios siempre en marcha**
