# Sistema de Autenticación - Neon Watchdog Dashboard

## 📋 Descripción

El dashboard utiliza un sistema de autenticación tipo **WordPress** con usuarios almacenados en archivo de texto y contraseñas hasheadas con **bcrypt**.

## 🔐 Características

- ✅ Autenticación HTTP Basic Auth
- ✅ Contraseñas hasheadas con bcrypt (costo 10)
- ✅ Usuarios almacenados en `users.txt`
- ✅ No requiere permisos de sistema
- ✅ Fácil gestión de usuarios
- ✅ Compatible con cualquier navegador

## 👥 Usuarios Predefinidos

| Usuario   | Contraseña     | Descripción           |
|-----------|----------------|-----------------------|
| `admin`   | `admin123`     | Administrador         |
| `viewer`  | `viewer123`    | Visualizador          |
| `operator`| `operator123`  | Operador              |
| `neon`    | `watchdog2026` | Usuario Neon          |

## 📁 Archivo users.txt

### Formato

```
# Comentarios con #
usuario:$2a$10$hash_bcrypt_completo
```

### Ejemplo

```txt
# Usuario: admin / Password: admin123
admin:$2a$10$N9qo8uLOickgx2ZMRZoMyeIjreAyn2Lh8H58I9gKmYFwzQq7LXhRW

# Usuario: viewer / Password: viewer123  
viewer:$2a$10$rVXQE.xJz5vQVzKXfJ5OGuPZ8qJ5XqZJ5yxJ5xJ5xJ5xJ5xJ5xJ5O
```

## 🔧 Gestión de Usuarios

### Generar Nuevos Usuarios

#### Opción 1: Usar el script incluido

```bash
./generate-users.sh
```

Este script:
- Genera 4 usuarios de ejemplo
- Crea hashes bcrypt seguros
- Actualiza `users.txt` automáticamente

#### Opción 2: Generar hash manualmente con Python

```bash
python3 << 'EOF'
import bcrypt
password = b"mi_password"
hashed = bcrypt.hashpw(password, bcrypt.gensalt(rounds=10))
print(f"Hash: {hashed.decode()}")
EOF
```

#### Opción 3: Usar htpasswd (Apache)

```bash
# Instalar si no está disponible
sudo apt install apache2-utils

# Generar hash
htpasswd -nbB usuario password
```

### Añadir Usuario Manualmente

1. Genera el hash de la contraseña
2. Edita `users.txt`
3. Añade línea: `usuario:$2a$10$hash...`
4. Guarda el archivo
5. El cambio es inmediato (no requiere reinicio)

### Eliminar Usuario

1. Edita `users.txt`
2. Elimina o comenta la línea del usuario
3. Guarda el archivo

### Cambiar Contraseña

1. Genera nuevo hash para la nueva contraseña
2. Reemplaza el hash del usuario en `users.txt`
3. Guarda el archivo

## 🌐 Acceso al Dashboard

### 1. Navegador Web

Visita: `http://localhost:8080`

Se mostrará el diálogo de autenticación HTTP Basic:
- **Usuario**: `admin`
- **Contraseña**: `admin123`

### 2. API REST con curl

```bash
# Con autenticación
curl -u admin:admin123 http://localhost:8080/api/status

# Ver targets
curl -u admin:admin123 http://localhost:8080/api/targets

# Añadir target
curl -u admin:admin123 -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "apache",
    "type": "systemd",
    "target": "apache2.service",
    "enabled": true
  }'
```

### 3. Cerrar Sesión

Para cerrar sesión en el navegador:
- **Firefox/Chrome**: Cierra el navegador completamente
- **O usa**: `http://logout@localhost:8080` (algunos navegadores)

## 🔒 Seguridad

### Nivel de Seguridad

- ✅ **Bcrypt**: Algoritmo diseñado para passwords (resistente a rainbow tables)
- ✅ **Salt automático**: Cada hash tiene salt único
- ✅ **Costo 10**: ~100ms por hash (balance seguridad/rendimiento)
- ✅ **HTTP Basic**: Estándar ampliamente soportado

### Recomendaciones

#### Para Desarrollo (Local)
- ✅ HTTP está bien para `localhost`
- ✅ Usa contraseñas de ejemplo

#### Para Producción
- ⚠️ **OBLIGATORIO**: Usar HTTPS (proxy reverso nginx/caddy)
- ⚠️ Cambiar TODAS las contraseñas predefinidas
- ⚠️ Usar contraseñas fuertes (12+ caracteres)
- ⚠️ Configurar firewall (solo IPs permitidas)
- ⚠️ Considerar autenticación adicional (2FA, VPN)

### Configurar HTTPS con Nginx

```nginx
server {
    listen 443 ssl http2;
    server_name watchdog.ejemplo.com;
    
    ssl_certificate /etc/ssl/certs/watchdog.crt;
    ssl_certificate_key /etc/ssl/private/watchdog.key;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Authorization $http_authorization;
        proxy_pass_header Authorization;
    }
}
```

## 🐛 Troubleshooting

### Usuario/Contraseña Incorrectos

**Síntoma**: "401 - Autenticación Requerida"

**Solución**:
1. Verifica usuario en `users.txt`
2. Revisa que el hash sea válido
3. Prueba regenerar el hash
4. Verifica que no haya espacios extra

### Archivo users.txt No Encontrado

**Síntoma**: Error en logs "failed to open users file"

**Solución**:
```bash
# Verificar ubicación
pwd  # Debe ser el directorio del proyecto

# Verificar que existe
ls -la users.txt

# Regenerar si es necesario
./generate-users.sh
```

### No Aparece Diálogo de Login

**Síntoma**: Página carga sin pedir credenciales

**Solución**:
1. Limpia caché del navegador
2. Usa modo incógnito
3. Cierra completamente el navegador
4. Verifica logs del dashboard

### Logs del Dashboard

```bash
# Ver logs en tiempo real
tail -f neon-watchdog.log

# Buscar errores de autenticación
grep -i "auth" neon-watchdog.log
```

## 📊 Ejemplo Completo

```bash
# 1. Generar usuarios
./generate-users.sh

# 2. Verificar usuarios creados
cat users.txt

# 3. Iniciar dashboard
./neon-watchdog run -c examples/config-dashboard.yml

# 4. Abrir navegador
xdg-open http://localhost:8080

# 5. Login con: admin / admin123

# 6. Probar API
curl -u admin:admin123 http://localhost:8080/api/status
```

## 🔐 Mejores Prácticas

### DO ✅

- Usa contraseñas únicas por usuario
- Cambia contraseñas regularmente (producción)
- Mantén `users.txt` fuera de control de versiones en producción
- Usa HTTPS en producción
- Limita acceso por firewall
- Audita logs de autenticación

### DON'T ❌

- No uses HTTP en producción
- No compartas usuarios entre personas
- No versiones `users.txt` con contraseñas reales
- No uses las contraseñas de ejemplo en producción
- No expongas puerto 8080 directamente a internet

## 📝 Notas

- Los cambios en `users.txt` son **inmediatos** (no requiere reinicio)
- El dashboard lee el archivo en cada intento de autenticación
- Los hashes bcrypt son seguros para almacenar
- HTTP Basic Auth envía credenciales en cada request
- Por eso es CRÍTICO usar HTTPS en producción
