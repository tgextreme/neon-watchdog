package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/tgextreme/neon-watchdog/internal/actions"
	"github.com/tgextreme/neon-watchdog/internal/checks"
	"github.com/tgextreme/neon-watchdog/internal/config"
	"github.com/tgextreme/neon-watchdog/internal/logger"
)

// TargetState mantiene el estado de un target
type TargetState struct {
	Name                 string    `json:"name"`
	ConsecutiveFailures  int       `json:"consecutive_failures"`
	LastCheckTime        time.Time `json:"last_check_time"`
	LastRestartTime      time.Time `json:"last_restart_time"`
	RestartsInLastHour   []time.Time `json:"restarts_in_last_hour"`
	IsHealthy            bool      `json:"is_healthy"`
}

// State mantiene el estado global del watchdog
type State struct {
	Targets map[string]*TargetState `json:"targets"`
	mu      sync.RWMutex
}

// Engine es el motor principal del watchdog
type Engine struct {
	config *config.Config
	logger *logger.Logger
	state  *State
}

// New crea un nuevo engine
func New(cfg *config.Config, log *logger.Logger) *Engine {
	state := &State{
		Targets: make(map[string]*TargetState),
	}
	
	// Inicializar estado para cada target
	for _, target := range cfg.GetActiveTargets() {
		state.Targets[target.Name] = &TargetState{
			Name:                target.Name,
			ConsecutiveFailures: 0,
			IsHealthy:           true,
			RestartsInLastHour:  []time.Time{},
		}
	}
	
	return &Engine{
		config: cfg,
		logger: log,
		state:  state,
	}
}

// LoadState carga el estado desde un archivo
func (e *Engine) LoadState(path string) error {
	if path == "" {
		return nil
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No es un error si no existe
		}
		return fmt.Errorf("error reading state file: %w", err)
	}
	
	e.state.mu.Lock()
	defer e.state.mu.Unlock()
	
	if err := json.Unmarshal(data, e.state); err != nil {
		return fmt.Errorf("error parsing state file: %w", err)
	}
	
	e.logger.Info("state loaded from file", logger.Fields("path", path))
	return nil
}

// SaveState guarda el estado a un archivo
func (e *Engine) SaveState(path string) error {
	if path == "" {
		return nil
	}
	
	e.state.mu.RLock()
	data, err := json.MarshalIndent(e.state, "", "  ")
	e.state.mu.RUnlock()
	
	if err != nil {
		return fmt.Errorf("error marshaling state: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing state file: %w", err)
	}
	
	return nil
}

// CheckOnce ejecuta una pasada de checks sobre todos los targets
func (e *Engine) CheckOnce(ctx context.Context) bool {
	allHealthy := true
	
	for _, target := range e.config.GetActiveTargets() {
		healthy := e.checkTarget(ctx, target)
		if !healthy {
			allHealthy = false
		}
	}
	
	// Guardar estado si está configurado
	if e.config.StateFile != "" {
		if err := e.SaveState(e.config.StateFile); err != nil {
			e.logger.Error("failed to save state", logger.Fields("error", err))
		}
	}
	
	return allHealthy
}

// checkTarget verifica un target individual
func (e *Engine) checkTarget(ctx context.Context, target config.Target) bool {
	e.state.mu.Lock()
	state := e.state.Targets[target.Name]
	if state == nil {
		state = &TargetState{
			Name:               target.Name,
			IsHealthy:          true,
			RestartsInLastHour: []time.Time{},
		}
		e.state.Targets[target.Name] = state
	}
	e.state.mu.Unlock()
	
	// Crear contexto con timeout
	checkCtx, cancel := context.WithTimeout(ctx, time.Duration(e.config.TimeoutSeconds)*time.Second)
	defer cancel()
	
	// Ejecutar todos los checks
	allChecksPassed := true
	for _, checkCfg := range target.Checks {
		checker, err := checks.NewChecker(checkCfg)
		if err != nil {
			e.logger.Error("failed to create checker", logger.Fields(
				"target", target.Name,
				"error", err,
			))
			allChecksPassed = false
			continue
		}
		
		result := checker.Check(checkCtx)
		
		fields := logger.Fields(
			"target", target.Name,
			"check", result.CheckType,
			"result", map[string]bool{"OK": true, "FAIL": false}[fmt.Sprintf("%v", result.Success)],
			"latency_ms", result.Latency.Milliseconds(),
		)
		
		if result.Success {
			e.logger.Debug("check passed", fields)
		} else {
			e.logger.Warn("check failed", logger.Fields(
				"target", target.Name,
				"check", result.CheckType,
				"reason", result.Message,
				"latency_ms", result.Latency.Milliseconds(),
			))
			allChecksPassed = false
		}
	}
	
	// Actualizar estado
	e.state.mu.Lock()
	state.LastCheckTime = time.Now()
	
	if allChecksPassed {
		if !state.IsHealthy {
			e.logger.Info("target recovered", logger.Fields("target", target.Name))
		}
		state.IsHealthy = true
		state.ConsecutiveFailures = 0
		e.state.mu.Unlock()
		return true
	}
	
	// El target falló
	state.ConsecutiveFailures++
	state.IsHealthy = false
	consecutiveFailures := state.ConsecutiveFailures
	e.state.mu.Unlock()
	
	e.logger.Warn("target unhealthy", logger.Fields(
		"target", target.Name,
		"consecutive_failures", consecutiveFailures,
		"threshold", target.Policy.FailThreshold,
	))
	
	// Decidir si ejecutar acción de recuperación
	if consecutiveFailures >= target.Policy.FailThreshold {
		e.executeRecoveryAction(ctx, target, state)
	}
	
	return false
}

// executeRecoveryAction ejecuta la acción de recuperación para un target
func (e *Engine) executeRecoveryAction(ctx context.Context, target config.Target, state *TargetState) {
	e.state.mu.Lock()
	
	// Verificar cooldown
	if !state.LastRestartTime.IsZero() {
		cooldown := time.Duration(target.Policy.RestartCooldownSeconds) * time.Second
		timeSinceLastRestart := time.Since(state.LastRestartTime)
		if timeSinceLastRestart < cooldown {
			e.state.mu.Unlock()
			e.logger.Warn("restart blocked by cooldown", logger.Fields(
				"target", target.Name,
				"cooldown_remaining_seconds", (cooldown - timeSinceLastRestart).Seconds(),
			))
			return
		}
	}
	
	// Verificar rate limit (reinicios por hora)
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	
	// Filtrar reinicios de la última hora
	recentRestarts := []time.Time{}
	for _, t := range state.RestartsInLastHour {
		if t.After(oneHourAgo) {
			recentRestarts = append(recentRestarts, t)
		}
	}
	state.RestartsInLastHour = recentRestarts
	
	if len(state.RestartsInLastHour) >= target.Policy.MaxRestartsPerHour {
		e.state.mu.Unlock()
		e.logger.Error("restart blocked by rate limit", logger.Fields(
			"target", target.Name,
			"restarts_in_last_hour", len(state.RestartsInLastHour),
			"max_restarts_per_hour", target.Policy.MaxRestartsPerHour,
		))
		return
	}
	
	e.state.mu.Unlock()
	
	// Crear acción
	isFirstFailure := state.ConsecutiveFailures == target.Policy.FailThreshold
	action, err := actions.NewAction(target.Action, isFirstFailure)
	if err != nil {
		e.logger.Error("failed to create action", logger.Fields(
			"target", target.Name,
			"error", err,
		))
		return
	}
	
	e.logger.Info("executing recovery action", logger.Fields(
		"target", target.Name,
		"action", action.Name(),
		"consecutive_failures", state.ConsecutiveFailures,
	))
	
	// Ejecutar acción con timeout
	actionCtx, cancel := context.WithTimeout(ctx, time.Duration(e.config.TimeoutSeconds)*time.Second)
	defer cancel()
	
	result := action.Execute(actionCtx)
	
	// Actualizar estado
	e.state.mu.Lock()
	if result.Success {
		state.LastRestartTime = now
		state.RestartsInLastHour = append(state.RestartsInLastHour, now)
		state.ConsecutiveFailures = 0 // Reset after successful restart
		e.state.mu.Unlock()
		
		e.logger.Info("recovery action succeeded", logger.Fields(
			"target", target.Name,
			"action", action.Name(),
			"latency_ms", result.Latency.Milliseconds(),
		))
	} else {
		e.state.mu.Unlock()
		e.logger.Error("recovery action failed", logger.Fields(
			"target", target.Name,
			"action", action.Name(),
			"error", result.Message,
			"latency_ms", result.Latency.Milliseconds(),
		))
	}
}

// Run ejecuta el engine en modo daemon (loop continuo)
func (e *Engine) Run(ctx context.Context) error {
	if e.config.IntervalSeconds <= 0 {
		return fmt.Errorf("interval_seconds must be > 0 for daemon mode")
	}
	
	e.logger.Info("starting watchdog daemon", logger.Fields(
		"interval_seconds", e.config.IntervalSeconds,
		"targets", len(e.config.GetActiveTargets()),
	))
	
	ticker := time.NewTicker(time.Duration(e.config.IntervalSeconds) * time.Second)
	defer ticker.Stop()
	
	// Primera ejecución inmediata
	e.CheckOnce(ctx)
	
	for {
		select {
		case <-ctx.Done():
			e.logger.Info("watchdog stopped", logger.Fields("reason", ctx.Err()))
			return ctx.Err()
		case <-ticker.C:
			e.CheckOnce(ctx)
		}
	}
}
