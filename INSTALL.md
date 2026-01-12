# Guía de Instalación - Neon Watchdog

Esta guía proporciona instrucciones detalladas para instalar y configurar Neon Watchdog en tu sistema Linux.

## 📋 Requisitos Previos

### Sistema Operativo

- Linux (probado en Ubuntu 20.04+, Debian 10+, CentOS 8+, RHEL 8+)
- systemd (para integración con el sistema)

### Software

- **Go 1.21 o superior** (solo para compilar desde código fuente)
- **git** (para clonar el repositorio)
- **make** (opcional, facilita la compilación)

### Permisos

- Acceso sudo para instalar en rutas del sistema
- Permisos para ejecutar `systemctl` (o configurar sudoers)

## 🚀 Método 1: Compilar desde Código Fuente (Recomendado)

### Paso 1: Clonar el Repositorio

```bash
git clone https://github.com/tgextreme/neon-watchdog.git
cd neon-watchdog
```

### Paso 2: Compilar

```bash
# Opción A: Usando Make (recomendado)
make build

# Opción B: Compilar manualmente con Go
go build -o neon-watchdog -ldflags="-s -w" ./cmd/neon-watchdog
```

### Paso 3: Instalar

```bash
# Instalar con Make (incluye binario, config y systemd)
sudo make install

# O instalar manualmente
sudo cp neon-watchdog /usr/local/bin/
sudo mkdir -p /etc/neon-watchdog
sudo cp examples/config.yml /etc/neon-watchdog/
sudo cp systemd/*.service systemd/*.timer /etc/systemd/system/
sudo systemctl daemon-reload
```

### Paso 4: Verificar Instalación

```bash
# Verificar que el binario está instalado
neon-watchdog version

# Verificar archivos systemd
systemctl list-unit-files | grep neon-watchdog
```

## 🐳 Método 2: Instalación con Binario Pre-compilado

```bash
# Descargar binario (cuando esté disponible en releases)
wget https://github.com/tgextreme/neon-watchdog/releases/download/v1.0.0/neon-watchdog-linux-amd64

# Hacer ejecutable
chmod +x neon-watchdog-linux-amd64

# Mover a PATH
sudo mv neon-watchdog-linux-amd64 /usr/local/bin/neon-watchdog

# Crear directorio de configuración
sudo mkdir -p /etc/neon-watchdog

# Descargar configuración de ejemplo
sudo wget -O /etc/neon-watchdog/config.yml \
  https://raw.githubusercontent.com/tgextreme/neon-watchdog/main/examples/config.yml
```

## ⚙️ Configuración Inicial

### Paso 1: Editar Configuración

Edita el archivo de configuración principal:

```bash
sudo nano /etc/neon-watchdog/config.yml
```

### Configuración Mínima

```yaml
# Configuración global
log_level: INFO                    # DEBUG, INFO, WARN, ERROR
timeout_seconds: 10                # Timeout para checks y acciones
state_file: /var/lib/neon-watchdog/state.json

# Política por defecto (puede sobrescribirse por target)
default_policy:
  fail_threshold: 1                # Fallos consecutivos antes de actuar
  restart_cooldown_seconds: 60     # Tiempo mínimo entre reinicios
  max_restarts_per_hour: 10        # Límite de reinicios por hora

# Servicios a monitorizar
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

### Paso 2: Crear Directorios Necesarios

```bash
# Directorio para estado persistente
sudo mkdir -p /var/lib/neon-watchdog
sudo chmod 755 /var/lib/neon-watchdog

# Directorio para logs (si usas archivos)
sudo mkdir -p /var/log/neon-watchdog
sudo chmod 755 /var/log/neon-watchdog
```

### Paso 3: Validar Configuración

```bash
# Probar que la configuración es válida
neon-watchdog test-config -c /etc/neon-watchdog/config.yml

# Ejecutar un check manual para probar
neon-watchdog check -c /etc/neon-watchdog/config.yml --verbose
```

## 🔧 Configuración de systemd

### Opción A: Modo Timer (Recomendado)

El modo timer ejecuta checks periódicamente usando systemd timers:

```bash
# Habilitar y arrancar el timer
sudo systemctl enable neon-watchdog.timer
sudo systemctl start neon-watchdog.timer

# Verificar estado
systemctl status neon-watchdog.timer

# Ver cuándo se ejecutará el próximo check
systemctl list-timers neon-watchdog.timer
```

#### Personalizar Intervalo del Timer

Edita `/etc/systemd/system/neon-watchdog.timer`:

```ini
[Unit]
Description=Neon Watchdog Check Timer
Requires=neon-watchdog.service

[Timer]
OnBootSec=1min           # Ejecutar 1 minuto después de boot
OnUnitActiveSec=5min     # Ejecutar cada 5 minutos
AccuracySec=30s          # Precisión de 30 segundos

[Install]
WantedBy=timers.target
```

Después de editar:

```bash
sudo systemctl daemon-reload
sudo systemctl restart neon-watchdog.timer
```

### Opción B: Modo Daemon

El modo daemon ejecuta checks continuamente en un loop:

```bash
# Deshabilitar timer si está activo
sudo systemctl disable --now neon-watchdog.timer

# Habilitar daemon
sudo systemctl enable neon-watchdog-daemon.service
sudo systemctl start neon-watchdog-daemon.service

# Verificar estado
systemctl status neon-watchdog-daemon.service
```

#### Configurar Intervalo en Daemon

Edita tu `config.yml`:

```yaml
interval_seconds: 300  # Ejecutar checks cada 5 minutos
```

## 🔐 Configuración de Seguridad

### Ejecutar como Usuario Dedicado

Para mayor seguridad, crea un usuario específico:

```bash
# Crear usuario del sistema
sudo useradd -r -s /bin/false -d /var/lib/neon-watchdog neon-watchdog

# Asignar permisos
sudo chown -R neon-watchdog:neon-watchdog /var/lib/neon-watchdog
sudo chown -R neon-watchdog:neon-watchdog /var/log/neon-watchdog

# Editar el servicio systemd
sudo nano /etc/systemd/system/neon-watchdog.service
```

Descomentar líneas:

```ini
User=neon-watchdog
Group=neon-watchdog
```

Recargar y reiniciar:

```bash
sudo systemctl daemon-reload
sudo systemctl restart neon-watchdog.timer
```

### Configurar Permisos Sudo

Si el usuario `neon-watchdog` necesita ejecutar systemctl:

```bash
# Crear archivo sudoers
sudo visudo -f /etc/sudoers.d/neon-watchdog
```

Añadir:

```
# Permitir reiniciar servicios específicos sin password
neon-watchdog ALL=(root) NOPASSWD: /bin/systemctl restart nginx.service
neon-watchdog ALL=(root) NOPASSWD: /bin/systemctl restart apache2.service
neon-watchdog ALL=(root) NOPASSWD: /bin/systemctl start *
neon-watchdog ALL=(root) NOPASSWD: /bin/systemctl stop *
```

## 📊 Habilitar Dashboard Web (Opcional)

### Paso 1: Configurar Dashboard

Añade a tu `config.yml`:

```yaml
dashboard:
  enabled: true
  port: 8080
  path: "/"
```

### Paso 2: Crear Usuarios

```bash
# Instalar htpasswd si no está disponible
sudo apt-get install apache2-utils  # Debian/Ubuntu
sudo yum install httpd-tools         # CentOS/RHEL

# Crear usuario admin
htpasswd -B -c users.txt admin

# Añadir más usuarios
htpasswd -B users.txt usuario2
```

### Paso 3: Reiniciar Servicio

```bash
sudo systemctl restart neon-watchdog.timer
# o
sudo systemctl restart neon-watchdog-daemon.service
```

### Paso 4: Acceder al Dashboard

Abre en tu navegador: `http://localhost:8080/`

Usuario: `admin`, Password: el que configuraste

## 📈 Habilitar Métricas Prometheus (Opcional)

Añade a tu `config.yml`:

```yaml
metrics:
  enabled: true
  port: 9090
  path: /metrics
```

Luego configura Prometheus para scraping:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'neon-watchdog'
    static_configs:
      - targets: ['localhost:9090']
```

## 🔔 Configurar Notificaciones (Opcional)

### Email

```yaml
notifications:
  - type: email
    enabled: true
    email:
      smtp_host: smtp.gmail.com
      smtp_port: 587
      username: alert@example.com
      password: ${SMTP_PASSWORD}  # Usar variable de entorno
      from: neon-watchdog@example.com
      to:
        - admin@example.com
      use_tls: true
```

### Webhook (Slack, Discord, etc.)

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

## 📝 Ver Logs

### Logs con journalctl

```bash
# Ver logs en tiempo real
journalctl -u neon-watchdog.service -f

# Últimas 100 líneas
journalctl -u neon-watchdog.service -n 100

# Logs desde hace 1 hora
journalctl -u neon-watchdog.service --since "1 hour ago"

# Filtrar por nivel de error
journalctl -u neon-watchdog.service -p err

# Filtrar por target específico
journalctl -u neon-watchdog.service | grep 'target=nginx'
```

### Configurar Rotación de Logs

Si usas archivos de logs en lugar de journald:

```bash
# Crear configuración de logrotate
sudo nano /etc/logrotate.d/neon-watchdog
```

```
/var/log/neon-watchdog/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0644 neon-watchdog neon-watchdog
    postrotate
        systemctl reload neon-watchdog-daemon.service > /dev/null 2>&1 || true
    endscript
}
```

## 🧪 Verificación Post-Instalación

### Checklist de Verificación

```bash
# 1. Binario instalado
which neon-watchdog

# 2. Versión correcta
neon-watchdog version

# 3. Configuración válida
neon-watchdog test-config -c /etc/neon-watchdog/config.yml

# 4. Systemd timer activo
systemctl is-active neon-watchdog.timer

# 5. Ver próxima ejecución
systemctl list-timers neon-watchdog.timer

# 6. Ejecutar check manual
neon-watchdog check -c /etc/neon-watchdog/config.yml --verbose

# 7. Ver logs
journalctl -u neon-watchdog.service -n 20
```

### Test Completo

```bash
# Script de test completo
cat > /tmp/test-neon.sh << 'EOF'
#!/bin/bash
echo "=== Neon Watchdog Installation Test ==="
echo ""
echo "1. Binary check..."
neon-watchdog version || exit 1
echo "✓ Binary OK"
echo ""
echo "2. Config check..."
neon-watchdog test-config -c /etc/neon-watchdog/config.yml || exit 1
echo "✓ Config OK"
echo ""
echo "3. Manual execution..."
neon-watchdog check -c /etc/neon-watchdog/config.yml --verbose || exit 1
echo "✓ Execution OK"
echo ""
echo "4. Systemd check..."
systemctl is-enabled neon-watchdog.timer >/dev/null 2>&1 || {
    echo "⚠ Timer not enabled"
    exit 1
}
echo "✓ Systemd OK"
echo ""
echo "=== All tests passed! ==="
EOF

chmod +x /tmp/test-neon.sh
/tmp/test-neon.sh
```

## 🛠️ Troubleshooting

### Binario no encontrado

```bash
# Verificar PATH
echo $PATH

# Buscar binario
which neon-watchdog
sudo find / -name neon-watchdog 2>/dev/null

# Reinstalar
cd neon-watchdog
sudo make install
```

### Timer no se ejecuta

```bash
# Verificar estado
systemctl status neon-watchdog.timer

# Ver errores
journalctl -u neon-watchdog.timer -n 50

# Reiniciar timer
sudo systemctl restart neon-watchdog.timer

# Verificar próxima ejecución
systemctl list-timers --all | grep neon
```

### Permisos denegados

```bash
# Verificar permisos del binario
ls -l /usr/local/bin/neon-watchdog

# Verificar permisos de config
ls -l /etc/neon-watchdog/config.yml

# Verificar logs
journalctl -u neon-watchdog.service | grep -i permission
```

### Error al reiniciar servicios

```bash
# Probar systemctl manualmente
sudo systemctl restart nginx.service

# Verificar usuario que ejecuta neon-watchdog
ps aux | grep neon-watchdog

# Configurar sudoers si es necesario
sudo visudo -f /etc/sudoers.d/neon-watchdog
```

## 🗑️ Desinstalación

```bash
# Detener y deshabilitar servicios
sudo systemctl stop neon-watchdog.timer
sudo systemctl stop neon-watchdog-daemon.service
sudo systemctl disable neon-watchdog.timer
sudo systemctl disable neon-watchdog-daemon.service

# Eliminar archivos systemd
sudo rm /etc/systemd/system/neon-watchdog.service
sudo rm /etc/systemd/system/neon-watchdog.timer
sudo rm /etc/systemd/system/neon-watchdog-daemon.service
sudo systemctl daemon-reload

# Eliminar binario
sudo rm /usr/local/bin/neon-watchdog

# Eliminar configuración (opcional)
sudo rm -rf /etc/neon-watchdog

# Eliminar datos (opcional)
sudo rm -rf /var/lib/neon-watchdog
sudo rm -rf /var/log/neon-watchdog

# Eliminar usuario (opcional)
sudo userdel neon-watchdog
```

O usar Make:

```bash
cd neon-watchdog
sudo make uninstall
```

## 📚 Próximos Pasos

1. **Configurar targets**: Añade los servicios que quieres monitorizar en `config.yml`
2. **Leer documentación completa**: Ver [README.md](README.md)
3. **Explorar API REST**: Ver [API-REST.md](API-REST.md)
4. **Configurar notificaciones**: Añade email/webhook/telegram
5. **Habilitar métricas**: Integra con Prometheus
6. **Revisar logs**: Familiarízate con el output de journalctl

## 📮 Soporte

- **Documentación**: [README.md](README.md)
- **Issues**: https://github.com/tgextreme/neon-watchdog/issues
- **Ejemplos**: Ver carpeta `examples/` en el repositorio
