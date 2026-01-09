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
	IntervalSeconds int            `yaml:"interval_seconds" json:"interval_seconds"`
	TimeoutSeconds  int            `yaml:"timeout_seconds" json:"timeout_seconds"`
	LogLevel        string         `yaml:"log_level" json:"log_level"`
	StateFile       string         `yaml:"state_file" json:"state_file"`
	DefaultPolicy   Policy         `yaml:"default_policy" json:"default_policy"`
	Targets         []Target       `yaml:"targets" json:"targets"`
}

// Policy define la política de reintentos y rate limiting
type Policy struct {
	FailThreshold           int `yaml:"fail_threshold" json:"fail_threshold"`
	RestartCooldownSeconds  int `yaml:"restart_cooldown_seconds" json:"restart_cooldown_seconds"`
	MaxRestartsPerHour      int `yaml:"max_restarts_per_hour" json:"max_restarts_per_hour"`
}

// Target representa un servicio/proceso a monitorizar
type Target struct {
	Name    string  `yaml:"name" json:"name"`
	Enabled bool    `yaml:"enabled" json:"enabled"`
	Checks  []Check `yaml:"checks" json:"checks"`
	Action  Action  `yaml:"action" json:"action"`
	Policy  *Policy `yaml:"policy,omitempty" json:"policy,omitempty"`
}

// Check representa un tipo de verificación
type Check struct {
	Type        string   `yaml:"type" json:"type"` // process_name, pid_file, tcp_port, command
	ProcessName string   `yaml:"process_name,omitempty" json:"process_name,omitempty"`
	PidFile     string   `yaml:"pid_file,omitempty" json:"pid_file,omitempty"`
	TcpPort     string   `yaml:"tcp_port,omitempty" json:"tcp_port,omitempty"`
	Command     []string `yaml:"command,omitempty" json:"command,omitempty"`
}

// Action representa la acción a ejecutar cuando falla un target
type Action struct {
	Type    string         `yaml:"type" json:"type"` // exec, systemd
	Exec    *ExecAction    `yaml:"exec,omitempty" json:"exec,omitempty"`
	Systemd *SystemdAction `yaml:"systemd,omitempty" json:"systemd,omitempty"`
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
	}

	if !validTypes[check.Type] {
		return fmt.Errorf("target[%s].checks[%d]: invalid type '%s' (must be: process_name, pid_file, tcp_port, command)",
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
