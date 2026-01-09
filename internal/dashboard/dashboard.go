package dashboard

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tgextreme/neon-watchdog/internal/config"
	"github.com/tgextreme/neon-watchdog/internal/logger"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// Dashboard proporciona API REST y UI web con gestión de configuración
type Dashboard struct {
	cfg        *config.DashboardConfig
	log        *logger.Logger
	mu         sync.RWMutex
	status     *Status
	configPath string
	fullConfig *config.Config
}

// Status representa el estado actual del watchdog
type Status struct {
	Uptime    time.Duration           `json:"uptime"`
	StartTime time.Time               `json:"start_time"`
	Targets   map[string]TargetStatus `json:"targets"`
}

// TargetStatus estado de un target
type TargetStatus struct {
	Name                string    `json:"name"`
	Healthy             bool      `json:"healthy"`
	Enabled             bool      `json:"enabled"`
	LastCheck           time.Time `json:"last_check"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	TotalRestarts       int       `json:"total_restarts"`
	LastRestart         time.Time `json:"last_restart,omitempty"`
	Message             string    `json:"message"`
}

// NewDashboard crea un nuevo dashboard
func NewDashboard(cfg *config.DashboardConfig, log *logger.Logger) *Dashboard {
	if cfg == nil {
		cfg = &config.DashboardConfig{Enabled: false}
	}

	if cfg.Path == "" {
		cfg.Path = "/"
	}

	return &Dashboard{
		cfg: cfg,
		log: log,
		status: &Status{
			StartTime: time.Now(),
			Targets:   make(map[string]TargetStatus),
		},
		configPath: "",
		fullConfig: nil,
	}
}

// SetConfigPath establece la ruta del archivo de configuración
func (d *Dashboard) SetConfigPath(path string, cfg *config.Config) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.configPath = path
	d.fullConfig = cfg
}

// authenticateUser valida usuario/contraseña contra archivo users.txt
func (d *Dashboard) authenticateUser(username, password string) bool {
	if username == "" || password == "" {
		return false
	}

	// Leer archivo de usuarios
	file, err := os.Open("users.txt")
	if err != nil {
		d.log.Error("failed to open users file", logger.Fields("error", err.Error()))
		return false
	}
	defer file.Close()

	// Buscar usuario y validar hash
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Saltar líneas vacías y comentarios
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Formato: usuario:hash
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		fileUser := strings.TrimSpace(parts[0])
		fileHash := strings.TrimSpace(parts[1])

		// Si encontramos el usuario, validar password
		if fileUser == username {
			err := bcrypt.CompareHashAndPassword([]byte(fileHash), []byte(password))
			if err == nil {
				d.log.Info("user authenticated successfully", logger.Fields("user", username))
				return true
			}
			d.log.Warn("authentication failed - invalid password", logger.Fields("user", username))
			return false
		}
	}

	d.log.Warn("authentication failed - user not found", logger.Fields("user", username))
	return false
}

// authMiddleware middleware de autenticación HTTP Basic
func (d *Dashboard) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			d.requestAuth(w)
			return
		}

		// Parsear Basic Auth
		const prefix = "Basic "
		if !strings.HasPrefix(auth, prefix) {
			d.requestAuth(w)
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
		if err != nil {
			d.requestAuth(w)
			return
		}

		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 {
			d.requestAuth(w)
			return
		}

		username := credentials[0]
		password := credentials[1]

		// Validar contra sistema
		if !d.authenticateUser(username, password) {
			d.requestAuth(w)
			return
		}

		// Autenticación exitosa
		next(w, r)
	}
}

// requestAuth solicita autenticación HTTP Basic
func (d *Dashboard) requestAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Neon Watchdog Dashboard"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("401 - Autenticación Requerida\n\nUsuarios disponibles en users.txt"))
}

// Start inicia el servidor del dashboard
func (d *Dashboard) Start() error {
	if !d.cfg.Enabled {
		return nil
	}

	if d.cfg.Port == 0 {
		d.cfg.Port = 8080
	}

	// Proteger todas las rutas con autenticación
	http.HandleFunc(d.cfg.Path, d.authMiddleware(d.handleUI))
	http.HandleFunc("/api/status", d.authMiddleware(d.handleAPIStatus))
	http.HandleFunc("/api/health", d.authMiddleware(d.handleAPIHealth))
	http.HandleFunc("/api/targets", d.authMiddleware(d.handleAPITargets))
	http.HandleFunc("/api/targets/", d.authMiddleware(d.handleAPITargetByName))
	http.HandleFunc("/api/config", d.authMiddleware(d.handleAPIConfig))

	addr := fmt.Sprintf(":%d", d.cfg.Port)
	d.log.Info("dashboard starting", logger.Fields(
		"port", d.cfg.Port,
		"path", d.cfg.Path,
	))

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			d.log.Error("dashboard server failed", logger.Fields("error", err.Error()))
		}
	}()

	return nil
}

// UpdateTarget actualiza el estado de un target
func (d *Dashboard) UpdateTarget(name string, healthy bool, enabled bool, consecutiveFailures int, message string) {
	if !d.cfg.Enabled {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	ts := d.status.Targets[name]
	ts.Name = name
	ts.Healthy = healthy
	ts.Enabled = enabled
	ts.LastCheck = time.Now()
	ts.ConsecutiveFailures = consecutiveFailures
	ts.Message = message

	d.status.Targets[name] = ts
	d.status.Uptime = time.Since(d.status.StartTime)
}

// RecordRestart registra un restart
func (d *Dashboard) RecordRestart(name string) {
	if !d.cfg.Enabled {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	ts := d.status.Targets[name]
	ts.TotalRestarts++
	ts.LastRestart = time.Now()
	d.status.Targets[name] = ts
}

// handleAPIStatus retorna JSON con el estado completo
func (d *Dashboard) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	d.status.Uptime = time.Since(d.status.StartTime)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d.status)
}

// handleAPIHealth endpoint simple de health check
func (d *Dashboard) handleAPIHealth(w http.ResponseWriter, r *http.Request) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	allHealthy := true
	for _, ts := range d.status.Targets {
		if ts.Enabled && !ts.Healthy {
			allHealthy = false
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if allHealthy {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"targets": len(d.status.Targets),
		})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unhealthy",
			"targets": len(d.status.Targets),
		})
	}
}

// handleUI retorna la interfaz HTML
func (d *Dashboard) handleUI(w http.ResponseWriter, r *http.Request) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	d.status.Uptime = time.Since(d.status.StartTime)

	// Usar template básico de visualización
	tmpl := template.Must(template.New("dashboard").Parse(htmlTemplate))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, d.status)
}

// handleAPITargets maneja GET (list) y POST (create) de targets
func (d *Dashboard) handleAPITargets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		d.listTargets(w, r)
	case http.MethodPost:
		d.createTarget(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPITargetByName maneja GET, PUT, DELETE de un target específico
func (d *Dashboard) handleAPITargetByName(w http.ResponseWriter, r *http.Request) {
	// Extraer nombre del path: /api/targets/{name}
	name := r.URL.Path[len("/api/targets/"):]
	if name == "" {
		http.Error(w, "Target name required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		d.getTarget(w, r, name)
	case http.MethodPut:
		d.updateTarget(w, r, name)
	case http.MethodDelete:
		d.deleteTarget(w, r, name)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPIConfig retorna la configuración completa
func (d *Dashboard) handleAPIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.fullConfig == nil {
		http.Error(w, "Config not available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d.fullConfig)
}

// listTargets lista todos los targets de la configuración
func (d *Dashboard) listTargets(w http.ResponseWriter, r *http.Request) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.fullConfig == nil {
		http.Error(w, "Config not available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d.fullConfig.Targets)
}

// getTarget obtiene un target específico
func (d *Dashboard) getTarget(w http.ResponseWriter, r *http.Request, name string) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.fullConfig == nil {
		http.Error(w, "Config not available", http.StatusServiceUnavailable)
		return
	}

	for _, target := range d.fullConfig.Targets {
		if target.Name == name {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(target)
			return
		}
	}

	http.Error(w, "Target not found", http.StatusNotFound)
}

// createTarget crea un nuevo target
func (d *Dashboard) createTarget(w http.ResponseWriter, r *http.Request) {
	var newTarget config.Target
	if err := json.NewDecoder(r.Body).Decode(&newTarget); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.fullConfig == nil || d.configPath == "" {
		http.Error(w, "Config not available", http.StatusServiceUnavailable)
		return
	}

	// Verificar que no exista
	for _, target := range d.fullConfig.Targets {
		if target.Name == newTarget.Name {
			http.Error(w, "Target already exists", http.StatusConflict)
			return
		}
	}

	// Añadir target
	d.fullConfig.Targets = append(d.fullConfig.Targets, newTarget)

	// Guardar en disco
	if err := d.saveConfig(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	d.log.Info("target created via dashboard", logger.Fields("name", newTarget.Name))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTarget)
}

// updateTarget actualiza un target existente
func (d *Dashboard) updateTarget(w http.ResponseWriter, r *http.Request, name string) {
	var updatedTarget config.Target
	if err := json.NewDecoder(r.Body).Decode(&updatedTarget); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.fullConfig == nil || d.configPath == "" {
		http.Error(w, "Config not available", http.StatusServiceUnavailable)
		return
	}

	// Buscar y actualizar
	found := false
	for i, target := range d.fullConfig.Targets {
		if target.Name == name {
			d.fullConfig.Targets[i] = updatedTarget
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Target not found", http.StatusNotFound)
		return
	}

	// Guardar en disco
	if err := d.saveConfig(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	d.log.Info("target updated via dashboard", logger.Fields("name", name))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTarget)
}

// deleteTarget elimina un target
func (d *Dashboard) deleteTarget(w http.ResponseWriter, r *http.Request, name string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.fullConfig == nil || d.configPath == "" {
		http.Error(w, "Config not available", http.StatusServiceUnavailable)
		return
	}

	// Buscar y eliminar
	found := false
	for i, target := range d.fullConfig.Targets {
		if target.Name == name {
			d.fullConfig.Targets = append(d.fullConfig.Targets[:i], d.fullConfig.Targets[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Target not found", http.StatusNotFound)
		return
	}

	// Guardar en disco
	if err := d.saveConfig(); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	d.log.Info("target deleted via dashboard", logger.Fields("name", name))

	w.WriteHeader(http.StatusNoContent)
}

// saveConfig guarda la configuración en disco (debe llamarse con lock activo)
func (d *Dashboard) saveConfig() error {
	data, err := yaml.Marshal(d.fullConfig)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	// Crear backup
	if _, err := os.Stat(d.configPath); err == nil {
		backup := d.configPath + ".backup"
		input, err := os.ReadFile(d.configPath)
		if err == nil {
			os.WriteFile(backup, input, 0644)
		}
	}

	// Escribir nueva configuración
	if err := os.WriteFile(d.configPath, data, 0644); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	return nil
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Neon Watchdog Dashboard</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        .header {
            background: white;
            border-radius: 12px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .header h1 {
            color: #667eea;
            font-size: 2em;
            margin-bottom: 10px;
        }
        .header .uptime {
            color: #666;
            font-size: 0.9em;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-bottom: 20px;
        }
        .stat-card {
            background: white;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .stat-card .number {
            font-size: 2.5em;
            font-weight: bold;
            color: #667eea;
        }
        .stat-card .label {
            color: #666;
            margin-top: 5px;
        }
        .targets {
            display: grid;
            gap: 15px;
        }
        .target-card {
            background: white;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            border-left: 4px solid #ccc;
        }
        .target-card.healthy { border-left-color: #10b981; }
        .target-card.unhealthy { border-left-color: #ef4444; }
        .target-card.disabled { border-left-color: #9ca3af; opacity: 0.6; }
        .target-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
        }
        .target-name {
            font-size: 1.3em;
            font-weight: bold;
            color: #333;
        }
        .status-badge {
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 0.85em;
            font-weight: bold;
        }
        .status-badge.healthy {
            background: #d1fae5;
            color: #065f46;
        }
        .status-badge.unhealthy {
            background: #fee2e2;
            color: #991b1b;
        }
        .status-badge.disabled {
            background: #f3f4f6;
            color: #6b7280;
        }
        .target-details {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 10px;
            font-size: 0.9em;
            color: #666;
        }
        .detail-item {
            padding: 8px;
            background: #f9fafb;
            border-radius: 4px;
        }
        .detail-label {
            font-weight: bold;
            color: #374151;
        }
        .refresh-notice {
            text-align: center;
            color: white;
            margin-top: 20px;
            font-size: 0.9em;
        }
    </style>
    <script>
        // Auto-refresh cada 5 segundos
        setTimeout(() => location.reload(), 5000);
    </script>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🐺 Neon Watchdog Dashboard</h1>
            <div class="uptime">Uptime: {{printf "%.0f" .Uptime.Seconds}}s | Started: {{.StartTime.Format "2006-01-02 15:04:05"}}</div>
        </div>

        <div class="stats">
            <div class="stat-card">
                <div class="number">{{len .Targets}}</div>
                <div class="label">Total Targets</div>
            </div>
            <div class="stat-card">
                <div class="number">{{range .Targets}}{{if and .Enabled .Healthy}}1{{end}}{{end}}</div>
                <div class="label">Healthy</div>
            </div>
            <div class="stat-card">
                <div class="number">{{range .Targets}}{{if and .Enabled (not .Healthy)}}1{{end}}{{end}}</div>
                <div class="label">Unhealthy</div>
            </div>
        </div>

        <div class="targets">
            {{range .Targets}}
            <div class="target-card {{if .Enabled}}{{if .Healthy}}healthy{{else}}unhealthy{{end}}{{else}}disabled{{end}}">
                <div class="target-header">
                    <div class="target-name">{{.Name}}</div>
                    <span class="status-badge {{if .Enabled}}{{if .Healthy}}healthy{{else}}unhealthy{{end}}{{else}}disabled{{end}}">
                        {{if .Enabled}}{{if .Healthy}}✓ Healthy{{else}}✗ Unhealthy{{end}}{{else}}Disabled{{end}}
                    </span>
                </div>
                <div class="target-details">
                    <div class="detail-item">
                        <div class="detail-label">Last Check</div>
                        <div>{{if not .LastCheck.IsZero}}{{.LastCheck.Format "15:04:05"}}{{else}}Never{{end}}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Failures</div>
                        <div>{{.ConsecutiveFailures}}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Total Restarts</div>
                        <div>{{.TotalRestarts}}</div>
                    </div>
                    {{if not .LastRestart.IsZero}}
                    <div class="detail-item">
                        <div class="detail-label">Last Restart</div>
                        <div>{{.LastRestart.Format "15:04:05"}}</div>
                    </div>
                    {{end}}
                </div>
                {{if .Message}}
                <div style="margin-top: 10px; padding: 10px; background: #fef3c7; border-radius: 4px; font-size: 0.85em;">
                    {{.Message}}
                </div>
                {{end}}
            </div>
            {{end}}
        </div>

        <div class="refresh-notice">⟳ Auto-refresh in 5 seconds</div>
    </div>
</body>
</html>`
