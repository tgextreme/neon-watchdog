# Neon Watchdog - Comandos Útiles de Referencia Rápida

## 🔨 Desarrollo

```bash
# Compilar
make build

# Compilar + formatear + verificar
make dev

# Limpiar builds
make clean

# Ejecutar tests
make test

# Formatear código
make fmt

# Verificar código
make vet
```

## 📦 Instalación/Desinstalación

```bash
# Instalar en el sistema
sudo make install

# Desinstalar
sudo make uninstall

# Build e instalación personalizada
go build -o /custom/path/neon-watchdog ./cmd/neon-watchdog
```

## 🚀 Ejecución

```bash
# Ejecutar check una vez
neon-watchdog check -c /etc/neon-watchdog/config.yml

# Ejecutar en modo daemon
neon-watchdog run -c /etc/neon-watchdog/config.yml

# Con verbose logging
neon-watchdog check -c config.yml --verbose

# Dry-run (no ejecutar acciones)
neon-watchdog check -c config.yml --dry-run

# Validar configuración
neon-watchdog test-config -c config.yml

# Ver versión
neon-watchdog version

# Ayuda
neon-watchdog help
```

## 🔧 Systemd

### Timer Mode (Recomendado)

```bash
# Habilitar y arrancar
sudo systemctl enable --now neon-watchdog.timer

# Ver estado del timer
systemctl status neon-watchdog.timer

# Ver cuándo se ejecutará
systemctl list-timers neon-watchdog.timer

# Ver todos los timers
systemctl list-timers --all

# Detener timer
sudo systemctl stop neon-watchdog.timer

# Deshabilitar
sudo systemctl disable neon-watchdog.timer

# Reiniciar timer
sudo systemctl restart neon-watchdog.timer

# Editar timer interval
sudo systemctl edit neon-watchdog.timer
# Añadir: [Timer]
#         OnUnitActiveSec=60s
sudo systemctl daemon-reload
sudo systemctl restart neon-watchdog.timer
```

### Daemon Mode

```bash
# Habilitar y arrancar daemon
sudo systemctl enable --now neon-watchdog-daemon.service

# Ver estado
systemctl status neon-watchdog-daemon.service

# Detener
sudo systemctl stop neon-watchdog-daemon.service

# Reiniciar
sudo systemctl restart neon-watchdog-daemon.service

# Deshabilitar
sudo systemctl disable neon-watchdog-daemon.service
```

### Ambos Modos

```bash
# Recargar systemd después de cambios
sudo systemctl daemon-reload

# Ver configuración del service
systemctl cat neon-watchdog.service

# Editar service
sudo systemctl edit neon-watchdog.service

# Ver dependencias
systemctl list-dependencies neon-watchdog.service
```

## 📊 Logs

```bash
# Ver logs en tiempo real
journalctl -u neon-watchdog.service -f

# Últimas 50 líneas
journalctl -u neon-watchdog.service -n 50

# Logs desde hoy
journalctl -u neon-watchdog.service --since today

# Logs de la última hora
journalctl -u neon-watchdog.service --since "1 hour ago"

# Logs entre fechas
journalctl -u neon-watchdog.service --since "2026-01-09 10:00" --until "2026-01-09 12:00"

# Solo errores
journalctl -u neon-watchdog.service -p err

# Formato JSON
journalctl -u neon-watchdog.service -o json

# Filtrar por target específico
journalctl -u neon-watchdog.service | grep 'target=nginx'

# Ver logs con scroll
journalctl -u neon-watchdog.service --no-pager | less

# Exportar logs
journalctl -u neon-watchdog.service > logs.txt

# Borrar logs antiguos (requiere sudo)
sudo journalctl --vacuum-time=7d  # Mantener solo 7 días
sudo journalctl --vacuum-size=100M  # Mantener solo 100MB
```

## 🧪 Testing y Debugging

```bash
# Ejecutar script de test
./test.sh

# Test manual con servidor de prueba
python3 -m http.server 8888 &
neon-watchdog check -c examples/config.yml --verbose
pkill -f "http.server 8888"
neon-watchdog check -c examples/config.yml --verbose

# Verificar sintaxis de config
neon-watchdog test-config -c /etc/neon-watchdog/config.yml

# Ver qué proceso se detectaría
pgrep -x nginx
pgrep -af nginx  # Con argumentos completos

# Verificar puerto TCP
nc -zv 127.0.0.1 80
curl -v http://127.0.0.1:80

# Verificar PID file
cat /var/run/nginx.pid
ps -p $(cat /var/run/nginx.pid)

# Simular fallo de servicio
sudo systemctl stop nginx
neon-watchdog check -c /etc/neon-watchdog/config.yml --verbose

# Strace (debug avanzado)
sudo strace -f neon-watchdog check -c config.yml
```

## 📝 Configuración

```bash
# Editar config principal
sudo nano /etc/neon-watchdog/config.yml

# Validar después de editar
neon-watchdog test-config -c /etc/neon-watchdog/config.yml

# Backup de config
sudo cp /etc/neon-watchdog/config.yml /etc/neon-watchdog/config.yml.backup

# Restaurar backup
sudo cp /etc/neon-watchdog/config.yml.backup /etc/neon-watchdog/config.yml

# Ver config actual
cat /etc/neon-watchdog/config.yml

# Comparar configs
diff config1.yml config2.yml
```

## 🔐 Permisos y Seguridad

```bash
# Crear usuario dedicado
sudo useradd -r -s /bin/false -d /var/lib/neon-watchdog neon-watchdog

# Crear directorio de estado
sudo mkdir -p /var/lib/neon-watchdog
sudo chown neon-watchdog:neon-watchdog /var/lib/neon-watchdog
sudo chmod 750 /var/lib/neon-watchdog

# Dar permisos para systemctl (sudoers)
sudo visudo -f /etc/sudoers.d/neon-watchdog
# Añadir:
# neon-watchdog ALL=(root) NOPASSWD: /bin/systemctl restart nginx.service

# Ver permisos de archivos
ls -lh /usr/local/bin/neon-watchdog
ls -lh /etc/neon-watchdog/
ls -lh /var/lib/neon-watchdog/

# Verificar capabilities (si se usan)
getcap /usr/local/bin/neon-watchdog
```

## 🔍 Troubleshooting

```bash
# ¿El binario existe?
which neon-watchdog
ls -lh /usr/local/bin/neon-watchdog

# ¿El config existe?
ls -lh /etc/neon-watchdog/config.yml

# ¿Systemd units instalados?
ls -lh /etc/systemd/system/neon-watchdog.*

# ¿Timer activo?
systemctl is-active neon-watchdog.timer
systemctl is-enabled neon-watchdog.timer

# Ver estado completo de systemd
systemctl status neon-watchdog.timer --no-pager -l

# Ver últimos fallos
systemctl list-units --failed

# Verificar sintaxis de systemd unit
systemd-analyze verify /etc/systemd/system/neon-watchdog.service

# Ver qué procesos están ejecutándose
ps aux | grep neon-watchdog

# Ver puertos en uso
sudo netstat -tlnp | grep LISTEN
sudo ss -tlnp | grep LISTEN

# Verificar conectividad
ping -c 1 127.0.0.1
telnet 127.0.0.1 80

# Test de comando healthcheck
/usr/bin/curl -fsS --max-time 5 http://127.0.0.1:8080/health
echo $?  # Debe ser 0 si OK
```

## 💾 Estado y Datos

```bash
# Ver archivo de estado
cat /var/lib/neon-watchdog/state.json

# Ver con formato
jq . /var/lib/neon-watchdog/state.json

# Backup de estado
sudo cp /var/lib/neon-watchdog/state.json /tmp/state-backup.json

# Limpiar estado (resetear contadores)
sudo rm /var/lib/neon-watchdog/state.json
sudo systemctl restart neon-watchdog.timer

# Ver tamaño de logs
journalctl --disk-usage
```

## 🚀 Comandos de Producción

```bash
# Health check del watchdog mismo
systemctl is-active neon-watchdog.timer && echo "OK" || echo "FAIL"

# Ver si ha ejecutado recientemente
journalctl -u neon-watchdog.service --since "5 minutes ago" | tail

# Contar ejecuciones hoy
journalctl -u neon-watchdog.service --since today | grep -c "neon-watchdog starting"

# Ver targets con problemas
journalctl -u neon-watchdog.service --since today | grep "FAIL\|unhealthy"

# Ver acciones de recuperación ejecutadas
journalctl -u neon-watchdog.service --since today | grep "executing recovery action"

# Monitoreo continuo
watch -n 5 'systemctl status neon-watchdog.timer'

# Script de monitoreo simple
cat > /usr/local/bin/watchdog-status.sh << 'EOF'
#!/bin/bash
echo "=== Neon Watchdog Status ==="
echo ""
echo "Timer Status:"
systemctl status neon-watchdog.timer --no-pager | head -n 3
echo ""
echo "Last 5 executions:"
journalctl -u neon-watchdog.service -n 5 --no-pager
EOF
chmod +x /usr/local/bin/watchdog-status.sh
```

## 🔄 Updates y Mantenimiento

```bash
# Actualizar a nueva versión
cd neon-watchdog
git pull
make build
sudo make install
sudo systemctl daemon-reload
sudo systemctl restart neon-watchdog.timer

# Verificar versión instalada
neon-watchdog version

# Rotar logs manualmente
sudo journalctl --rotate
sudo journalctl --vacuum-time=30d
```

## 📤 Exportar/Importar Config

```bash
# Exportar configuración actual
sudo cp /etc/neon-watchdog/config.yml ~/neon-watchdog-backup.yml

# Importar a otro servidor
scp ~/neon-watchdog-backup.yml user@server:/tmp/
ssh user@server 'sudo cp /tmp/neon-watchdog-backup.yml /etc/neon-watchdog/config.yml'
```

---

**Tip:** Guarda este archivo como referencia rápida o añádelo a tus dotfiles.
