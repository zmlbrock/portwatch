// Package notifier provides implementations of the alert.Notifier interface
// for delivering port change notifications through various channels.
package notifier

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/user/portwatch/internal/state"
)

// EmailConfig holds configuration for the email notifier.
type EmailConfig struct {
	// SMTPHost is the hostname of the SMTP server (e.g. "smtp.gmail.com").
	SMTPHost string
	// SMTPPort is the port of the SMTP server (e.g. 587).
	SMTPPort int
	// Username is the SMTP authentication username.
	Username string
	// Password is the SMTP authentication password.
	Password string
	// From is the sender email address.
	From string
	// To is a list of recipient email addresses.
	To []string
	// Subject is the email subject line. Defaults to "portwatch alert" if empty.
	Subject string
}

// EmailNotifier sends port change alerts via email using SMTP.
type EmailNotifier struct {
	cfg  EmailConfig
	auth smtp.Auth
	addr string
}

// NewEmailNotifier creates a new EmailNotifier with the provided configuration.
// Returns an error if required fields (SMTPHost, From, To) are missing.
func NewEmailNotifier(cfg EmailConfig) (*EmailNotifier, error) {
	if cfg.SMTPHost == "" {
		return nil, fmt.Errorf("email notifier: SMTPHost is required")
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("email notifier: From address is required")
	}
	if len(cfg.To) == 0 {
		return nil, fmt.Errorf("email notifier: at least one To address is required")
	}
	if cfg.SMTPPort == 0 {
		cfg.SMTPPort = 587
	}
	if cfg.Subject == "" {
		cfg.Subject = "portwatch alert"
	}

	var auth smtp.Auth
	if cfg.Username != "" && cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
	}

	return &EmailNotifier{
		cfg:  cfg,
		auth: auth,
		addr: fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
	}, nil
}

// Notify sends an email summarising the provided port changes.
// It is a no-op when changes is empty.
func (e *EmailNotifier) Notify(changes []state.Change) error {
	if len(changes) == 0 {
		return nil
	}

	body := e.buildBody(changes)
	msg := e.buildMessage(body)

	err := smtp.SendMail(e.addr, e.auth, e.cfg.From, e.cfg.To, []byte(msg))
	if err != nil {
		return fmt.Errorf("email notifier: failed to send email: %w", err)
	}
	return nil
}

// buildMessage formats a complete RFC 2822 email message.
func (e *EmailNotifier) buildMessage(body string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", e.cfg.From))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(e.cfg.To, ", ")))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", e.cfg.Subject))
	sb.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	return sb.String()
}

// buildBody produces a human-readable summary of all port changes.
func (e *EmailNotifier) buildBody(changes []state.Change) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("portwatch detected %d port change(s) at %s:\n\n",
		len(changes), time.Now().Format(time.RFC3339)))

	for _, c := range changes {
		switch c.Type {
		case state.ChangeOpened:
			sb.WriteString(fmt.Sprintf("  [OPENED] %s/%d (pid %d)\n",
				strings.ToUpper(c.Port.Protocol), c.Port.Port, c.Port.PID))
		case state.ChangesClosed:
			sb.WriteString(fmt.Sprintf("  [CLOSED] %s/%d (pid %d)\n",
				strings.ToUpper(c.Port.Protocol), c.Port.Port, c.Port.PID))
		default:
			sb.WriteString(fmt.Sprintf("  [CHANGED] %s/%d\n",
				strings.ToUpper(c.Port.Protocol), c.Port.Port))
		}
	}

	sb.WriteString("\n-- portwatch\n")
	return sb.String()
}
