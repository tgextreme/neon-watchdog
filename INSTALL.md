# Instalación Rápida de Neon Watchdog

Este documento proporciona instrucciones concisas para poner en marcha Neon Watchdog en minutos.

## Requisitos Previos

- Linux (probado en Ubuntu 20.04+, Debian 10+, CentOS 8+)
- Go 1.21 o superior (solo para compilar)
- systemd
- Permisos sudo

## Instalación en 3 Pasos

### 1. Compilar e Instalar

```bash
# Clonar o descargar el proyecto
cd neon-watchdog

# Compilar e instalar (requiere sudo)
make build
sudo make install
```

### 2. Configurar

Edita `/etc/neon-watchdog/config.yml` con tus targets:

```bash
sudo nano /etc/neon-watchdog/config.yml
```

Ejemplo mínimo:
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
    action:
      type: systemd
      systemd:
        unit: nginx.service
        method: restart
```

### 3. Activar

```bash
# Validar configuración
neon-watchdog test-config -c /etc/neon-watchdog/config.yml

# Habilitar y arrancar (modo timer recomendado)
sudo systemctl enable --now neon-watchdog.timer

# Ver logs
journalctl -u neon-watchdog.service -f
```

## Verificación

```bash
# Estado del timer
systemctl status neon-watchdog.timer

# Ver cuándo se ejecutará el próximo check
systemctl list-timers neon-watchdog.timer

# Ver última ejecución
journalctl -u neon-watchdog.service -n 50
```

## Troubleshooting Rápido

**El timer no se ejecuta:**
```bash
sudo systemctl start neon-watchdog.timer
systemctl list-timers --all | grep neon
```

**Ver más detalles en logs:**
```bash
# Cambiar log_level a DEBUG en config.yml
sudo nano /etc/neon-watchdog/config.yml
# Reiniciar timer
sudo systemctl restart neon-watchdog.timer
```

**Probar manualmente:**
```bash
neon-watchdog check -c /etc/neon-watchdog/config.yml --verbose
```

## Alternativa: Modo Daemon

Si prefieres modo daemon en lugar de timer:

```bash
# Deshabilitar timer
sudo systemctl disable --now neon-watchdog.timer

# Habilitar daemon
sudo systemctl enable --now neon-watchdog-daemon.service

# Ver estado
sudo systemctl status neon-watchdog-daemon.service
```

## Desinstalación

```bash
cd neon-watchdog
sudo make uninstall
```

## Más Información

- Documentación completa: [README.md](README.md)
- Ejemplos de configuración: [examples/config.yml](examples/config.yml)
- Testing local: Ejecuta `./test.sh`

## Ayuda

```bash
neon-watchdog help
```
