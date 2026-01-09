#!/bin/bash
# 🚀 Neon Watchdog - Simple Build Script

echo "╔════════════════════════════════════════════════════════╗"
echo "║     🐺 Neon Watchdog v2.0 - Build Script             ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Limpiar
echo "🧹 Limpiando build anterior..."
rm -f neon-watchdog
echo ""

# Compilar
echo "🔨 Compilando..."
START=$(date +%s)
go build -o neon-watchdog ./cmd/neon-watchdog
if [ $? -eq 0 ]; then
    END=$(date +%s)
    DURATION=$((END - START))
    SIZE=$(du -h neon-watchdog | cut -f1)
    echo "✅ Compilación exitosa en ${DURATION}s"
    echo "📦 Tamaño: $SIZE"
else
    echo "❌ Error en compilación"
    exit 1
fi
echo ""

# Verificar
echo "🧪 Probando binario..."
if [ -x "neon-watchdog" ]; then
    echo "✅ Binario ejecutable"
    
    echo ""
    echo "📋 Version:"
    ./neon-watchdog version
    
    echo ""
    echo "⚙️  Validando config de ejemplo:"
    ./neon-watchdog test-config -c examples/config.yml
    
    if [ $? -eq 0 ]; then
        echo ""
        echo "╔════════════════════════════════════════════════════════╗"
        echo "║          ✅ TODO COMPILADO Y FUNCIONANDO              ║"
        echo "╚════════════════════════════════════════════════════════╝"
        echo ""
        echo "🚀 Comandos disponibles:"
        echo "   ./neon-watchdog version"
        echo "   ./neon-watchdog test-config -c examples/config.yml"
        echo "   ./neon-watchdog check -c examples/config.yml"
        echo "   sudo ./test-apache.sh  (test completo)"
        echo ""
    fi
else
    echo "❌ Error: binario no ejecutable"
    exit 1
fi
