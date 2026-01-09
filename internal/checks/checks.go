package checks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/tgextreme/neon-watchdog/internal/config"
)

// Result representa el resultado de un check
type Result struct {
	Success   bool
	Message   string
	Latency   time.Duration
	CheckType string
}

// Checker es la interfaz que implementan todos los checkers
type Checker interface {
	Check(ctx context.Context) Result
	Name() string
}

// ProcessNameChecker verifica si un proceso está corriendo
type ProcessNameChecker struct {
	ProcessName     string
	IgnoreExitCodes []int
}

func (c *ProcessNameChecker) Name() string {
	return fmt.Sprintf("process_name:%s", c.ProcessName)
}

func (c *ProcessNameChecker) Check(ctx context.Context) Result {
	start := time.Now()

	// Usar pgrep para buscar el proceso
	cmd := exec.CommandContext(ctx, "pgrep", "-x", c.ProcessName)
	output, err := cmd.Output()

	latency := time.Since(start)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return Result{
				Success:   false,
				Message:   fmt.Sprintf("process '%s' not found", c.ProcessName),
				Latency:   latency,
				CheckType: "process_name",
			}
		}
		return Result{
			Success:   false,
			Message:   fmt.Sprintf("pgrep error: %v", err),
			Latency:   latency,
			CheckType: "process_name",
		}
	}

	// Si hay PIDs, verificar exit codes si están configurados
	// (para graceful shutdown detection, verificamos /proc/PID/stat)
	pids := strings.TrimSpace(string(output))
	return Result{
		Success:   true,
		Message:   fmt.Sprintf("process '%s' found (PIDs: %s)", c.ProcessName, pids),
		Latency:   latency,
		CheckType: "process_name",
	}
}

// PidFileChecker verifica si el PID de un archivo existe
type PidFileChecker struct {
	PidFile string
}

func (c *PidFileChecker) Name() string {
	return fmt.Sprintf("pid_file:%s", c.PidFile)
}

func (c *PidFileChecker) Check(ctx context.Context) Result {
	start := time.Now()

	data, err := os.ReadFile(c.PidFile)
	if err != nil {
		return Result{
			Success:   false,
			Message:   fmt.Sprintf("cannot read pid file: %v", err),
			Latency:   time.Since(start),
			CheckType: "pid_file",
		}
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return Result{
			Success:   false,
			Message:   fmt.Sprintf("invalid PID in file: %s", pidStr),
			Latency:   time.Since(start),
			CheckType: "pid_file",
		}
	}

	// Verificar si el proceso existe
	process, err := os.FindProcess(pid)
	if err != nil {
		return Result{
			Success:   false,
			Message:   fmt.Sprintf("process %d not found", pid),
			Latency:   time.Since(start),
			CheckType: "pid_file",
		}
	}

	// En Unix, FindProcess siempre tiene éxito, así que probamos con Signal 0
	err = process.Signal(os.Signal(nil))
	if err != nil {
		return Result{
			Success:   false,
			Message:   fmt.Sprintf("process %d not running", pid),
			Latency:   time.Since(start),
			CheckType: "pid_file",
		}
	}

	return Result{
		Success:   true,
		Message:   fmt.Sprintf("process %d is running", pid),
		Latency:   time.Since(start),
		CheckType: "pid_file",
	}
}

// TcpPortChecker verifica si un puerto TCP está escuchando
type TcpPortChecker struct {
	Address string
}

func (c *TcpPortChecker) Name() string {
	return fmt.Sprintf("tcp_port:%s", c.Address)
}

func (c *TcpPortChecker) Check(ctx context.Context) Result {
	start := time.Now()

	// Si no tiene host, asumir localhost
	address := c.Address
	if !strings.Contains(address, ":") {
		address = "127.0.0.1:" + address
	} else if strings.HasPrefix(address, ":") {
		address = "127.0.0.1" + address
	}

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", address)
	latency := time.Since(start)

	if err != nil {
		return Result{
			Success:   false,
			Message:   fmt.Sprintf("connection to %s failed: %v", address, err),
			Latency:   latency,
			CheckType: "tcp_port",
		}
	}

	conn.Close()
	return Result{
		Success:   true,
		Message:   fmt.Sprintf("connection to %s successful", address),
		Latency:   latency,
		CheckType: "tcp_port",
	}
}

// CommandChecker ejecuta un comando y verifica su exit code
type CommandChecker struct {
	Command []string
}

func (c *CommandChecker) Name() string {
	return fmt.Sprintf("command:%s", strings.Join(c.Command, " "))
}

func (c *CommandChecker) Check(ctx context.Context) Result {
	start := time.Now()

	if len(c.Command) == 0 {
		return Result{
			Success:   false,
			Message:   "empty command",
			Latency:   time.Since(start),
			CheckType: "command",
		}
	}

	cmd := exec.CommandContext(ctx, c.Command[0], c.Command[1:]...)
	output, err := cmd.CombinedOutput()
	latency := time.Since(start)

	if err != nil {
		outputStr := strings.TrimSpace(string(output))
		if len(outputStr) > 200 {
			outputStr = outputStr[:200] + "..."
		}
		return Result{
			Success:   false,
			Message:   fmt.Sprintf("command failed: %v (output: %s)", err, outputStr),
			Latency:   latency,
			CheckType: "command",
		}
	}

	return Result{
		Success:   true,
		Message:   "command succeeded",
		Latency:   latency,
		CheckType: "command",
	}
}

// NewChecker crea un checker basado en la configuración
func NewChecker(check config.Check) (Checker, error) {
	switch check.Type {
	case "process_name":
		return &ProcessNameChecker{
			ProcessName:     check.ProcessName,
			IgnoreExitCodes: check.IgnoreExitCodes,
		}, nil
	case "pid_file":
		return &PidFileChecker{PidFile: check.PidFile}, nil
	case "tcp_port":
		return &TcpPortChecker{Address: check.TcpPort}, nil
	case "command":
		return &CommandChecker{Command: check.Command}, nil
	case "http":
		return NewHTTPChecker(check.HTTP)
	case "script":
		return NewScriptChecker(check.Script)
	case "logic":
		return NewLogicChecker(check.Logic, check.Checks)
	default:
		return nil, fmt.Errorf("unknown check type: %s", check.Type)
	}
}

// HTTPChecker verifica un endpoint HTTP
type HTTPChecker struct {
	URL            string
	Method         string
	ExpectedStatus int
	Headers        map[string]string
	Body           string
	Timeout        time.Duration
	client         *http.Client
}

// NewHTTPChecker crea un nuevo HTTP checker
func NewHTTPChecker(cfg *config.HTTPCheck) (*HTTPChecker, error) {
	if cfg == nil || cfg.URL == "" {
		return nil, fmt.Errorf("http check requires url")
	}

	method := cfg.Method
	if method == "" {
		method = "GET"
	}

	expectedStatus := cfg.ExpectedStatus
	if expectedStatus == 0 {
		expectedStatus = 200
	}

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &HTTPChecker{
		URL:            cfg.URL,
		Method:         method,
		ExpectedStatus: expectedStatus,
		Headers:        cfg.Headers,
		Body:           cfg.Body,
		Timeout:        timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *HTTPChecker) Name() string {
	return fmt.Sprintf("http:%s", c.URL)
}

func (c *HTTPChecker) Check(ctx context.Context) Result {
	start := time.Now()

	var bodyReader io.Reader
	if c.Body != "" {
		bodyReader = bytes.NewBufferString(c.Body)
	}

	req, err := http.NewRequestWithContext(ctx, c.Method, c.URL, bodyReader)
	if err != nil {
		return Result{
			Success:   false,
			Message:   fmt.Sprintf("failed to create request: %v", err),
			Latency:   time.Since(start),
			CheckType: "http",
		}
	}

	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return Result{
			Success:   false,
			Message:   fmt.Sprintf("http request failed: %v", err),
			Latency:   latency,
			CheckType: "http",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != c.ExpectedStatus {
		return Result{
			Success:   false,
			Message:   fmt.Sprintf("unexpected status code: got %d, expected %d", resp.StatusCode, c.ExpectedStatus),
			Latency:   latency,
			CheckType: "http",
		}
	}

	return Result{
		Success:   true,
		Message:   fmt.Sprintf("http check passed (status: %d)", resp.StatusCode),
		Latency:   latency,
		CheckType: "http",
	}
}

// ScriptChecker ejecuta un script personalizado
type ScriptChecker struct {
	Path             string
	Args             []string
	SuccessExitCodes []int
	WarningExitCodes []int
}

// NewScriptChecker crea un nuevo script checker
func NewScriptChecker(cfg *config.ScriptCheck) (*ScriptChecker, error) {
	if cfg == nil || cfg.Path == "" {
		return nil, fmt.Errorf("script check requires path")
	}

	successCodes := cfg.SuccessExitCodes
	if len(successCodes) == 0 {
		successCodes = []int{0}
	}

	return &ScriptChecker{
		Path:             cfg.Path,
		Args:             cfg.Args,
		SuccessExitCodes: successCodes,
		WarningExitCodes: cfg.WarningExitCodes,
	}, nil
}

func (c *ScriptChecker) Name() string {
	return fmt.Sprintf("script:%s", c.Path)
}

func (c *ScriptChecker) Check(ctx context.Context) Result {
	start := time.Now()

	cmd := exec.CommandContext(ctx, c.Path, c.Args...)
	output, err := cmd.CombinedOutput()
	latency := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return Result{
				Success:   false,
				Message:   fmt.Sprintf("failed to execute script: %v", err),
				Latency:   latency,
				CheckType: "script",
			}
		}
	}

	// Verificar si es exit code de success
	for _, code := range c.SuccessExitCodes {
		if exitCode == code {
			return Result{
				Success:   true,
				Message:   fmt.Sprintf("script succeeded (exit code: %d)", exitCode),
				Latency:   latency,
				CheckType: "script",
			}
		}
	}

	// Verificar si es exit code de warning (no falla pero registra)
	for _, code := range c.WarningExitCodes {
		if exitCode == code {
			return Result{
				Success:   true,
				Message:   fmt.Sprintf("script warning (exit code: %d): %s", exitCode, strings.TrimSpace(string(output))),
				Latency:   latency,
				CheckType: "script",
			}
		}
	}

	// Fallo
	outputStr := strings.TrimSpace(string(output))
	if len(outputStr) > 200 {
		outputStr = outputStr[:200] + "..."
	}
	return Result{
		Success:   false,
		Message:   fmt.Sprintf("script failed (exit code: %d): %s", exitCode, outputStr),
		Latency:   latency,
		CheckType: "script",
	}
}

// LogicChecker combina múltiples checks con lógica AND/OR
type LogicChecker struct {
	Logic    string // AND o OR
	Checkers []Checker
}

// NewLogicChecker crea un nuevo logic checker
func NewLogicChecker(logic string, checks []config.Check) (*LogicChecker, error) {
	if logic != "AND" && logic != "OR" {
		return nil, fmt.Errorf("logic must be AND or OR, got: %s", logic)
	}

	if len(checks) == 0 {
		return nil, fmt.Errorf("logic checker requires at least one check")
	}

	checkers := make([]Checker, 0, len(checks))
	for _, check := range checks {
		checker, err := NewChecker(check)
		if err != nil {
			return nil, fmt.Errorf("failed to create checker: %w", err)
		}
		checkers = append(checkers, checker)
	}

	return &LogicChecker{
		Logic:    logic,
		Checkers: checkers,
	}, nil
}

func (c *LogicChecker) Name() string {
	names := make([]string, len(c.Checkers))
	for i, checker := range c.Checkers {
		names[i] = checker.Name()
	}
	return fmt.Sprintf("logic:%s[%s]", c.Logic, strings.Join(names, ","))
}

func (c *LogicChecker) Check(ctx context.Context) Result {
	start := time.Now()

	results := make([]Result, len(c.Checkers))
	for i, checker := range c.Checkers {
		results[i] = checker.Check(ctx)
	}

	latency := time.Since(start)

	if c.Logic == "AND" {
		// Todos deben pasar
		for _, result := range results {
			if !result.Success {
				messages := make([]string, len(results))
				for i, r := range results {
					status := "✓"
					if !r.Success {
						status = "✗"
					}
					messages[i] = fmt.Sprintf("%s %s", status, r.Message)
				}
				return Result{
					Success:   false,
					Message:   fmt.Sprintf("AND logic failed: %s", strings.Join(messages, "; ")),
					Latency:   latency,
					CheckType: "logic",
				}
			}
		}
		return Result{
			Success:   true,
			Message:   "all AND checks passed",
			Latency:   latency,
			CheckType: "logic",
		}
	}

	// OR: al menos uno debe pasar
	for _, result := range results {
		if result.Success {
			return Result{
				Success:   true,
				Message:   fmt.Sprintf("OR logic passed: %s", result.Message),
				Latency:   latency,
				CheckType: "logic",
			}
		}
	}

	messages := make([]string, len(results))
	for i, r := range results {
		messages[i] = r.Message
	}
	return Result{
		Success:   false,
		Message:   fmt.Sprintf("OR logic failed (all checks failed): %s", strings.Join(messages, "; ")),
		Latency:   latency,
		CheckType: "logic",
	}
}
