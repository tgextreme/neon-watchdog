# Changelog

Todos los cambios notables en este proyecto serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/lang/es/).

## [1.0.0] - 2026-01-09

### Añadido
- ✨ CLI completo con comandos `run`, `check`, `test-config`
- 📝 Configuración declarativa en YAML/JSON
- ✅ 4 tipos de checks: process_name, pid_file, tcp_port, command
- 🔧 2 tipos de acciones: systemd y exec
- 📊 Logging estructurado compatible con journald
- 🛡️ Sistema de políticas con fail threshold, cooldown y rate limiting
- ⏱️ Integración con systemd (timer + daemon modes)
- 💾 Persistencia de estado opcional
- 📖 Documentación completa y ejemplos
- 🏗️ Makefile para build e instalación
- 🧪 Validación completa de configuración

### Características
- Soporte para múltiples targets simultáneos
- Checks paralelos por target
- Timeouts configurables
- Rate limiting para evitar tormentas de reinicio
- Exit codes útiles para scripting
- Logs estructurados con clave=valor

## [Unreleased]

### Por añadir
- Reload de configuración sin reinicio (SIGHUP)
- Healthcheck HTTP nativo
- Notificaciones (email, Telegram, Discord)
- Métricas Prometheus
- Dashboard web
- Soporte Docker/Podman
