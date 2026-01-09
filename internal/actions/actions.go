package actions

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/tgextreme/neon-watchdog/internal/config"
)

// Result representa el resultado de ejecutar una acción
type Result struct {
	Success bool
	Message string
	Latency time.Duration
}

// Action es la interfaz que implementan todas las acciones
type Action interface {
	Execute(ctx context.Context) Result
	Name() string
}

// ExecAction ejecuta comandos shell
type ExecAction struct {
	Command []string
	Type    string // "start" o "restart"
}

func (a *ExecAction) Name() string {
	return fmt.Sprintf("exec:%s", a.Type)
}

func (a *ExecAction) Execute(ctx context.Context) Result {
	start := time.Now()
	
	if len(a.Command) == 0 {
		return Result{
			Success: false,
			Message: "empty command",
			Latency: time.Since(start),
		}
	}
	
	cmd := exec.CommandContext(ctx, a.Command[0], a.Command[1:]...)
	output, err := cmd.CombinedOutput()
	latency := time.Since(start)
	
	if err != nil {
		outputStr := strings.TrimSpace(string(output))
		if len(outputStr) > 300 {
			outputStr = outputStr[:300] + "..."
		}
		return Result{
			Success: false,
			Message: fmt.Sprintf("command failed: %v (output: %s)", err, outputStr),
			Latency: latency,
		}
	}
	
	return Result{
		Success: true,
		Message: fmt.Sprintf("command executed successfully"),
		Latency: latency,
	}
}

// SystemdAction ejecuta acciones sobre unidades systemd
type SystemdAction struct {
	Unit   string
	Method string // restart, start, stop
}

func (a *SystemdAction) Name() string {
	return fmt.Sprintf("systemd:%s %s", a.Method, a.Unit)
}

func (a *SystemdAction) Execute(ctx context.Context) Result {
	start := time.Now()
	
	cmd := exec.CommandContext(ctx, "systemctl", a.Method, a.Unit)
	output, err := cmd.CombinedOutput()
	latency := time.Since(start)
	
	if err != nil {
		outputStr := strings.TrimSpace(string(output))
		if len(outputStr) > 300 {
			outputStr = outputStr[:300] + "..."
		}
		return Result{
			Success: false,
			Message: fmt.Sprintf("systemctl %s %s failed: %v (output: %s)", a.Method, a.Unit, err, outputStr),
			Latency: latency,
		}
	}
	
	return Result{
		Success: true,
		Message: fmt.Sprintf("systemctl %s %s succeeded", a.Method, a.Unit),
		Latency: latency,
	}
}

// NewAction crea una acción basada en la configuración
func NewAction(actionCfg config.Action, isFirstFailure bool) (Action, error) {
	switch actionCfg.Type {
	case "exec":
		if actionCfg.Exec == nil {
			return nil, fmt.Errorf("exec action config is nil")
		}
		
		// Decidir si usar start o restart
		var command []string
		actionType := "restart"
		
		if isFirstFailure && len(actionCfg.Exec.Start) > 0 {
			command = actionCfg.Exec.Start
			actionType = "start"
		} else if len(actionCfg.Exec.Restart) > 0 {
			command = actionCfg.Exec.Restart
		} else if len(actionCfg.Exec.Start) > 0 {
			command = actionCfg.Exec.Start
			actionType = "start"
		} else {
			return nil, fmt.Errorf("no command defined in exec action")
		}
		
		return &ExecAction{
			Command: command,
			Type:    actionType,
		}, nil
		
	case "systemd":
		if actionCfg.Systemd == nil {
			return nil, fmt.Errorf("systemd action config is nil")
		}
		
		method := actionCfg.Systemd.Method
		if method == "" {
			method = "restart"
		}
		
		return &SystemdAction{
			Unit:   actionCfg.Systemd.Unit,
			Method: method,
		}, nil
		
	default:
		return nil, fmt.Errorf("unknown action type: %s", actionCfg.Type)
	}
}
