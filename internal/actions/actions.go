package actions

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/tgextreme/neon-watchdog/internal/config"
	"github.com/tgextreme/neon-watchdog/internal/logger"
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

// ActionWithHooks envuelve una acción con hooks
type ActionWithHooks struct {
	action Action
	hooks  *config.ActionHooks
	log    *logger.Logger
}

// NewActionWithHooks crea una acción con hooks
func NewActionWithHooks(action Action, hooks *config.ActionHooks, log *logger.Logger) Action {
	if hooks == nil || (len(hooks.BeforeRestart) == 0 && len(hooks.AfterRestart) == 0 && len(hooks.OnFailure) == 0) {
		return action
	}
	return &ActionWithHooks{
		action: action,
		hooks:  hooks,
		log:    log,
	}
}

func (a *ActionWithHooks) Name() string {
	return a.action.Name() + " (with hooks)"
}

func (a *ActionWithHooks) Execute(ctx context.Context) Result {
	// Before hooks
	if len(a.hooks.BeforeRestart) > 0 {
		a.log.Debug("executing before_restart hooks", logger.Fields("count", len(a.hooks.BeforeRestart)))
		for _, cmd := range a.hooks.BeforeRestart {
			if err := runHook(ctx, cmd); err != nil {
				a.log.Warn("before_restart hook failed", logger.Fields("command", cmd, "error", err.Error()))
			}
		}
	}

	// Execute main action
	result := a.action.Execute(ctx)

	// After hooks (solo si success)
	if result.Success && len(a.hooks.AfterRestart) > 0 {
		a.log.Debug("executing after_restart hooks", logger.Fields("count", len(a.hooks.AfterRestart)))
		for _, cmd := range a.hooks.AfterRestart {
			if err := runHook(ctx, cmd); err != nil {
				a.log.Warn("after_restart hook failed", logger.Fields("command", cmd, "error", err.Error()))
			}
		}
	}

	// Failure hooks (solo si failed)
	if !result.Success && len(a.hooks.OnFailure) > 0 {
		a.log.Debug("executing on_failure hooks", logger.Fields("count", len(a.hooks.OnFailure)))
		for _, cmd := range a.hooks.OnFailure {
			if err := runHook(ctx, cmd); err != nil {
				a.log.Warn("on_failure hook failed", logger.Fields("command", cmd, "error", err.Error()))
			}
		}
	}

	return result
}

func runHook(ctx context.Context, cmdStr string) error {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty hook command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
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
func NewAction(actionCfg config.Action, isFirstFailure bool, log *logger.Logger) (Action, error) {
	var baseAction Action
	var err error

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

		baseAction = &ExecAction{
			Command: command,
			Type:    actionType,
		}

	case "systemd":
		if actionCfg.Systemd == nil {
			return nil, fmt.Errorf("systemd action config is nil")
		}

		method := actionCfg.Systemd.Method
		if method == "" {
			method = "restart"
		}

		baseAction = &SystemdAction{
			Unit:   actionCfg.Systemd.Unit,
			Method: method,
		}

	default:
		return nil, fmt.Errorf("unknown action type: %s", actionCfg.Type)
	}

	if err != nil {
		return nil, err
	}

	// Envolver con hooks si están configurados
	return NewActionWithHooks(baseAction, actionCfg.Hooks, log), nil
}
