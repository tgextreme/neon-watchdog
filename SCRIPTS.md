# ✅ Neon Watchdog v2.0 - Scripts Disponibles

## 🚀 Scripts de Compilación y Testing

### 1. **run-all.sh** 🔥 (VERIFICACIÓN COMPLETA)
Script maestro que compila y verifica TODO el proyecto en 10 pasos.

```bash
chmod +x run-all.sh
./run-all.sh
```

**Qué hace:**
- ✅ Verifica dependencias (Go, Git)
- ✅ Limpia builds anteriores
- ✅ Sincroniza módulos de Go
- ✅ Compila el proyecto
- ✅ Verifica binario funcional
- ✅ Valida todas las configuraciones
- ✅ Verifica módulos v2.0 (notificaciones, métricas, dashboard, historial)
- ✅ Ejecuta check de ejemplo
- ✅ Verifica documentación
- ✅ Muestra resumen completo con estadísticas
- ⏱️ Duración: ~25-30 segundos

---

### 2. **build.sh** ⭐ (RÁPIDO)
Script simple y rápido para compilar y verificar el proyecto.

```bash
chmod +x build.sh
./build.sh
```

**Qué hace:**
- ✅ Limpia build anterior
- ✅ Compila el proyecto
- ✅ Verifica que el binario funciona
- ✅ Valida la configuración de ejemplo
- ⏱️ Duración: ~1-2 segundos

---

### 3. **test-apache.sh**
Script para probar el watchdog con Apache2 (test funcional completo).

```bash
sudo ./test-apache.sh
```

**Qué hace:**
- ✅ Instala Apache2 si no existe
- ✅ Crea configuración de test
- ✅ Verifica Apache healthy
- ✅ Simula fallo (detiene Apache)
- ✅ Watchdog detecta y recupera
- ✅ Verifica recuperación exitosa
- 📋 Genera log con timestamp

---

### 4. **build-and-test.sh**
Script completo con verificación exhaustiva (avanzado).

```bash
chmod +x build-and-test.sh
./build-and-test.sh
```

**Qué hace:**
- ✅ Verifica dependencias (Go, Git)
- ✅ Valida estructura del proyecto
- ✅ Compila con tiempos
- ✅ Verifica todos los módulos v2.0
- ✅ Valida configuraciones
- ✅ Prueba todos los comandos
- ✅ Analiza binario (tamaño, símbolos)
- ✅ Verifica features implementadas
- 📋 Log detallado guardado

---

### 5. **verify-v2.sh**
Script de verificación de features v2.0.

```bash
chmod +x verify-v2.sh
./verify-v2.sh
```

**Qué hace:**
- ✅ Verifica módulos TIER 1, 2 y 3
- ✅ Comprueba documentación
- ✅ Valida compilación

---

## 📋 Comandos del Binario

Una vez compilado con `./build.sh`, puedes usar:

### Verificar versión
```bash
./neon-watchdog version
```

### Validar configuración
```bash
./neon-watchdog test-config -c examples/config.yml
./neon-watchdog test-config -c examples/config-v2-full.yml
```

### Ejecutar un check (una vez)
```bash
./neon-watchdog check -c examples/config.yml
```

### Ejecutar en modo daemon
```bash
./neon-watchdog run -c examples/config.yml
```

### Ver ayuda
```bash
./neon-watchdog help
```

---

## 🔧 Desarrollo

### Compilar manualmente
```bash
go build -o neon-watchdog ./cmd/neon-watchdog
```

### Limpiar
```bash
rm -f neon-watchdog
go clean
```

### Actualizar dependencias
```bash
go mod tidy
go mod download
```

### Ver módulos
```bash
go list -m all
```

---

## 🎯 Workflow Típico

### 1. Verificación completa (RECOMENDADO)
```bash
./run-all.sh
```

### 2. Compilación rápida (desarrollo)
```bash
./build.sh
```

### 3. Test funcional completo
```bash
sudo ./test-apache.sh
```

### 3. Instalar en el sistema
```bash
sudo make install
```

### 4. Habilitar servicio
```bash
sudo systemctl enable --now neon-watchdog.timer
```

### 5. Ver logs
```bash
journalctl -u neon-watchdog -f
```

---

## ✅ Verificación Rápida

```bash
# Verificación completa (recomendado primera vez)
./run-all.sh

# O compilación rápida para desarrollo
./build.sh

# Si todo OK con run-all.sh, deberías ver:
# ✅ 10 pasos completados
# ✅ Binario: 9.2MB
# ✅ TODO COMPILADO, VERIFICADO Y FUNCIONANDO PERFECTAMENTE
```

---

## 📊 Resumen de Scripts

| Script | Tiempo | Complejidad | Uso |
|--------|--------|-------------|-----|
| **run-all.sh** | ~25s | Completo | Verificación total (RECOMENDADO primera vez) ⭐ |
| **build.sh** | ~1s | Básico | Desarrollo diario rápido |
| test-apache.sh | ~10s | Medio | Test funcional |
| build-and-test.sh | ~5s | Avanzado | CI/CD, validación completa |
| verify-v2.sh | ~3s | Medio | Verificar features |

---

## 🐛 Troubleshooting

### Error: "go: command not found"
```bash
# Instalar Go
sudo apt install golang-go  # Debian/Ubuntu
sudo dnf install golang      # Fedora/RHEL
```

### Error de compilación
```bash
# Limpiar y reintentar
go clean -cache
go mod tidy
./build.sh
```

### Binario muy grande
```bash
# Compilar con optimizaciones
go build -ldflags="-s -w" -o neon-watchdog ./cmd/neon-watchdog
```

---

## 💡 Tips

1. **Usa `run-all.sh` la primera vez o después de cambios grandes** - Verifica todo en 10 pasos
2. **Usa `build.sh` para desarrollo rápido** - Solo compila y valida básico
3. **Usa `test-apache.sh` antes de commit** - Asegura que funciona end-to-end
4. **El log de test-apache.sh** tiene timestamp único - Útil para debugging
5. **Los scripts son idempotentes** - Puedes ejecutarlos múltiples veces
6. **`run-all.sh` muestra estadísticas detalladas** - Incluyendo líneas de código de cada módulo

---

## 🎉 Todo Listo

Si `./run-all.sh` termina con:
```
✅ TODO COMPILADO, VERIFICADO Y FUNCIONANDO PERFECTAMENTE
```

¡Entonces todo está perfecto! 🚀

**El script verifica:**
- ✅ Dependencias instaladas
- ✅ Módulos de Go sincronizados
- ✅ Compilación exitosa
- ✅ Binario funcional
- ✅ Configuraciones válidas
- ✅ Todos los módulos v2.0 presentes
- ✅ Documentación completa
- ✅ Comandos funcionando
