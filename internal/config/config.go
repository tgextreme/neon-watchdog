package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config representa la configuración completa del watchdog
type Config struct {
	IntervalSeconds int              `yaml:"interval_seconds" json:"interval_seconds"`
	TimeoutSeconds  int              `yaml:"timeout_seconds" json:"timeout_seconds"`
	LogLevel        string           `yaml:"log_level" json:"log_level"`
	StateFile       string           `yaml:"state_file" json:"state_file"`
	DefaultPolicy   Policy           `yaml:"default_policy" json:"default_policy"`
	Targets         []Target         `yaml:"targets" json:"targets"`
	Notifications   []Notification   `yaml:"notifications,omitempty" json:"notifications,omitempty"`
	Metrics         *MetricsConfig   `yaml:"metrics,omitempty" json:"metrics,omitempty"`
	Dashboard       *DashboardConfig `yaml:"dashboard,omitempty" json:"dashboard,omitempty"`
	History         *HistoryConfig   `yaml:"history,omitempty" json:"history,omitempty"`
}

// Policy define la política de reintentos y rate limiting
type Policy struct {
	FailThreshold          int    `yaml:"fail_threshold" json:"fail_threshold"`
	RestartCooldownSeconds int    `yaml:"restart_cooldown_seconds" json:"restart_cooldown_seconds"`
	MaxRestartsPerHour     int    `yaml:"max_restarts_per_hour" json:"max_restarts_per_hour"`
	BackoffStrategy        string `yaml:"backoff_strategy,omitempty" json:"backoff_strategy,omitempty"` // linear, exponential
	MaxBackoffSeconds      int    `yaml:"max_backoff_seconds,omitempty" json:"max_backoff_seconds,omitempty"`
}

// Notification define configuración de notificaciones
type Notification struct {
	Type     string          `yaml:"type" json:"type"` // email, webhook, telegram
	Enabled  bool            `yaml:"enabled" json:"enabled"`
	Email    *EmailConfig    `yaml:"email,omitempty" json:"email,omitempty"`
	Webhook  *WebhookConfig  `yaml:"webhook,omitempty" json:"webhook,omitempty"`
	Telegram *TelegramConfig `yaml:"telegram,omitempty" json:"telegram,omitempty"`
}

// EmailConfig configuración para notificaciones por email
type EmailConfig struct {
	SMTPHost string   `yaml:"smtp_host" json:"smtp_host"`
	SMTPPort int      `yaml:"smtp_port" json:"smtp_port"`
	Username string   `yaml:"username,omitempty" json:"username,omitempty"`
	Password string   `yaml:"password,omitempty" json:"password,omitempty"`
	From     string   `yaml:"from" json:"from"`
	To       []string `yaml:"to" json:"to"`
	UseTLS   bool     `yaml:"use_tls" json:"use_tls"`
}

// WebhookConfig configuración para notificaciones por webhook
type WebhookConfig struct {
	URL     string            `yaml:"url" json:"url"`
	Method  string            `yaml:"method,omitempty" json:"method,omitempty"` // default: POST
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Timeout int               `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// TelegramConfig configuración para notificaciones por Telegram
type TelegramConfig struct {
	BotToken string `yaml:"bot_token" json:"bot_token"`
	ChatID   string `yaml:"chat_id" json:"chat_id"`
}

// MetricsConfig configuración de métricas Prometheus
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Port    int    `yaml:"port" json:"port"`
	Path    string `yaml:"path,omitempty" json:"path,omitempty"`
}

// DashboardConfig configuración del dashboard web
type DashboardConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Port    int    `yaml:"port" json:"port"`
	Path    string `yaml:"path,omitempty" json:"path,omitempty"`
}

// HistoryConfig configuración del historial persistente
type HistoryConfig struct {
	MaxEntries     int `yaml:"max_entries" json:"max_entries"`
	RetentionHours int `yaml:"retention_hours" json:"retention_hours"`
}

// Target representa un servicio/proceso a monitorizar
type Target struct {
	Name      string   `yaml:"name" json:"name"`
	Enabled   bool     `yaml:"enabled" json:"enabled"`
	DependsOn []string `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Checks    []Check  `yaml:"checks" json:"checks"`
	Action    Action   `yaml:"action" json:"action"`
	Policy    *Policy  `yaml:"policy,omitempty" json:"policy,omitempty"`
}

// Check representa un tipo de verificación
type Check struct {
	Type            string       `yaml:"type" json:"type"` // process_name, pid_file, tcp_port, command, http, script, logic
	ProcessName     string       `yaml:"process_name,omitempty" json:"process_name,omitempty"`
	IgnoreExitCodes []int        `yaml:"ignore_exit_codes,omitempty" json:"ignore_exit_codes,omitempty"`
	PidFile         string       `yaml:"pid_file,omitempty" json:"pid_file,omitempty"`
	TcpPort         string       `yaml:"tcp_port,omitempty" json:"tcp_port,omitempty"`
	Command         []string     `yaml:"command,omitempty" json:"command,omitempty"`
	HTTP            *HTTPCheck   `yaml:"http,omitempty" json:"http,omitempty"`
	Script          *ScriptCheck `yaml:"script,omitempty" json:"script,omitempty"`
	Logic           string       `yaml:"logic,omitempty" json:"logic,omitempty"`   // AND, OR
	Checks          []Check      `yaml:"checks,omitempty" json:"checks,omitempty"` // For logic groups
}

// HTTPCheck configuración para health checks HTTP
type HTTPCheck struct {
	URL            string            `yaml:"url" json:"url"`
	Method         string            `yaml:"method,omitempty" json:"method,omitempty"`                   // default: GET
	ExpectedStatus int               `yaml:"expected_status,omitempty" json:"expected_status,omitempty"` // default: 200
	Headers        map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body           string            `yaml:"body,omitempty" json:"body,omitempty"`
	TimeoutSeconds int               `yaml:"timeout_seconds,omitempty" json:"timeout_seconds,omitempty"`
}

// ScriptCheck configuración para scripts personalizados
type ScriptCheck struct {
	Path             string   `yaml:"path" json:"path"`
	Args             []string `yaml:"args,omitempty" json:"args,omitempty"`
	SuccessExitCodes []int    `yaml:"success_exit_codes,omitempty" json:"success_exit_codes,omitempty"`
	WarningExitCodes []int    `yaml:"warning_exit_codes,omitempty" json:"warning_exit_codes,omitempty"`
}

// Action representa la acción a ejecutar cuando falla un target
type Action struct {
	Type    string         `yaml:"type" json:"type"` // exec, systemd
	Exec    *ExecAction    `yaml:"exec,omitempty" json:"exec,omitempty"`
	Systemd *SystemdAction `yaml:"systemd,omitempty" json:"systemd,omitempty"`
	Hooks   *ActionHooks   `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

// ActionHooks define hooks para ejecutar antes/después de acciones
type ActionHooks struct {
	BeforeRestart []string `yaml:"before_restart,omitempty" json:"before_restart,omitempty"`
	AfterRestart  []string `yaml:"after_restart,omitempty" json:"after_restart,omitempty"`
	OnFailure     []string `yaml:"on_failure,omitempty" json:"on_failure,omitempty"`
}

// ExecAction define comandos a ejecutar
type ExecAction struct {
	Start   []string `yaml:"start,omitempty" json:"start,omitempty"`
	Restart []string `yaml:"restart,omitempty" json:"restart,omitempty"`
}

// SystemdAction define una acción sobre una unidad systemd
type SystemdAction struct {
	Unit   string `yaml:"unit" json:"unit"`
	Method string `yaml:"method" json:"method"` // restart, start
}

// Load carga y parsea el archivo de configuración
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	cfg := &Config{}
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("error parsing YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("error parsing JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s (use .yaml, .yml, or .json)", ext)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	cfg.SetDefaults()
	return cfg, nil
}

// Validate valida la configuración
func (c *Config) Validate() error {
	if c.IntervalSeconds < 0 {
		return fmt.Errorf("interval_seconds must be >= 0")
	}

	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = 10 // default
	}

	if c.LogLevel == "" {
		c.LogLevel = "INFO"
	}

	validLevels := map[string]bool{"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true}
	if !validLevels[strings.ToUpper(c.LogLevel)] {
		return fmt.Errorf("invalid log_level: %s (must be DEBUG, INFO, WARN, or ERROR)", c.LogLevel)
	}

	// Validar política por defecto
	if c.DefaultPolicy.FailThreshold <= 0 {
		c.DefaultPolicy.FailThreshold = 1
	}

	if len(c.Targets) == 0 {
		return fmt.Errorf("no targets defined")
	}

	// Validar cada target
	for i, target := range c.Targets {
		if target.Name == "" {
			return fmt.Errorf("target[%d]: name is required", i)
		}

		if len(target.Checks) == 0 {
			return fmt.Errorf("target[%s]: at least one check is required", target.Name)
		}

		// Validar checks
		for j, check := range target.Checks {
			if err := validateCheck(check, target.Name, j); err != nil {
				return err
			}
		}

		// Validar action
		if err := validateAction(target.Action, target.Name); err != nil {
			return err
		}
	}

	return nil
}

// validateCheck valida un check individual
func validateCheck(check Check, targetName string, index int) error {
	validTypes := map[string]bool{
		"process_name": true,
		"pid_file":     true,
		"tcp_port":     true,
		"command":      true,
		"http":         true,
		"script":       true,
		"logic":        true,
	}

	if !validTypes[check.Type] {
		return fmt.Errorf("target[%s].checks[%d]: invalid type '%s' (must be: process_name, pid_file, tcp_port, command, http, script, logic)",
			targetName, index, check.Type)
	}

	switch check.Type {
	case "process_name":
		if check.ProcessName == "" {
			return fmt.Errorf("target[%s].checks[%d]: process_name is required for type 'process_name'", targetName, index)
		}
	case "pid_file":
		if check.PidFile == "" {
			return fmt.Errorf("target[%s].checks[%d]: pid_file is required for type 'pid_file'", targetName, index)
		}
	case "tcp_port":
		if check.TcpPort == "" {
			return fmt.Errorf("target[%s].checks[%d]: tcp_port is required for type 'tcp_port'", targetName, index)
		}
	case "command":
		if len(check.Command) == 0 {
			return fmt.Errorf("target[%s].checks[%d]: command is required for type 'command'", targetName, index)
		}
	case "http":
		if check.HTTP == nil || check.HTTP.URL == "" {
			return fmt.Errorf("target[%s].checks[%d]: http.url is required for type 'http'", targetName, index)
		}
	case "script":
		if check.Script == nil || check.Script.Path == "" {
			return fmt.Errorf("target[%s].checks[%d]: script.path is required for type 'script'", targetName, index)
		}
	case "logic":
		if check.Logic != "AND" && check.Logic != "OR" {
			return fmt.Errorf("target[%s].checks[%d]: logic must be 'AND' or 'OR'", targetName, index)
		}
		if len(check.Checks) == 0 {
			return fmt.Errorf("target[%s].checks[%d]: logic groups must have at least one check", targetName, index)
		}
		// Validar checks anidados
		for j, subCheck := range check.Checks {
			if err := validateCheck(subCheck, targetName, j); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateAction valida una acción
func validateAction(action Action, targetName string) error {
	validTypes := map[string]bool{"exec": true, "systemd": true}

	if !validTypes[action.Type] {
		return fmt.Errorf("target[%s].action: invalid type '%s' (must be: exec, systemd)", targetName, action.Type)
	}

	switch action.Type {
	case "exec":
		if action.Exec == nil {
			return fmt.Errorf("target[%s].action: exec configuration is required for type 'exec'", targetName)
		}
		if len(action.Exec.Start) == 0 && len(action.Exec.Restart) == 0 {
			return fmt.Errorf("target[%s].action: at least one of 'start' or 'restart' must be defined", targetName)
		}
	case "systemd":
		if action.Systemd == nil {
			return fmt.Errorf("target[%s].action: systemd configuration is required for type 'systemd'", targetName)
		}
		if action.Systemd.Unit == "" {
			return fmt.Errorf("target[%s].action: systemd.unit is required", targetName)
		}
		if action.Systemd.Method == "" {
			action.Systemd.Method = "restart"
		}
	}

	return nil
}

// SetDefaults establece valores por defecto
func (c *Config) SetDefaults() {
	if c.LogLevel == "" {
		c.LogLevel = "INFO"
	}

	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = 10
	}

	if c.DefaultPolicy.FailThreshold <= 0 {
		c.DefaultPolicy.FailThreshold = 1
	}

	if c.DefaultPolicy.RestartCooldownSeconds <= 0 {
		c.DefaultPolicy.RestartCooldownSeconds = 60
	}

	if c.DefaultPolicy.MaxRestartsPerHour <= 0 {
		c.DefaultPolicy.MaxRestartsPerHour = 10
	}

	// Aplicar política por defecto a targets que no la tienen
	for i := range c.Targets {
		if c.Targets[i].Policy == nil {
			c.Targets[i].Policy = &c.DefaultPolicy
		} else {
			// Completar valores faltantes con los defaults
			if c.Targets[i].Policy.FailThreshold <= 0 {
				c.Targets[i].Policy.FailThreshold = c.DefaultPolicy.FailThreshold
			}
			if c.Targets[i].Policy.RestartCooldownSeconds <= 0 {
				c.Targets[i].Policy.RestartCooldownSeconds = c.DefaultPolicy.RestartCooldownSeconds
			}
			if c.Targets[i].Policy.MaxRestartsPerHour <= 0 {
				c.Targets[i].Policy.MaxRestartsPerHour = c.DefaultPolicy.MaxRestartsPerHour
			}
		}
	}
}

// GetActiveTargets retorna solo los targets habilitados
func (c *Config) GetActiveTargets() []Target {
	active := []Target{}
	for _, target := range c.Targets {
		if target.Enabled {
			active = append(active, target)
		}
	}
	return active
}
