.PHONY: build install clean test help

BINARY_NAME=neon-watchdog
BUILD_DIR=./bin
INSTALL_DIR=/usr/local/bin
CONFIG_DIR=/etc/neon-watchdog
SYSTEMD_DIR=/etc/systemd/system
STATE_DIR=/var/lib/neon-watchdog

VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"

help:
	@echo "Neon Watchdog - Makefile"
	@echo ""
	@echo "Targets disponibles:"
	@echo "  make build         - Compilar el binario"
	@echo "  make install       - Instalar en el sistema (requiere sudo)"
	@echo "  make uninstall     - Desinstalar del sistema (requiere sudo)"
	@echo "  make clean         - Limpiar archivos generados"
	@echo "  make test          - Ejecutar tests"
	@echo "  make run           - Compilar y ejecutar"
	@echo "  make deps          - Descargar dependencias"
	@echo "  make fmt           - Formatear código"
	@echo "  make vet           - Verificar código"
	@echo ""

deps:
	@echo "📦 Descargando dependencias..."
	go mod download
	go mod tidy

build: deps
	@echo "🔨 Compilando $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/neon-watchdog
	@echo "✅ Binario creado en $(BUILD_DIR)/$(BINARY_NAME)"

install: build
	@echo "📦 Instalando $(BINARY_NAME)..."
	@# Copiar binario
	install -m 0755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✅ Binario instalado en $(INSTALL_DIR)/$(BINARY_NAME)"
	
	@# Crear directorio de configuración
	@if [ ! -d "$(CONFIG_DIR)" ]; then \
		mkdir -p $(CONFIG_DIR); \
		echo "✅ Directorio de configuración creado: $(CONFIG_DIR)"; \
	fi
	
	@# Copiar configuración de ejemplo si no existe
	@if [ ! -f "$(CONFIG_DIR)/config.yml" ]; then \
		install -m 0644 examples/config.yml $(CONFIG_DIR)/config.yml; \
		echo "✅ Configuración de ejemplo instalada en $(CONFIG_DIR)/config.yml"; \
		echo "⚠️  IMPORTANTE: Edita $(CONFIG_DIR)/config.yml antes de usar"; \
	else \
		echo "⏭️  Config existente en $(CONFIG_DIR)/config.yml - no se sobrescribe"; \
	fi
	
	@# Crear directorio de estado
	@if [ ! -d "$(STATE_DIR)" ]; then \
		mkdir -p $(STATE_DIR); \
		chmod 755 $(STATE_DIR); \
		echo "✅ Directorio de estado creado: $(STATE_DIR)"; \
	fi
	
	@# Instalar archivos systemd
	@install -m 0644 systemd/neon-watchdog.service $(SYSTEMD_DIR)/neon-watchdog.service
	@install -m 0644 systemd/neon-watchdog.timer $(SYSTEMD_DIR)/neon-watchdog.timer
	@install -m 0644 systemd/neon-watchdog-daemon.service $(SYSTEMD_DIR)/neon-watchdog-daemon.service
	@echo "✅ Archivos systemd instalados"
	
	@# Recargar systemd
	@systemctl daemon-reload
	@echo "✅ Systemd recargado"
	
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🎉 Instalación completada!"
	@echo ""
	@echo "Próximos pasos:"
	@echo "  1. Editar configuración:"
	@echo "     sudo nano $(CONFIG_DIR)/config.yml"
	@echo ""
	@echo "  2. Validar configuración:"
	@echo "     neon-watchdog test-config -c $(CONFIG_DIR)/config.yml"
	@echo ""
	@echo "  3. Habilitar systemd timer (recomendado):"
	@echo "     sudo systemctl enable --now neon-watchdog.timer"
	@echo ""
	@echo "  O usar modo daemon:"
	@echo "     sudo systemctl enable --now neon-watchdog-daemon.service"
	@echo ""
	@echo "  4. Ver logs:"
	@echo "     journalctl -u neon-watchdog.service -f"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

uninstall:
	@echo "🗑️  Desinstalando $(BINARY_NAME)..."
	@# Detener y deshabilitar servicios
	@-systemctl stop neon-watchdog.timer 2>/dev/null || true
	@-systemctl disable neon-watchdog.timer 2>/dev/null || true
	@-systemctl stop neon-watchdog-daemon.service 2>/dev/null || true
	@-systemctl disable neon-watchdog-daemon.service 2>/dev/null || true
	@echo "✅ Servicios detenidos"
	
	@# Eliminar archivos systemd
	@rm -f $(SYSTEMD_DIR)/neon-watchdog.service
	@rm -f $(SYSTEMD_DIR)/neon-watchdog.timer
	@rm -f $(SYSTEMD_DIR)/neon-watchdog-daemon.service
	@systemctl daemon-reload
	@echo "✅ Archivos systemd eliminados"
	
	@# Eliminar binario
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✅ Binario eliminado"
	
	@echo ""
	@echo "⚠️  Archivos NO eliminados (hazlo manualmente si lo deseas):"
	@echo "  - Configuración: $(CONFIG_DIR)/"
	@echo "  - Estado: $(STATE_DIR)/"
	@echo ""
	@echo "✅ Desinstalación completada"

clean:
	@echo "🧹 Limpiando..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "✅ Limpieza completada"

test:
	@echo "🧪 Ejecutando tests..."
	go test -v ./...

run: build
	@echo "🚀 Ejecutando $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME)

fmt:
	@echo "🎨 Formateando código..."
	go fmt ./...

vet:
	@echo "🔍 Verificando código..."
	go vet ./...

# Target de desarrollo
dev: fmt vet build
	@echo "✅ Desarrollo: formato + vet + build completados"
