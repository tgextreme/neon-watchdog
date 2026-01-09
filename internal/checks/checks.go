package checks

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/tgextreme/neon-watchdog/internal/config"
)

// Result representa el resultado de un check
type Result struct {
	Success  bool
	Message  string
	Latency  time.Duration
	CheckType string
}

// Checker es la interfaz que implementan todos los checkers
type Checker interface {
	Check(ctx context.Context) Result
	Name() string
}

// ProcessNameChecker verifica si un proceso está corriendo
type ProcessNameChecker struct {
	ProcessName string
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
		return &ProcessNameChecker{ProcessName: check.ProcessName}, nil
	case "pid_file":
		return &PidFileChecker{PidFile: check.PidFile}, nil
	case "tcp_port":
		return &TcpPortChecker{Address: check.TcpPort}, nil
	case "command":
		return &CommandChecker{Command: check.Command}, nil
	default:
		return nil, fmt.Errorf("unknown check type: %s", check.Type)
	}
}
