# 🐺 Neon Watchdog v2.0 - Complete Feature List

## 🆕 What's New in v2.0

### 🔥 TIER 1: High Impact Features (IMPLEMENTED)

#### 1. ✅ Notificaciones (Email, Webhook, Telegram)
- **Email**: SMTP con TLS para alertas por correo
- **Webhook**: Integración con Slack, Discord, PagerDuty, etc.
- **Telegram**: Bot notifications con formato Markdown

**Configuración:**
```yaml
notifications:
  - type: email
    enabled: true
    email:
      smtp_host: smtp.gmail.com
      smtp_port: 587
      username: your-email@gmail.com
      password: your-app-password
      from: watchdog@example.com
      to: [admin@example.com]
      use_tls: true

  - type: webhook
    enabled: true
    webhook:
      url: https://hooks.slack.com/services/YOUR/WEBHOOK
      method: POST
      timeout: 10

  - type: telegram
    enabled: true
    telegram:
      bot_token: "BOT_TOKEN_HERE"
      chat_id: "CHAT_ID_HERE"
```

#### 2. ✅ Métricas Prometheus
- Exporta métricas en formato Prometheus
- Endpoint HTTP configurable
- Métricas incluidas:
  - `neon_watchdog_uptime_seconds`
  - `neon_watchdog_target_healthy`
  - `neon_watchdog_checks_total`
  - `neon_watchdog_checks_failed_total`
  - `neon_watchdog_check_duration_seconds`
  - `neon_watchdog_consecutive_failures`
  - `neon_watchdog_recoveries_total`

**Configuración:**
```yaml
metrics:
  enabled: true
  port: 9090
  path: /metrics
```

**Usar con Prometheus:**
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'neon-watchdog'
    static_configs:
      - targets: ['localhost:9090']
```

#### 3. ✅ HTTP Health Checks Nativos
- Health checks HTTP sin necesidad de curl
- Configuración de headers, body, status esperado
- Timeout configurable

**Ejemplo:**
```yaml
checks:
  - type: http
    http:
      url: http://localhost:8080/health
      method: GET
      expected_status: 200
      headers:
        Authorization: "Bearer token123"
      timeout_seconds: 5
```

#### 4. ✅ Pre/Post Hooks
- Hooks antes de restart
- Hooks después de restart
- Hooks en caso de fallo

**Ejemplo:**
```yaml
action:
  type: systemd
  systemd:
    unit: nginx.service
  hooks:
    before_restart:
      - /usr/local/bin/backup-logs.sh
    after_restart:
      - /usr/local/bin/notify-team.sh nginx recovered
    on_failure:
      - /usr/local/bin/emergency-alert.sh nginx failed
```

### 🎯 TIER 2: Alto Valor (IMPLEMENTED)

#### 5. ✅ Dependency Chains
- Los targets pueden depender de otros
- No reinicia servicios si sus dependencias están caídas
- Orden de verificación automático

**Ejemplo:**
```yaml
targets:
  - name: postgresql
    enabled: true
    checks:
      - type: process_name
        process_name: postgres

  - name: backend-api
    enabled: true
    depends_on:
      - postgresql  # No reinicia si PostgreSQL está caído
    checks:
      - type: http
        url: http://localhost:3000/health
```

#### 6. ✅ Graceful Shutdown Detection
- Ignora exit codes específicos (ej: SIGTERM durante deploy)
- Evita false positives durante reinicios planeados

**Ejemplo:**
```yaml
checks:
  - type: process_name
    process_name: myapp
    ignore_exit_codes: [0, 143]  # 0=normal, 143=SIGTERM
```

#### 7. ✅ Rate Limiting Inteligente (Exponential Backoff)
- Backoff lineal o exponencial
- Evita crashlooping
- Tiempo máximo de backoff configurable

**Ejemplo:**
```yaml
policy:
  backoff_strategy: exponential  # 1m, 2m, 4m, 8m, 16m...
  max_backoff_seconds: 3600      # Max 1 hora
```

#### 8. ✅ Config Hot-Reload
- Recarga configuración sin reiniciar el watchdog
- Enviar señal SIGHUP al proceso
- No interrumpe checks en progreso

**Uso:**
```bash
# Obtener PID
PID=$(pgrep neon-watchdog)

# Recargar configuración
kill -HUP $PID
```

### 🔮 TIER 3: Nice to Have (IMPLEMENTED)

#### 9. ✅ Multi-Check Logic (AND/OR)
- Combina múltiples checks con lógica booleana
- AND: Todos deben pasar
- OR: Al menos uno debe pasar

**Ejemplo AND:**
```yaml
checks:
  - type: logic
    logic: AND
    checks:
      - type: process_name
        process_name: nginx
      - type: tcp_port
        tcp_port: "80"
```

**Ejemplo OR (alta disponibilidad):**
```yaml
checks:
  - type: logic
    logic: OR
    checks:
      - type: tcp_port
        tcp_port: "8080"
      - type: tcp_port
        tcp_port: "8081"  # Fallback port
```

#### 10. ✅ Dashboard Web Básico
- UI web con estado en tiempo real
- API REST JSON
- Health check endpoint
- Auto-refresh cada 5 segundos
- Sin autenticación (para MVP interno)

**Configuración:**
```yaml
dashboard:
  enabled: true
  port: 8080
  path: /
```

**Endpoints:**
- `GET /` - UI web
- `GET /api/status` - Estado completo JSON
- `GET /api/health` - Health check simple

#### 11. ✅ Custom Health Scripts con Exit Codes
- Ejecuta scripts personalizados
- Soporte para exit codes de warning (no reinicia)
- Captura stdout/stderr

**Ejemplo:**
```yaml
checks:
  - type: script
    script:
      path: /opt/healthchecks/db-check.sh
      args: [--strict, --timeout=5]
      success_exit_codes: [0]
      warning_exit_codes: [1]  # Warning pero no restart
```

#### 12. ✅ Estado Persistente Mejorado con Historial
- Historial de eventos con timestamps
- Estadísticas agregadas por target
- Retención configurable
- Análisis post-mortem

**Configuración:**
```yaml
history:
  max_entries: 1000
  retention_hours: 168  # 7 días
```

**Estadísticas incluidas:**
- Total checks por target
- Checks fallidos/exitosos
- Total de recuperaciones
- Último check/fallo/recuperación
- Fallos consecutivos actuales

---

## 📊 Tipos de Checks Disponibles

| Tipo | Descripción | Ejemplo |
|------|-------------|---------|
| `process_name` | Verifica si proceso existe | `process_name: nginx` |
| `pid_file` | Lee PID de archivo y verifica | `pid_file: /var/run/app.pid` |
| `tcp_port` | Conexión TCP | `tcp_port: "127.0.0.1:80"` |
| `command` | Ejecuta comando | `command: [curl, -f, http://localhost]` |
| `http` | HTTP health check nativo | `http: {url: ..., expected_status: 200}` |
| `script` | Script personalizado | `script: {path: ..., success_exit_codes: [0]}` |
| `logic` | Combina checks (AND/OR) | `logic: AND, checks: [...]` |

## ⚙️ Tipos de Acciones Disponibles

| Tipo | Descripción | Ejemplo |
|------|-------------|---------|
| `systemd` | Operación systemd | `systemd: {unit: nginx.service, method: restart}` |
| `exec` | Ejecuta comandos | `exec: {restart: [/usr/bin/restart.sh]}` |

## 🔧 Política de Reintentos

```yaml
default_policy:
  fail_threshold: 1                  # Fallos antes de reiniciar
  restart_cooldown_seconds: 60       # Mínimo entre reinicios
  max_restarts_per_hour: 10          # Rate limiting
  backoff_strategy: exponential      # linear o exponential
  max_backoff_seconds: 3600          # Límite de backoff
```

## 🚀 Quick Start con Todas las Features

### 1. Configuración Básica con Notificaciones

```yaml
# config.yml
interval_seconds: 30
log_level: INFO

# Métricas Prometheus
metrics:
  enabled: true
  port: 9090

# Dashboard web
dashboard:
  enabled: true
  port: 8080

# Notificaciones
notifications:
  - type: webhook
    enabled: true
    webhook:
      url: https://hooks.slack.com/services/YOUR/WEBHOOK

targets:
  - name: nginx
    enabled: true
    checks:
      - type: http
        http:
          url: http://localhost/
          expected_status: 200
    action:
      type: systemd
      systemd:
        unit: nginx.service
      hooks:
        after_restart:
          - echo "Nginx restarted" >> /var/log/watchdog-events.log
```

### 2. Ejecutar

```bash
# Instalar
sudo make install

# Habilitar timer
sudo systemctl enable --now neon-watchdog.timer

# Ver logs
journalctl -u neon-watchdog -f

# Ver métricas
curl http://localhost:9090/metrics

# Ver dashboard
open http://localhost:8080
```

## 📈 Integración con Stack Moderno

### Prometheus + Grafana

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'neon-watchdog'
    scrape_interval: 30s
    static_configs:
      - targets: ['localhost:9090']
```

### Alertmanager

```yaml
# alertmanager.yml
receivers:
  - name: 'watchdog-alerts'
    webhook_configs:
      - url: 'http://localhost:9000/alert'
```

### Grafana Dashboard

Importa métricas:
- `neon_watchdog_target_healthy`
- `neon_watchdog_checks_failed_total`
- `neon_watchdog_recoveries_total`

## 🆚 Comparación v1.0 vs v2.0

| Feature | v1.0 | v2.0 |
|---------|------|------|
| Process checks | ✅ | ✅ |
| TCP port checks | ✅ | ✅ |
| HTTP checks | ❌ | ✅ |
| Script checks | ❌ | ✅ |
| Logic groups | ❌ | ✅ |
| Notificaciones | ❌ | ✅ (3 tipos) |
| Métricas | ❌ | ✅ Prometheus |
| Dashboard | ❌ | ✅ Web UI |
| Hooks | ❌ | ✅ Before/After/OnFailure |
| Dependencies | ❌ | ✅ Dependency chains |
| Backoff | Linear | ✅ Exponential |
| Hot reload | ❌ | ✅ SIGHUP |
| Historial | Básico | ✅ Avanzado |

## 📝 Migration Guide v1 → v2

Tu configuración v1 sigue funcionando en v2. Las nuevas features son opcionales:

```yaml
# v1 (sigue funcionando)
targets:
  - name: nginx
    enabled: true
    checks:
      - type: process_name
        process_name: nginx
    action:
      type: systemd
      systemd:
        unit: nginx.service

# v2 (con nuevas features)
metrics:
  enabled: true
  port: 9090

dashboard:
  enabled: true
  port: 8080

targets:
  - name: nginx
    enabled: true
    checks:
      - type: http  # NUEVO
        http:
          url: http://localhost/
    action:
      type: systemd
      systemd:
        unit: nginx.service
      hooks:  # NUEVO
        after_restart:
          - /usr/local/bin/notify.sh
```

## 🎯 Casos de Uso

### 1. Microservicios con Dependencias
```yaml
targets:
  - name: database
    checks: [...]
    
  - name: cache
    checks: [...]
    
  - name: api
    depends_on: [database, cache]
    checks: [...]
```

### 2. Alta Disponibilidad (múltiples backends)
```yaml
checks:
  - type: logic
    logic: OR
    checks:
      - type: http
        http: {url: "http://backend1:8080/health"}
      - type: http
        http: {url: "http://backend2:8080/health"}
```

### 3. Health Check Complejo
```yaml
checks:
  - type: logic
    logic: AND
    checks:
      - type: process_name
        process_name: app
      - type: http
        http: {url: "http://localhost:8080/health"}
      - type: script
        script: {path: "/opt/check-db-connection.sh"}
```

---

## 🔗 Enlaces

- **GitHub**: https://github.com/tgextreme/neon-watchdog
- **Documentación**: Ver `INSTALL.md`, `COMMANDS.md`
- **Ejemplos**: `examples/config-v2-full.yml`

## 📄 Licencia

MIT License - Tomás González (@tgextreme)
