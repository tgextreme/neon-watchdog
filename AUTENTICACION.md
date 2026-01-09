# 🔐 Autenticación del Dashboard con Usuarios del Sistema

## ✅ Implementado

El dashboard web ahora requiere autenticación usando **los usuarios del sistema operativo Linux**.

## 🔑 Cómo Funciona

### Autenticación HTTP Basic
- **Protocolo**: HTTP Basic Authentication
- **Validación**: Contra usuarios del sistema (/etc/passwd + PAM)
- **Método**: Usa el comando `su` para validar credenciales

### Flujo de Autenticación

1. Usuario accede a `http://localhost:8080`
2. Navegador muestra diálogo de autenticación
3. Usuario ingresa:
   - **Usuario**: Su nombre de usuario del sistema (ej: `usuario`)
   - **Contraseña**: Su contraseña del sistema
4. El servidor valida usando `su username -c true`
5. Si es correcto, se permite el acceso
6. Si falla, se muestra error 401

## 🚀 Uso

### Acceder al Dashboard

```bash
# 1. Ejecutar watchdog con dashboard habilitado
./neon-watchdog run -c examples/config-dashboard.yml

# 2. Abrir navegador
firefox http://localhost:8080

# 3. Login con tus credenciales del sistema
Usuario: usuario
Contraseña: tu_password_del_sistema
```

### Usuarios Válidos

Cualquier usuario del sistema con contraseña puede autenticarse:

```bash
# Ver usuarios del sistema
cat /etc/passwd | grep -v nologin | grep -v false

# Crear un usuario específico para el dashboard (opcional)
sudo useradd -m -s /bin/bash watchdog-admin
sudo passwd watchdog-admin
```

## 🔒 Seguridad

### ✅ Ventajas

1. **No hay credenciales hardcodeadas** - Usa usuarios del sistema
2. **Gestión centralizada** - Los admins ya conocen cómo gestionar usuarios Linux
3. **Auditable** - Los logs muestran qué usuario se autenticó
4. **Reutiliza PAM** - Aprovecha la infraestructura de autenticación del sistema

### ⚠️ Consideraciones

1. **HTTPS Recomendado**
   - Las credenciales viajan en Base64 (no cifradas)
   - Usar reverse proxy con SSL/TLS

   ```nginx
   server {
       listen 443 ssl;
       server_name watchdog.example.com;
       
       ssl_certificate /etc/ssl/certs/watchdog.crt;
       ssl_certificate_key /etc/ssl/private/watchdog.key;
       
       location / {
           proxy_pass http://localhost:8080;
           proxy_set_header Authorization $http_authorization;
       }
   }
   ```

2. **Firewall**
   - Bloquear puerto 8080 desde fuera
   
   ```bash
   sudo ufw allow from 192.168.1.0/24 to any port 8080
   sudo ufw deny 8080
   ```

3. **Usuarios Restringidos**
   - Considerar crear usuarios específicos para el dashboard
   - No usar usuarios con privilegios sudo para login

## 📊 Logs de Autenticación

El dashboard registra todos los intentos de autenticación:

```bash
# Ver logs de autenticación
journalctl -u neon-watchdog | grep "authenticated"

# Ejemplos de logs:
2026-01-09T21:50:15.123Z level=INFO msg="user authenticated successfully" user=usuario
2026-01-09T21:50:45.456Z level=WARN msg="authentication failed" user=hacker
```

## 🧪 Testing

### Probar autenticación desde curl

```bash
# Con credenciales correctas
curl -u usuario:tu_password http://localhost:8080/api/status

# Credenciales incorrectas (401)
curl -u usuario:wrong_password http://localhost:8080/api/status

# Sin credenciales (401)
curl http://localhost:8080/api/status
```

### Probar desde navegador

```bash
# Abrir dashboard
xdg-open http://localhost:8080

# O con wget
wget --http-user=usuario --http-password=tu_password http://localhost:8080/api/status
```

## ⚙️ Configuración

La autenticación está **siempre activada** cuando el dashboard está habilitado:

```yaml
dashboard:
  enabled: true  # Activa dashboard + autenticación
  port: 8080
  path: "/"
```

No hay forma de desactivar la autenticación por seguridad.

## 🔧 Troubleshooting

### Error: "Authentication Failed" con credenciales correctas

**Causa**: El proceso watchdog no tiene permisos para ejecutar `su`

**Solución**:
```bash
# Ejecutar con permisos necesarios
sudo ./neon-watchdog run -c examples/config-dashboard.yml

# O configurar sudoers para permitir 'su'
```

### Error: "user does not exist"

**Verificar que el usuario existe**:
```bash
id usuario
```

### El diálogo de autenticación no aparece

**Limpiar caché del navegador**:
```bash
# Firefox
Ctrl+Shift+Delete → Limpiar historial

# O usar modo incógnito
firefox --private-window http://localhost:8080
```

## 💡 Mejoras Futuras (Opcionales)

Si necesitas características adicionales:

1. **Integración LDAP/Active Directory**
   - Validar contra servidor de autenticación corporativo
   - Requiere librería `github.com/go-ldap/ldap`

2. **Autenticación con Token JWT**
   - Login una vez, token persistente
   - Mejor para APIs

3. **Multi-factor Authentication (MFA)**
   - TOTP codes (Google Authenticator)
   - Requiere `github.com/pquerna/otp`

4. **Rate Limiting**
   - Prevenir fuerza bruta
   - Bloquear IPs después de N intentos

5. **Sesiones persistentes**
   - Cookies de sesión
   - No pedir credenciales cada vez

## 📝 Resumen

✅ **Autenticación implementada y funcionando**
✅ **Usa usuarios del sistema operativo**
✅ **HTTP Basic Auth (estándar)**
✅ **Logs de todos los accesos**
✅ **Protege TODAS las rutas (UI + API)**

**¡El dashboard ahora es seguro y solo accesible por usuarios autorizados del sistema!** 🔐
