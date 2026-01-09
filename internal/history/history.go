package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tgextreme/neon-watchdog/internal/config"
	"github.com/tgextreme/neon-watchdog/internal/logger"
)

// Event representa un evento histórico
type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"` // check_failed, check_passed, recovery_success, recovery_failed
	Target    string                 `json:"target"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// TargetStats estadísticas agregadas por target
type TargetStats struct {
	TotalChecks         int64     `json:"total_checks"`
	FailedChecks        int64     `json:"failed_checks"`
	SuccessfulChecks    int64     `json:"successful_checks"`
	TotalRecoveries     int64     `json:"total_recoveries"`
	FailedRecoveries    int64     `json:"failed_recoveries"`
	LastCheckTime       time.Time `json:"last_check_time"`
	LastFailureTime     time.Time `json:"last_failure_time,omitempty"`
	LastRecoveryTime    time.Time `json:"last_recovery_time,omitempty"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
}

// History gestiona el historial de eventos
type History struct {
	cfg       *config.HistoryConfig
	log       *logger.Logger
	mu        sync.RWMutex
	events    []Event
	stats     map[string]*TargetStats
	stateFile string
}

// State representa el estado completo persistente
type State struct {
	Events []Event                 `json:"events"`
	Stats  map[string]*TargetStats `json:"stats"`
}

// NewHistory crea un nuevo gestor de historial
func NewHistory(cfg *config.HistoryConfig, stateFile string, log *logger.Logger) *History {
	if cfg == nil {
		cfg = &config.HistoryConfig{
			MaxEntries:     1000,
			RetentionHours: 168, // 7 días
		}
	}

	if cfg.MaxEntries == 0 {
		cfg.MaxEntries = 1000
	}

	if cfg.RetentionHours == 0 {
		cfg.RetentionHours = 168
	}

	h := &History{
		cfg:       cfg,
		log:       log,
		events:    make([]Event, 0),
		stats:     make(map[string]*TargetStats),
		stateFile: stateFile,
	}

	// Cargar estado anterior si existe
	if err := h.Load(); err != nil {
		log.Warn("failed to load history", logger.Fields("error", err.Error()))
	}

	return h
}

// RecordEvent registra un nuevo evento
func (h *History) RecordEvent(eventType, target, message string, details map[string]interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	event := Event{
		Timestamp: time.Now(),
		Type:      eventType,
		Target:    target,
		Message:   message,
		Details:   details,
	}

	h.events = append(h.events, event)

	// Actualizar estadísticas
	stats, ok := h.stats[target]
	if !ok {
		stats = &TargetStats{}
		h.stats[target] = stats
	}

	stats.LastCheckTime = event.Timestamp

	switch eventType {
	case "check_failed":
		stats.FailedChecks++
		stats.TotalChecks++
		stats.ConsecutiveFailures++
		stats.LastFailureTime = event.Timestamp
	case "check_passed":
		stats.SuccessfulChecks++
		stats.TotalChecks++
		stats.ConsecutiveFailures = 0
	case "recovery_success":
		stats.TotalRecoveries++
		stats.LastRecoveryTime = event.Timestamp
		stats.ConsecutiveFailures = 0
	case "recovery_failed":
		stats.FailedRecoveries++
	}

	// Limpiar eventos antiguos
	h.cleanOldEvents()

	// Limitar número de eventos
	if len(h.events) > h.cfg.MaxEntries {
		h.events = h.events[len(h.events)-h.cfg.MaxEntries:]
	}
}

// cleanOldEvents elimina eventos más antiguos que RetentionHours
func (h *History) cleanOldEvents() {
	cutoff := time.Now().Add(-time.Duration(h.cfg.RetentionHours) * time.Hour)

	newEvents := make([]Event, 0)
	for _, event := range h.events {
		if event.Timestamp.After(cutoff) {
			newEvents = append(newEvents, event)
		}
	}
	h.events = newEvents
}

// GetEvents retorna todos los eventos
func (h *History) GetEvents(target string, limit int) []Event {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if target == "" {
		// Retornar los últimos N eventos
		if limit > 0 && limit < len(h.events) {
			return h.events[len(h.events)-limit:]
		}
		return h.events
	}

	// Filtrar por target
	filtered := make([]Event, 0)
	for _, event := range h.events {
		if event.Target == target {
			filtered = append(filtered, event)
		}
	}

	if limit > 0 && limit < len(filtered) {
		return filtered[len(filtered)-limit:]
	}
	return filtered
}

// GetStats retorna estadísticas de un target
func (h *History) GetStats(target string) *TargetStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if target == "" {
		// Retornar estadísticas agregadas
		total := &TargetStats{}
		for _, stats := range h.stats {
			total.TotalChecks += stats.TotalChecks
			total.FailedChecks += stats.FailedChecks
			total.SuccessfulChecks += stats.SuccessfulChecks
			total.TotalRecoveries += stats.TotalRecoveries
			total.FailedRecoveries += stats.FailedRecoveries
		}
		return total
	}

	stats, ok := h.stats[target]
	if !ok {
		return &TargetStats{}
	}

	// Retornar copia
	statsCopy := *stats
	return &statsCopy
}

// GetAllStats retorna todas las estadísticas
func (h *History) GetAllStats() map[string]*TargetStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Retornar copia
	copy := make(map[string]*TargetStats)
	for k, v := range h.stats {
		statsCopy := *v
		copy[k] = &statsCopy
	}
	return copy
}

// Save persiste el historial al disco
func (h *History) Save() error {
	if h.stateFile == "" {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// Crear directorio si no existe
	dir := filepath.Dir(h.stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	state := State{
		Events: h.events,
		Stats:  h.stats,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Escribir a archivo temporal primero
	tmpFile := h.stateFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Renombrar (atómico)
	if err := os.Rename(tmpFile, h.stateFile); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// Load carga el historial del disco
func (h *History) Load() error {
	if h.stateFile == "" {
		return nil
	}

	data, err := os.ReadFile(h.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No hay estado previo
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.events = state.Events
	h.stats = state.Stats

	if h.events == nil {
		h.events = make([]Event, 0)
	}
	if h.stats == nil {
		h.stats = make(map[string]*TargetStats)
	}

	h.log.Info("history loaded", logger.Fields(
		"events", len(h.events),
		"targets", len(h.stats),
	))

	return nil
}
