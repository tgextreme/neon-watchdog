package notifications

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/tgextreme/neon-watchdog/internal/config"
	"github.com/tgextreme/neon-watchdog/internal/logger"
)

// Notifier interfaz para notificadores
type Notifier interface {
	Notify(event Event) error
	Type() string
}

// Event representa un evento a notificar
type Event struct {
	Type      string                 `json:"type"` // failure, recovery, warning
	Target    string                 `json:"target"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Severity  string                 `json:"severity"` // critical, warning, info
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Manager gestiona múltiples notificadores
type Manager struct {
	notifiers []Notifier
	log       *logger.Logger
}

// NewManager crea un nuevo manager de notificaciones
func NewManager(cfg *config.Config, log *logger.Logger) (*Manager, error) {
	m := &Manager{
		notifiers: []Notifier{},
		log:       log,
	}

	if cfg.Notifications == nil || len(cfg.Notifications) == 0 {
		return m, nil
	}

	for _, notifCfg := range cfg.Notifications {
		if !notifCfg.Enabled {
			continue
		}

		var notifier Notifier
		var err error

		switch notifCfg.Type {
		case "email":
			if notifCfg.Email == nil {
				return nil, fmt.Errorf("email notification enabled but no email config provided")
			}
			notifier, err = NewEmailNotifier(notifCfg.Email, log)
		case "webhook":
			if notifCfg.Webhook == nil {
				return nil, fmt.Errorf("webhook notification enabled but no webhook config provided")
			}
			notifier, err = NewWebhookNotifier(notifCfg.Webhook, log)
		case "telegram":
			if notifCfg.Telegram == nil {
				return nil, fmt.Errorf("telegram notification enabled but no telegram config provided")
			}
			notifier, err = NewTelegramNotifier(notifCfg.Telegram, log)
		default:
			return nil, fmt.Errorf("unknown notification type: %s", notifCfg.Type)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create %s notifier: %w", notifCfg.Type, err)
		}

		m.notifiers = append(m.notifiers, notifier)
		log.Info("notification handler enabled", logger.Fields("type", notifCfg.Type))
	}

	return m, nil
}

// Notify envía una notificación a todos los notificadores
func (m *Manager) Notify(event Event) {
	if len(m.notifiers) == 0 {
		return
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	for _, notifier := range m.notifiers {
		go func(n Notifier) {
			if err := n.Notify(event); err != nil {
				m.log.Error("notification failed", logger.Fields(
					"type", n.Type(),
					"target", event.Target,
					"error", err.Error(),
				))
			}
		}(notifier)
	}
}

// EmailNotifier envía notificaciones por email
type EmailNotifier struct {
	cfg *config.EmailConfig
	log *logger.Logger
}

// NewEmailNotifier crea un nuevo notificador de email
func NewEmailNotifier(cfg *config.EmailConfig, log *logger.Logger) (*EmailNotifier, error) {
	if cfg.SMTPHost == "" {
		return nil, fmt.Errorf("smtp_host is required")
	}
	if cfg.SMTPPort == 0 {
		cfg.SMTPPort = 587
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("from address is required")
	}
	if len(cfg.To) == 0 {
		return nil, fmt.Errorf("at least one recipient is required")
	}

	return &EmailNotifier{cfg: cfg, log: log}, nil
}

// Type retorna el tipo de notificador
func (e *EmailNotifier) Type() string {
	return "email"
}

// Notify envía la notificación por email
func (e *EmailNotifier) Notify(event Event) error {
	subject := fmt.Sprintf("[Neon Watchdog] %s - %s", strings.ToUpper(event.Type), event.Target)
	body := fmt.Sprintf(`Neon Watchdog Alert

Type: %s
Target: %s
Severity: %s
Time: %s

Message:
%s

---
This is an automated alert from Neon Watchdog
`, event.Type, event.Target, event.Severity, event.Timestamp.Format(time.RFC3339), event.Message)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", e.cfg.From, strings.Join(e.cfg.To, ", "), subject, body)

	auth := smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPHost, e.cfg.SMTPPort)

	// Si usa TLS, conectar con TLS
	if e.cfg.UseTLS {
		tlsConfig := &tls.Config{
			ServerName: e.cfg.SMTPHost,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("tls dial failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, e.cfg.SMTPHost)
		if err != nil {
			return fmt.Errorf("smtp client failed: %w", err)
		}
		defer client.Quit()

		if e.cfg.Username != "" && e.cfg.Password != "" {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("auth failed: %w", err)
			}
		}

		if err := client.Mail(e.cfg.From); err != nil {
			return fmt.Errorf("mail from failed: %w", err)
		}

		for _, to := range e.cfg.To {
			if err := client.Rcpt(to); err != nil {
				return fmt.Errorf("rcpt to failed: %w", err)
			}
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("data failed: %w", err)
		}

		_, err = w.Write([]byte(msg))
		if err != nil {
			return fmt.Errorf("write failed: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("close failed: %w", err)
		}

		return nil
	}

	// Sin TLS, usar STARTTLS
	err := smtp.SendMail(addr, auth, e.cfg.From, e.cfg.To, []byte(msg))
	if err != nil {
		return fmt.Errorf("send mail failed: %w", err)
	}

	e.log.Debug("email notification sent", logger.Fields(
		"target", event.Target,
		"recipients", len(e.cfg.To),
	))

	return nil
}

// WebhookNotifier envía notificaciones por webhook
type WebhookNotifier struct {
	cfg    *config.WebhookConfig
	log    *logger.Logger
	client *http.Client
}

// NewWebhookNotifier crea un nuevo notificador de webhook
func NewWebhookNotifier(cfg *config.WebhookConfig, log *logger.Logger) (*WebhookNotifier, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	if cfg.Method == "" {
		cfg.Method = "POST"
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 10
	}

	return &WebhookNotifier{
		cfg: cfg,
		log: log,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}, nil
}

// Type retorna el tipo de notificador
func (w *WebhookNotifier) Type() string {
	return "webhook"
}

// Notify envía la notificación por webhook
func (w *WebhookNotifier) Notify(event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequest(w.cfg.Method, w.cfg.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Neon-Watchdog/1.0")

	for key, value := range w.cfg.Headers {
		req.Header.Set(key, value)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	w.log.Debug("webhook notification sent", logger.Fields(
		"target", event.Target,
		"status", resp.StatusCode,
	))

	return nil
}

// TelegramNotifier envía notificaciones por Telegram
type TelegramNotifier struct {
	cfg    *config.TelegramConfig
	log    *logger.Logger
	client *http.Client
}

// NewTelegramNotifier crea un nuevo notificador de Telegram
func NewTelegramNotifier(cfg *config.TelegramConfig, log *logger.Logger) (*TelegramNotifier, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("bot_token is required")
	}
	if cfg.ChatID == "" {
		return nil, fmt.Errorf("chat_id is required")
	}

	return &TelegramNotifier{
		cfg: cfg,
		log: log,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Type retorna el tipo de notificador
func (t *TelegramNotifier) Type() string {
	return "telegram"
}

// Notify envía la notificación por Telegram
func (t *TelegramNotifier) Notify(event Event) error {
	icon := "🔴"
	switch event.Type {
	case "recovery":
		icon = "✅"
	case "warning":
		icon = "⚠️"
	}

	message := fmt.Sprintf("%s *Neon Watchdog Alert*\n\n"+
		"*Type:* %s\n"+
		"*Target:* `%s`\n"+
		"*Severity:* %s\n"+
		"*Time:* %s\n\n"+
		"*Message:*\n%s",
		icon,
		event.Type,
		event.Target,
		event.Severity,
		event.Timestamp.Format("2006-01-02 15:04:05"),
		event.Message,
	)

	payload := map[string]interface{}{
		"chat_id":    t.cfg.ChatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.cfg.BotToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	t.log.Debug("telegram notification sent", logger.Fields(
		"target", event.Target,
		"chat_id", t.cfg.ChatID,
	))

	return nil
}
