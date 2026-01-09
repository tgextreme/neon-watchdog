package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/tgextreme/neon-watchdog/internal/config"
	"github.com/tgextreme/neon-watchdog/internal/logger"
)

// Collector gestiona métricas y las expone en formato Prometheus
type Collector struct {
	cfg        *config.MetricsConfig
	log        *logger.Logger
	mu         sync.RWMutex
	checks     map[string]*CheckMetrics
	recoveries map[string]int64
	uptime     time.Time
}

// CheckMetrics métricas por target
type CheckMetrics struct {
	Healthy             bool
	LastCheckTime       time.Time
	TotalChecks         int64
	FailedChecks        int64
	SuccessfulChecks    int64
	LastCheckDuration   time.Duration
	ConsecutiveFailures int
}

// NewCollector crea un nuevo collector de métricas
func NewCollector(cfg *config.MetricsConfig, log *logger.Logger) *Collector {
	if cfg == nil {
		cfg = &config.MetricsConfig{Enabled: false}
	}

	if cfg.Path == "" {
		cfg.Path = "/metrics"
	}

	return &Collector{
		cfg:        cfg,
		log:        log,
		checks:     make(map[string]*CheckMetrics),
		recoveries: make(map[string]int64),
		uptime:     time.Now(),
	}
}

// Start inicia el servidor de métricas
func (c *Collector) Start() error {
	if !c.cfg.Enabled {
		return nil
	}

	if c.cfg.Port == 0 {
		c.cfg.Port = 9090
	}

	http.HandleFunc(c.cfg.Path, c.handleMetrics)

	addr := fmt.Sprintf(":%d", c.cfg.Port)
	c.log.Info("metrics server starting", logger.Fields(
		"port", c.cfg.Port,
		"path", c.cfg.Path,
	))

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			c.log.Error("metrics server failed", logger.Fields("error", err.Error()))
		}
	}()

	return nil
}

// RecordCheck registra una ejecución de check
func (c *Collector) RecordCheck(target string, healthy bool, duration time.Duration, consecutiveFailures int) {
	if !c.cfg.Enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	m, ok := c.checks[target]
	if !ok {
		m = &CheckMetrics{}
		c.checks[target] = m
	}

	m.Healthy = healthy
	m.LastCheckTime = time.Now()
	m.LastCheckDuration = duration
	m.TotalChecks++
	m.ConsecutiveFailures = consecutiveFailures

	if healthy {
		m.SuccessfulChecks++
	} else {
		m.FailedChecks++
	}
}

// RecordRecovery registra una recuperación exitosa
func (c *Collector) RecordRecovery(target string) {
	if !c.cfg.Enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.recoveries[target]++
}

// handleMetrics maneja el endpoint de métricas
func (c *Collector) handleMetrics(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Uptime
	fmt.Fprintf(w, "# HELP neon_watchdog_uptime_seconds Time since watchdog started\n")
	fmt.Fprintf(w, "# TYPE neon_watchdog_uptime_seconds gauge\n")
	fmt.Fprintf(w, "neon_watchdog_uptime_seconds %d\n\n", int64(time.Since(c.uptime).Seconds()))

	// Target health
	fmt.Fprintf(w, "# HELP neon_watchdog_target_healthy Target health status (1=healthy, 0=unhealthy)\n")
	fmt.Fprintf(w, "# TYPE neon_watchdog_target_healthy gauge\n")
	for target, metrics := range c.checks {
		health := 0
		if metrics.Healthy {
			health = 1
		}
		fmt.Fprintf(w, "neon_watchdog_target_healthy{target=\"%s\"} %d\n", target, health)
	}
	fmt.Fprintln(w)

	// Total checks
	fmt.Fprintf(w, "# HELP neon_watchdog_checks_total Total number of checks performed\n")
	fmt.Fprintf(w, "# TYPE neon_watchdog_checks_total counter\n")
	for target, metrics := range c.checks {
		fmt.Fprintf(w, "neon_watchdog_checks_total{target=\"%s\"} %d\n", target, metrics.TotalChecks)
	}
	fmt.Fprintln(w)

	// Failed checks
	fmt.Fprintf(w, "# HELP neon_watchdog_checks_failed_total Total number of failed checks\n")
	fmt.Fprintf(w, "# TYPE neon_watchdog_checks_failed_total counter\n")
	for target, metrics := range c.checks {
		fmt.Fprintf(w, "neon_watchdog_checks_failed_total{target=\"%s\"} %d\n", target, metrics.FailedChecks)
	}
	fmt.Fprintln(w)

	// Successful checks
	fmt.Fprintf(w, "# HELP neon_watchdog_checks_successful_total Total number of successful checks\n")
	fmt.Fprintf(w, "# TYPE neon_watchdog_checks_successful_total counter\n")
	for target, metrics := range c.checks {
		fmt.Fprintf(w, "neon_watchdog_checks_successful_total{target=\"%s\"} %d\n", target, metrics.SuccessfulChecks)
	}
	fmt.Fprintln(w)

	// Check duration
	fmt.Fprintf(w, "# HELP neon_watchdog_check_duration_seconds Duration of last check in seconds\n")
	fmt.Fprintf(w, "# TYPE neon_watchdog_check_duration_seconds gauge\n")
	for target, metrics := range c.checks {
		fmt.Fprintf(w, "neon_watchdog_check_duration_seconds{target=\"%s\"} %.3f\n",
			target, metrics.LastCheckDuration.Seconds())
	}
	fmt.Fprintln(w)

	// Consecutive failures
	fmt.Fprintf(w, "# HELP neon_watchdog_consecutive_failures Current consecutive failures for target\n")
	fmt.Fprintf(w, "# TYPE neon_watchdog_consecutive_failures gauge\n")
	for target, metrics := range c.checks {
		fmt.Fprintf(w, "neon_watchdog_consecutive_failures{target=\"%s\"} %d\n", target, metrics.ConsecutiveFailures)
	}
	fmt.Fprintln(w)

	// Recoveries
	fmt.Fprintf(w, "# HELP neon_watchdog_recoveries_total Total number of successful recoveries\n")
	fmt.Fprintf(w, "# TYPE neon_watchdog_recoveries_total counter\n")
	for target, count := range c.recoveries {
		fmt.Fprintf(w, "neon_watchdog_recoveries_total{target=\"%s\"} %d\n", target, count)
	}
	fmt.Fprintln(w)

	// Last check timestamp
	fmt.Fprintf(w, "# HELP neon_watchdog_last_check_timestamp_seconds Timestamp of last check\n")
	fmt.Fprintf(w, "# TYPE neon_watchdog_last_check_timestamp_seconds gauge\n")
	for target, metrics := range c.checks {
		fmt.Fprintf(w, "neon_watchdog_last_check_timestamp_seconds{target=\"%s\"} %d\n",
			target, metrics.LastCheckTime.Unix())
	}
}
