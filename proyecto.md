A continuación tienes una especificación minuciosa (pero realista para un MVP en 2 horas) de **Neon Watchdog**, un “guardian” de procesos orientado a Linux con **CLI**, **logs**, y **configuración en YAML/JSON**, integrable con **systemd** (service + timer) o ejecutable como demonio con bucle interno.

---

## 1) Objetivo del proyecto

**Neon Watchdog** monitoriza uno o varios “targets” (procesos/servicios) definidos por configuración y ejecuta una acción de recuperación cuando detecta que el target está “caído” o “no saludable”.

**Qué significa “caído” (mínimo viable):**

* No existe un proceso con ese nombre (o PID file no válido), **o**
* Un puerto TCP no está escuchando, **o**
* Un comando de healthcheck devuelve error.

**Qué significa “recuperar” (mínimo viable):**

* Ejecutar un comando de arranque/reinicio definido por el usuario, **o**
* Llamar a `systemctl restart <unidad>` si el target es un servicio systemd.

---

## 2) Alcance MVP (en 2 horas) y alcance ampliable

### MVP en 2 horas (lo que *debería* entrar)

1. **CLI** con al menos:

   * `neon-watchdog run -c /ruta/config.yml`
   * `neon-watchdog check -c /ruta/config.yml` (una pasada, exit code útil)
   * `neon-watchdog test-config -c ...` (validación y “dry-run”)
2. **Config YAML/JSON** con:

   * Intervalo de chequeo
   * Lista de targets
   * Política de reintentos/backoff simple
   * Acciones de reinicio
3. **Logs**:

   * Niveles (INFO/WARN/ERROR/DEBUG)
   * Mensajes estructurados (clave=valor) para que journald los indexe bien
4. **systemd**:

   * Opción A: servicio “oneshot” + timer (recomendable para simpleza)
   * Opción B: servicio persistente (Type=simple) con bucle interno y `sleep`

### Ampliaciones naturales (post-MVP)

* Reload de config sin reiniciar (SIGHUP o `systemctl reload`)
* Healthcheck HTTP (GET a `/healthz`)
* Notificaciones (email/Telegram/Discord webhook)
* Métricas Prometheus (endpoint local)
* Modo “auto-install systemd” desde CLI
* Rate limiting avanzado (ventana temporal, jitter)
* Soporte de múltiples métodos de identificación (cgroup, cmdline match, regex)

---

## 3) Arquitectura conceptual

### Componentes

1. **Loader/Parser de configuración**

   * Lee YAML/JSON
   * Valida tipos y valores (p.ej. intervalos > 0)
   * Normaliza defaults (si faltan campos)
2. **Scheduler**

   * Define cuándo ejecutar chequeos (timer systemd o bucle interno)
3. **Engine de chequeo (Checkers)**

   * `ProcessNameChecker`: busca proceso por nombre (p.ej. `pgrep`)
   * `PidFileChecker`: lee PID y valida si existe
   * `TcpPortChecker`: intenta conectar a `host:port` o verifica escucha local
   * `CommandChecker`: ejecuta comando y usa exit code
4. **Policy (decisión)**

   * Decide si un target está OK / FAIL
   * Define cuándo reiniciar (al primer fallo o tras N fallos consecutivos)
   * Controla backoff y límite de reinicios
5. **Actions (recuperación)**

   * `ExecAction`: ejecuta comando(s) definidos
   * `SystemdRestartAction`: `systemctl restart <unit>`
6. **Logger**

   * Output a stdout/stderr (systemd lo captura)
   * Formato consistente: timestamp, target, checker, resultado, latencia
7. **State store (mínimo)**

   * Estado en memoria: contador de fallos consecutivos, último reinicio, etc.
   * (Opcional) persistir en `/var/lib/neon-watchdog/state.json` para no “olvidar” tras reinicio

---

## 4) Flujo de ejecución

Por cada ciclo (o invocación oneshot):

1. Cargar config
2. Para cada target:

   * Ejecutar uno o varios checkers
   * Si todos OK → log INFO “healthy”
   * Si alguno FAIL → incrementar `fail_count`
3. Si `fail_count` alcanza umbral (por defecto 1 en MVP):

   * Comprobar rate limit (no reiniciar 20 veces en 1 minuto)
   * Ejecutar acción de recuperación
   * Registrar resultado (success/failure) y resetear contadores según política
4. Salir con código:

   * `0` si todo OK
   * `2` si al menos un target está FAIL (útil para systemd/journal y scripts)

---

## 5) Configuración (YAML/JSON) — qué debe incluir

### Sección global

* `interval_seconds`: cada cuánto comprobar (si modo daemon)
* `timeout_seconds`: timeout por chequeo (comandos/red)
* `log_level`: INFO por defecto
* `state_file` (opcional): persistencia básica
* `default_policy`:

  * `fail_threshold` (N fallos consecutivos antes de reiniciar)
  * `restart_cooldown_seconds` (mínimo entre reinicios)
  * `max_restarts_per_hour` (rate limit simple)

### Targets

Cada target debería tener:

* `name`: identificador legible (aparece en logs)
* `enabled`: true/false
* `checks`: lista de chequeos (1..N)
* `action`: qué hacer si falla

**Checks soportables en MVP:**

* `process_name`: `"nginx"` (o `match_cmdline_contains`)
* `pid_file`: `"/run/nginx.pid"`
* `tcp_port`: `127.0.0.1:8080`
* `command`: `["/usr/bin/curl","-fsS","http://127.0.0.1:8080/health"]` (si lo amplías)

  * Para MVP, incluso un comando simple tipo `["pgrep","-x","nginx"]` sirve

**Actions MVP:**

* `exec`:

  * `start`: comando (o lista) para levantarlo
  * `restart`: comando (o lista) para reiniciar
* `systemd`:

  * `unit`: `"nginx.service"`
  * `method`: restart

### Defaults y validación

* Si un target no define policy propia, hereda la global.
* Validaciones mínimas:

  * Intervalos > 0
  * Puertos 1..65535
  * Rutas absolutas donde aplique
  * Acción definida si hay checks

---

## 6) CLI: comandos y comportamiento esperado

**Comandos mínimos recomendados:**

* `run -c <config>`
  Ejecuta en bucle (si eliges modo daemon) o “oneshot” si así lo decides.
* `check -c <config>`
  Ejecuta una pasada; exit code comunica estado (ideal con systemd timer).
* `test-config -c <config>`
  Parsea y valida; imprime resumen de targets y checks detectados.

**Flags útiles:**

* `--once` (si `run` es bucle)
* `--verbose` / `--debug`
* `--json-logs` (si quieres logs más machine-readable)
* `--dry-run` (no ejecutar acciones, solo simular)

---

## 7) Logging y observabilidad (qué “debe” verse en journald)

**Eventos que deben loguearse siempre:**

* Inicio y versión (con build info si existe)
* Config cargada (ruta, número de targets activos)
* Por target:

  * resultado del check: OK/FAIL
  * latencia del check (ms)
  * motivo del fallo (p.ej. “port connection refused”)
* Si ejecuta reinicio:

  * motivo + política aplicada (fail_count, cooldown, rate limit)
  * comando/unidad systemd invocada (sin exponer secretos)
  * resultado (exit code, stderr recortado)

**Formato recomendado (clave=valor):**

* `target=api port=8080 check=tcp result=FAIL err="connection refused"`

Esto facilita filtrar con `journalctl -u neon-watchdog`.

---

## 8) systemd: dos patrones y cuándo usar cada uno

### Patrón A (recomendado en MVP): service oneshot + timer

* **Ventaja:** simple, robusto, reinicios y scheduling los controla systemd.
* **Cómo encaja:** el timer dispara `neon-watchdog check -c ...` cada X segundos/minutos.

Casos ideales: watchdogs sencillos, sin necesidad de estado complejo.

### Patrón B: service persistente (daemon)

* **Ventaja:** estado en memoria, latencia mínima, lógica de backoff más fina.
* **Riesgo:** tienes que manejar tú señales, reinicios, y evitar “loops” peligrosos.

Casos ideales: entornos donde quieres checks muy frecuentes o con estado avanzado.

---

## 9) Confiabilidad y seguridad (mínimo sensato)

1. **Usuario dedicado** (si no necesitas root)

   * `User=neon-watchdog`
   * Permisos mínimos sobre PID files y sockets necesarios
2. **Evitar tormentas de reinicio**

   * `restart_cooldown_seconds`
   * `max_restarts_per_hour`
3. **Timeouts**

   * En checks de comando/red para no colgar el watchdog
4. **Lock para evitar instancias dobles**

   * lockfile en `/run/neon-watchdog.lock` (o systemd ya lo evita si es service único)
5. **Sanitización básica de comandos**

   * Si aceptas acciones `exec`, mejor modelarlas como array (argv) y no como string shell.
6. **No loguear secretos**

   * Si hay tokens en comandos, enmascararlos o prohibirlos en config.

---

## 10) Estructura de proyecto (organización típica)

Una estructura limpia (independiente del lenguaje) suele quedar así:

* `cmd/` o `src/cli/` → parsing de argumentos y comandos
* `src/config/` → loader YAML/JSON + validación
* `src/checks/` → implementaciones de checkers
* `src/actions/` → implementaciones de acciones
* `src/engine/` → loop principal + policy
* `packaging/systemd/` → unit files y timer
* `examples/` → config de ejemplo
* `README.md` → uso, instalación, ejemplos, troubleshooting

---

## 11) Pruebas mínimas (para que no sea frágil)

En MVP, al menos:

* Unit test: parseo/validación de config (casos buenos y malos)
* Smoke test manual:

  1. target “dummy” con un proceso que matas
  2. ver que detecta FAIL
  3. ver que ejecuta acción y vuelve a OK

Si quieres un test automatizable: un script que levanta un `python -m http.server` y luego lo mata.

---

## 12) Entregables concretos que “debería tener” el repositorio

1. Binario/ejecutable `neon-watchdog`
2. `examples/config.yml`
3. `systemd/neon-watchdog.service`
4. (si timer) `systemd/neon-watchdog.timer`
5. `README.md` con:

   * instalación
   * ejemplos
   * comandos de systemd (`systemctl enable --now ...`)
   * cómo ver logs (`journalctl -u ... -f`)
   * troubleshooting (permisos, PATH, unidades)
6. `LICENSE` (MIT/Apache-2.0 habitual)
7. `CHANGELOG.md` (opcional, pero útil)

---

## 13) Plan de ejecución para hacerlo sin agobio (en pasos cortos)

1. **Empieza por el modo timer (oneshot)**: implementa solo `check` + logs.
2. Añade **un único checker** primero (p.ej. `tcp_port`).
3. Añade **una acción simple** (p.ej. `systemctl restart`).
4. Integra con systemd: `timer` cada 15s/30s.
5. Solo cuando eso funcione, añade `process_name` y `exec action`.

Si quieres, en tu siguiente mensaje dime el lenguaje elegido (Rust, Go, Python, etc.) y si prefieres **timer** o **daemon**; con eso te convierto esta especificación en un checklist de implementación con decisiones ya cerradas (sin ambigüedades) y una config de ejemplo alineada al caso real que quieras vigilar.
