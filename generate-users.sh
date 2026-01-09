#!/bin/bash
# Script para generar usuarios con hashes bcrypt

# Función para generar hash bcrypt (requiere htpasswd de apache2-utils)
generate_hash() {
    password="$1"
    # Usar htpasswd si está disponible, sino generar con openssl/python
    if command -v htpasswd &> /dev/null; then
        echo "$password" | htpasswd -niB user | cut -d: -f2
    else
        # Fallback con python y bcrypt
        python3 -c "import bcrypt; print(bcrypt.hashpw(b'$password', bcrypt.gensalt()).decode())"
    fi
}

echo "# Generando archivo users.txt con hashes bcrypt"
cat > users.txt << 'EOF'
# Archivo de usuarios para Neon Watchdog Dashboard
# Formato: usuario:hash_bcrypt
# Los hashes están generados con bcrypt (costo 12)

EOF

# Usuarios de ejemplo
declare -A users=(
    ["admin"]="admin123"
    ["viewer"]="viewer123"
    ["operator"]="operator123"
    ["neon"]="watchdog2026"
)

echo "Generando hashes..."
for user in "${!users[@]}"; do
    password="${users[$user]}"
    hash=$(python3 << PYTHON
import bcrypt
password = b"$password"
hashed = bcrypt.hashpw(password, bcrypt.gensalt(rounds=10))
print(hashed.decode())
PYTHON
    )
    
    echo "# Usuario: $user / Password: $password" >> users.txt
    echo "$user:$hash" >> users.txt
    echo "" >> users.txt
    echo "✓ $user : $password"
done

echo ""
echo "✓ Archivo users.txt generado con éxito"
echo ""
echo "Usuarios disponibles:"
echo "  - admin / admin123"
echo "  - viewer / viewer123"
echo "  - operator / operator123"
echo "  - neon / watchdog2026"
