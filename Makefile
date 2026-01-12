# Makefile para Neon Watchdog

# Variables
BINARY_NAME=neon-watchdog
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
BUILD_DATE=$(shell date -u '+%Y-%m-%d')
GO_FILES=$(shell find . -name '*.go' -type f)

# Rutas de instalación
PREFIX=/usr/local
BINDIR=$(PREFIX)/bin
SYSCONFDIR=/etc
SYSTEMDDIR=/etc/systemd/system
STATEDIR=/var/lib/$(BINARY_NAME)

# Flags de compilación
LDFLAGS=-ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"

.PHONY: all build install uninstall clean test fmt vet lint help

# Default target
all: build

## build: Compilar el binario
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/$(BINARY_NAME)
	@echo "Build complete: $(BINARY_NAME)"

## install: Instalar binario, configuración y servicios systemd
install: build
	@echo "Installing $(BINARY_NAME)..."
	@install -d $(DESTDIR)$(BINDIR)
	@install -m 755 $(BINARY_NAME) $(DESTDIR)$(BINDIR)/
	@install -d $(DESTDIR)$(SYSCONFDIR)/$(BINARY_NAME)
	@if [ ! -f $(DESTDIR)$(SYSCONFDIR)/$(BINARY_NAME)/config.yml ]; then \
		install -m 644 examples/config.yml $(DESTDIR)$(SYSCONFDIR)/$(BINARY_NAME)/; \
	fi
	@install -d $(DESTDIR)$(SYSTEMDDIR)
	@install -m 644 systemd/$(BINARY_NAME).service $(DESTDIR)$(SYSTEMDDIR)/
	@install -m 644 systemd/$(BINARY_NAME).timer $(DESTDIR)$(SYSTEMDDIR)/
	@install -m 644 systemd/$(BINARY_NAME)-daemon.service $(DESTDIR)$(SYSTEMDDIR)/
	@install -d $(DESTDIR)$(STATEDIR)
	@systemctl daemon-reload 2>/dev/null || true
	@echo "Installation complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit configuration: sudo nano $(SYSCONFDIR)/$(BINARY_NAME)/config.yml"
	@echo "  2. Test configuration: $(BINARY_NAME) test-config -c $(SYSCONFDIR)/$(BINARY_NAME)/config.yml"
	@echo "  3. Enable timer: sudo systemctl enable --now $(BINARY_NAME).timer"

## uninstall: Desinstalar el watchdog
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@systemctl stop $(BINARY_NAME).timer 2>/dev/null || true
	@systemctl stop $(BINARY_NAME)-daemon.service 2>/dev/null || true
	@systemctl disable $(BINARY_NAME).timer 2>/dev/null || true
	@systemctl disable $(BINARY_NAME)-daemon.service 2>/dev/null || true
	@rm -f $(DESTDIR)$(BINDIR)/$(BINARY_NAME)
	@rm -f $(DESTDIR)$(SYSTEMDDIR)/$(BINARY_NAME).service
	@rm -f $(DESTDIR)$(SYSTEMDDIR)/$(BINARY_NAME).timer
	@rm -f $(DESTDIR)$(SYSTEMDDIR)/$(BINARY_NAME)-daemon.service
	@systemctl daemon-reload 2>/dev/null || true
	@echo "Uninstallation complete!"
	@echo "Note: Configuration files in $(SYSCONFDIR)/$(BINARY_NAME)/ were preserved"
	@echo "      State files in $(STATEDIR)/ were preserved"

## clean: Limpiar binarios y archivos temporales
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f bin/$(BINARY_NAME)
	@go clean
	@echo "Clean complete!"

## test: Ejecutar tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete!"

## fmt: Formatear código Go
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete!"

## vet: Ejecutar go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete!"

## lint: Ejecutar linter (requiere golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"

## run: Compilar y ejecutar en modo check
run: build
	@./$(BINARY_NAME) check -c examples/config.yml --verbose

## dev: Ejecutar en modo daemon con configuración de ejemplo
dev: build
	@./$(BINARY_NAME) run -c examples/config.yml --verbose

## help: Mostrar ayuda
help:
	@echo "Neon Watchdog - Makefile"
	@echo ""
	@echo "Targets disponibles:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "Ejemplos:"
	@echo "  make build          # Compilar binario"
	@echo "  make install        # Instalar en el sistema"
	@echo "  make test           # Ejecutar tests"
	@echo "  make clean          # Limpiar binarios"
