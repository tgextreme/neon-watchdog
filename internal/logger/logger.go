package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Level representa el nivel de log
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

// Logger maneja el logging estructurado
type Logger struct {
	level  Level
	output io.Writer
}

// New crea un nuevo logger
func New(levelStr string, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}

	level := INFO
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		level = DEBUG
	case "INFO":
		level = INFO
	case "WARN":
		level = WARN
	case "ERROR":
		level = ERROR
	}

	return &Logger{
		level:  level,
		output: output,
	}
}

// log escribe un mensaje con formato estructurado
func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	levelName := levelNames[level]

	// Formato: timestamp level=LEVEL msg="message" key1=value1 key2=value2
	parts := []string{
		timestamp,
		fmt.Sprintf("level=%s", levelName),
		fmt.Sprintf("msg=%q", msg),
	}

	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, formatValue(v)))
	}

	fmt.Fprintln(l.output, strings.Join(parts, " "))
}

// formatValue formatea un valor para el log
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		if strings.ContainsAny(val, " \t\n\"") {
			return fmt.Sprintf("%q", val)
		}
		return val
	case error:
		return fmt.Sprintf("%q", val.Error())
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Debug registra un mensaje de debug
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.log(DEBUG, msg, fields)
}

// Info registra un mensaje informativo
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log(INFO, msg, fields)
}

// Warn registra un warning
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.log(WARN, msg, fields)
}

// Error registra un error
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log(ERROR, msg, fields)
}

// WithFields es un helper para crear logs con campos
func Fields(kv ...interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := 0; i < len(kv)-1; i += 2 {
		if key, ok := kv[i].(string); ok {
			fields[key] = kv[i+1]
		}
	}
	return fields
}
