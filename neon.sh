#!/bin/bash
# Script de utilidades para Neon Watchdog

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_NAME="neon-watchdog"
CONFIG_PATH="${CONFIG_PATH:-/etc/neon-watchdog/config.yml}"
LOCAL_CONFIG_PATH="${SCRIPT_DIR}/examples/config.yml"

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Funciones de ayuda
error() {
    echo -e "${RED}✗ Error:${NC} $1" >&2
    exit 1
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

info() {
    echo -e "$1"
}

# Verificar si el binario existe
check_binary() {
    if ! command -v "$BINARY_NAME" &> /dev/null; then
        error "Binary '$BINARY_NAME' not found. Run 'make install' first."
    fi
}

# Mostrar uso
usage() {
    cat << EOF
Neon Watchdog - Utility Script

Usage: $0 [command]

Commands:
  build           Compilar el proyecto
  install         Instalar en el sistema
  uninstall       Desinstalar del sistema
  
  start           Iniciar el timer de systemd
  stop            Detener el timer de systemd
  restart         Reiniciar el timer de systemd
  status          Ver estado del servicio
  logs            Ver logs en tiempo real
  
  check           Ejecutar un check manual
  test            Validar configuración
  
  enable-daemon   Habilitar modo daemon (en lugar de timer)
  disable-daemon  Deshabilitar modo daemon
  
  clean           Limpiar archivos compilados
  help            Mostrar esta ayuda

Examples:
  $0 build          # Compilar
  $0 install        # Instalar
  $0 start          # Iniciar servicio
  $0 logs           # Ver logs
  $0 check          # Ejecutar check manual

EOF
    exit 0
}

# Comando: build
cmd_build() {
    info "Building $BINARY_NAME..."
    cd "$SCRIPT_DIR"
    make build
    success "Build complete!"
}

# Comando: install
cmd_install() {
    info "Installing $BINARY_NAME..."
    cd "$SCRIPT_DIR"
    sudo make install
    success "Installation complete!"
}

# Comando: uninstall
cmd_uninstall() {
    warning "This will uninstall $BINARY_NAME from your system."
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cd "$SCRIPT_DIR"
        sudo make uninstall
        success "Uninstallation complete!"
    else
        info "Cancelled."
    fi
}

# Comando: start
cmd_start() {
    info "Starting $BINARY_NAME timer..."
    sudo systemctl enable --now "$BINARY_NAME.timer"
    success "Timer started!"
    systemctl status "$BINARY_NAME.timer" --no-pager
}

# Comando: stop
cmd_stop() {
    info "Stopping $BINARY_NAME timer..."
    sudo systemctl stop "$BINARY_NAME.timer"
    success "Timer stopped!"
}

# Comando: restart
cmd_restart() {
    info "Restarting $BINARY_NAME timer..."
    sudo systemctl restart "$BINARY_NAME.timer"
    success "Timer restarted!"
    systemctl status "$BINARY_NAME.timer" --no-pager
}

# Comando: status
cmd_status() {
    info "Service status:"
    echo ""
    systemctl status "$BINARY_NAME.timer" --no-pager || true
    echo ""
    info "Next execution:"
    systemctl list-timers "$BINARY_NAME.timer" --no-pager || true
}

# Comando: logs
cmd_logs() {
    info "Showing logs (Ctrl+C to exit)..."
    journalctl -u "$BINARY_NAME.service" -f
}

# Comando: check
cmd_check() {
    check_binary
    
    # Determinar qué configuración usar
    if [ -f "$CONFIG_PATH" ]; then
        CONFIG="$CONFIG_PATH"
    elif [ -f "$LOCAL_CONFIG_PATH" ]; then
        CONFIG="$LOCAL_CONFIG_PATH"
    else
        error "Configuration file not found. Tried: $CONFIG_PATH and $LOCAL_CONFIG_PATH"
    fi
    
    info "Running check with config: $CONFIG"
    "$BINARY_NAME" check -c "$CONFIG" --verbose
}

# Comando: test
cmd_test() {
    check_binary
    
    # Determinar qué configuración usar
    if [ -f "$CONFIG_PATH" ]; then
        CONFIG="$CONFIG_PATH"
    elif [ -f "$LOCAL_CONFIG_PATH" ]; then
        CONFIG="$LOCAL_CONFIG_PATH"
    else
        error "Configuration file not found. Tried: $CONFIG_PATH and $LOCAL_CONFIG_PATH"
    fi
    
    info "Testing configuration: $CONFIG"
    "$BINARY_NAME" test-config -c "$CONFIG"
}

# Comando: enable-daemon
cmd_enable_daemon() {
    info "Switching to daemon mode..."
    sudo systemctl disable --now "$BINARY_NAME.timer" 2>/dev/null || true
    sudo systemctl enable --now "$BINARY_NAME-daemon.service"
    success "Daemon mode enabled!"
    systemctl status "$BINARY_NAME-daemon.service" --no-pager
}

# Comando: disable-daemon
cmd_disable_daemon() {
    info "Switching to timer mode..."
    sudo systemctl disable --now "$BINARY_NAME-daemon.service" 2>/dev/null || true
    sudo systemctl enable --now "$BINARY_NAME.timer"
    success "Timer mode enabled!"
    systemctl status "$BINARY_NAME.timer" --no-pager
}

# Comando: clean
cmd_clean() {
    info "Cleaning build artifacts..."
    cd "$SCRIPT_DIR"
    make clean
    success "Clean complete!"
}

# Main
case "${1:-}" in
    build)
        cmd_build
        ;;
    install)
        cmd_install
        ;;
    uninstall)
        cmd_uninstall
        ;;
    start)
        cmd_start
        ;;
    stop)
        cmd_stop
        ;;
    restart)
        cmd_restart
        ;;
    status)
        cmd_status
        ;;
    logs)
        cmd_logs
        ;;
    check)
        cmd_check
        ;;
    test)
        cmd_test
        ;;
    enable-daemon)
        cmd_enable_daemon
        ;;
    disable-daemon)
        cmd_disable_daemon
        ;;
    clean)
        cmd_clean
        ;;
    help|--help|-h|"")
        usage
        ;;
    *)
        error "Unknown command: $1\nRun '$0 help' for usage."
        ;;
esac
