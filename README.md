# Neon Watchdog 🐺

**Linux Process Guardian** - Monitor y recuperación automática de procesos/servicios en Linux.

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Go](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

## 🎯 ¿Qué es Neon Watchdog?

Neon Watchdog es un sistema ligero y robusto para monitorizar servicios críticos en Linux y ejecutar acciones de recuperación automática cuando detecta fallos. Diseñado para integrarse nativamente con **systemd** y ofrecer máxima confiabilidad con mínimo overhead.

### Características principales

✅ **Múltiples tipos de checks:**
- Verificación de procesos por nombre
- Validación de PID files
- Healthcheck de puertos TCP
- Ejecución de comandos customizados

✅ **Acciones de recuperación:**
- Restart/start de servicios systemd
- Ejecución de scripts/comandos personalizados

✅ **Políticas inteligentes:**
- Umbrales de fallos consecutivos
- Cooldown entre reinicios
- Rate limiting (max reinicios por hora)

✅ **Integración con systemd:**
- Modo timer (oneshot) - **recomendado para MVP**
- Modo daemon (persistente)

✅ **Logging estructurado:**
- Formato clave=valor compatible con journald
- Niveles configurables (DEBUG/INFO/WARN/ERROR)

✅ **Configuración declarativa:**
- YAML o JSON
- Validación completa
- Hot-reload (futuro)

---

## 📦 Instalación

### Compilar desde código fuente

```bash
# Clonar el repositorio
git clone https://github.com/tgextreme/neon-watchdog.git
cd neon-watchdog

# Compilar
make build

# Instalar (requiere permisos sudo)
sudo make install
```

### Instalación manual

```bash
# Compilar el binario
go build -o neon-watchdog ./cmd/neon-watchdog

# Copiar binario
sudo cp neon-watchdog /usr/local/bin/

# Crear directorio de configuración
sudo mkdir -p /etc/neon-watchdog

# Copiar configuración de ejemplo
sudo cp examples/config.yml /etc/neon-watchdog/

# Instalar archivos systemd
sudo cp systemd/neon-watchdog.service /etc/systemd/system/
sudo cp systemd/neon-watchdog.timer /etc/systemd/system/
sudo systemctl daemon-reload
```

---

## 🚀 Inicio Rápido

### 1. Configurar targets

Edita `/etc/neon-watchdog/config.yml`:

```yaml
log_level: INFO
timeout_seconds: 10

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

### 2. Validar configuración

```bash
neon-watchdog test-config -c /etc/neon-watchdog/config.yml
```

### 3. Probar manualmente

```bash
# Ejecutar una sola vez
neon-watchdog check -c /etc/neon-watchdog/config.yml

# Ver logs
journalctl -f
```

### 4. Habilitar systemd timer (recomendado)

```bash
# Habilitar y arrancar el timer
sudo systemctl enable --now neon-watchdog.timer

# Ver estado
sudo systemctl status neon-watchdog.timer

# Ver logs
journalctl -u neon-watchdog.service -f
```

### Alternativa: Modo daemon

```bash
# Si prefieres modo daemon en lugar de timer
sudo systemctl enable --now neon-watchdog-daemon.service
```

---

## 📖 Uso

### Comandos disponibles

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

### Flags opcionales

- `-c, --config <path>`: Ruta al archivo de configuración (requerido)
- `--verbose`: Activar logging detallado (DEBUG)
- `--dry-run`: No ejecutar acciones de recuperación (solo simular)

---

## 🔧 Configuración

### Estructura del archivo de configuración

Ver [examples/config.yml](examples/config.yml) para ejemplos completos.

#### Sección global

```yaml
interval_seconds: 30          # Solo para modo daemon
timeout_seconds: 10           # Timeout para checks/acciones
log_level: INFO               # DEBUG, INFO, WARN, ERROR
state_file: /var/lib/neon-watchdog/state.json  # Opcional
```

#### Política por defecto

```yaml
default_policy:
  fail_threshold: 1                  # Fallos consecutivos antes de actuar
  restart_cooldown_seconds: 60       # Tiempo entre reinicios
  max_restarts_per_hour: 10          # Rate limit
```

#### Targets

Cada target representa un servicio/proceso a monitorizar:

```yaml
targets:
  - name: mi-servicio
    enabled: true
    checks:
      - type: process_name
        process_name: nginx
      - type: tcp_port
        tcp_port: "8080"
    action:
      type: systemd
      systemd:
        unit: mi-servicio.service
        method: restart
    policy:  # Opcional: sobrescribe default_policy
      fail_threshold: 2
      restart_cooldown_seconds: 120
```

### Tipos de checks

#### 1. Process name

Verifica si existe un proceso con el nombre especificado (usa `pgrep`):

```yaml
- type: process_name
  process_name: nginx
```

#### 2. PID file

Valida que el PID en el archivo existe y está corriendo:

```yaml
- type: pid_file
  pid_file: /var/run/myapp.pid
```

#### 3. TCP Port

Intenta conectar a un puerto TCP:

```yaml
- type: tcp_port
  tcp_port: "127.0.0.1:8080"  # o solo "8080" (asume localhost)
```

#### 4. Command

Ejecuta un comando y verifica el exit code (0 = success):

```yaml
- type: command
  command:
    - /usr/bin/curl
    - -fsS
    - --max-time
    - "5"
    - http://localhost:8080/health
```

### Tipos de acciones

#### 1. Systemd

Ejecuta `systemctl` sobre una unidad:

```yaml
action:
  type: systemd
  systemd:
    unit: nginx.service
    method: restart  # restart, start, stop
```

#### 2. Exec

Ejecuta comandos personalizados:

```yaml
action:
  type: exec
  exec:
    start:    # Se usa al primer fallo
      - /usr/local/bin/start-app.sh
    restart:  # Se usa en fallos subsiguientes
      - /usr/local/bin/restart-app.sh
```

---

## 📊 Logs y Observabilidad

### Ver logs con journalctl

```bash
# Logs en tiempo real
journalctl -u neon-watchdog.service -f

# Últimas 100 líneas
journalctl -u neon-watchdog.service -n 100

# Filtrar por nivel
journalctl -u neon-watchdog.service -p err

# Logs del timer
journalctl -u neon-watchdog.timer
```

### Formato de logs

Los logs usan formato estructurado `clave=valor`:

```
2026-01-09T10:30:45.123Z level=INFO msg="target healthy" target=nginx check=tcp_port result=OK latency_ms=2
2026-01-09T10:31:15.456Z level=WARN msg="check failed" target=api check=tcp_port reason="connection refused" latency_ms=5
2026-01-09T10:31:15.789Z level=INFO msg="executing recovery action" target=api action="systemd:restart api.service"
```

### Filtrar logs por target específico

```bash
journalctl -u neon-watchdog.service | grep 'target=nginx'
```

---

## 🔐 Seguridad y Confiabilidad

### Permisos

Para máxima seguridad, crea un usuario dedicado:

```bash
# Crear usuario
sudo useradd -r -s /bin/false -d /var/lib/neon-watchdog neon-watchdog

# Crear directorio de estado
sudo mkdir -p /var/lib/neon-watchdog
sudo chown neon-watchdog:neon-watchdog /var/lib/neon-watchdog

# Editar systemd service
sudo nano /etc/systemd/system/neon-watchdog.service
# Descomentar: User=neon-watchdog y Group=neon-watchdog

sudo systemctl daemon-reload
```

**Nota:** El usuario `neon-watchdog` necesitará permisos para:
- Leer PID files
- Conectar a puertos
- Ejecutar comandos/systemctl (puede requerir sudoers)

### Evitar tormentas de reinicio

Las políticas integradas previenen loops infinitos:

- `fail_threshold`: No reiniciar al primer fallo
- `restart_cooldown_seconds`: Mínimo tiempo entre reinicios
- `max_restarts_per_hour`: Rate limit absoluto

### Timeouts

Todos los checks y acciones tienen timeout configurable para evitar que el watchdog se cuelgue.

---

## 🛠️ Troubleshooting

### El watchdog no detecta el proceso

**Problema:** Check `process_name` falla pero el proceso existe.

**Solución:**
- Verifica el nombre exacto con `ps aux | grep nombre`
- `pgrep -x` busca nombre exacto (sin path)
- Considera usar `pid_file` o `tcp_port` en su lugar

### Permisos denegados en systemctl

**Problema:** `systemctl restart` falla con permiso denegado.

**Solución:**
- Ejecutar neon-watchdog como root, o
- Configurar sudoers para permitir comandos específicos sin password:

```bash
# /etc/sudoers.d/neon-watchdog
neon-watchdog ALL=(root) NOPASSWD: /bin/systemctl restart nginx.service
```

### El timer no se ejecuta

**Problema:** `systemctl status neon-watchdog.timer` muestra "inactive".

**Solución:**
```bash
sudo systemctl enable neon-watchdog.timer
sudo systemctl start neon-watchdog.timer
systemctl list-timers --all | grep neon
```

### Ver debug completo

```bash
# Modo verbose manual
neon-watchdog check -c /etc/neon-watchdog/config.yml --verbose

# O cambiar en config
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
│         Configuration Loader            │
│     (YAML/JSON + Validation)            │
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
│ - TCP port   │          └──────────────┘
│ - Command    │
└──────────────┘
```

---

## 🧪 Testing

### Test manual

1. Levanta un servidor de prueba:
```bash
python3 -m http.server 8888
```

2. Configura un target:
```yaml
- name: test-server
  enabled: true
  checks:
    - type: tcp_port
      tcp_port: "8888"
  action:
    type: exec
    exec:
      restart:
        - /bin/echo
        - "Server would restart here"
```

3. Ejecuta y mata el servidor para ver la recuperación:
```bash
neon-watchdog check -c test-config.yml --verbose
# En otra terminal: pkill -f "http.server 8888"
neon-watchdog check -c test-config.yml --verbose
```

---

## 🗺️ Roadmap

- [ ] Reload de configuración sin reinicio (SIGHUP)
- [ ] Healthcheck HTTP nativo (GET /health)
- [ ] Notificaciones (email, Telegram, Discord webhook)
- [ ] Métricas Prometheus (endpoint `/metrics`)
- [ ] Auto-instalación de systemd desde CLI
- [ ] Dashboard web (opcional)
- [ ] Soporte Docker/Podman containers
- [ ] Integración con Consul/etcd para config distribuida

---

## 📄 Licencia

MIT License - Ver [LICENSE](LICENSE) para más detalles.

---

## 🤝 Contribuir

Contribuciones son bienvenidas! Por favor:

1. Fork el proyecto
2. Crea una branch (`git checkout -b feature/amazing-feature`)
3. Commit tus cambios (`git commit -m 'Add amazing feature'`)
4. Push a la branch (`git push origin feature/amazing-feature`)
5. Abre un Pull Request

---

## 📞 Soporte

- **Issues:** [GitHub Issues](https://github.com/tgextreme/neon-watchdog/issues)
- **Discussions:** [GitHub Discussions](https://github.com/tgextreme/neon-watchdog/discussions)

---

**Hecho con ❤️ para la comunidad Linux**
