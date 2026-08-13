package alerting

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/store"
)

func sendEmail(ctx context.Context, channel store.AlertChannel, message Message) error {
	var cfg EmailConfig
	_ = json.Unmarshal(channel.Config, &cfg)
	mode := cfg.TLSMode
	if mode == "" {
		mode = "starttls"
	}
	address := net.JoinHostPort(cfg.Host, fmt.Sprint(cfg.Port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	var connection net.Conn
	var err error
	if mode == "implicit" {
		connection, err = (&tls.Dialer{NetDialer: dialer, Config: &tls.Config{MinVersion: tls.VersionTLS12, ServerName: cfg.Host}}).DialContext(ctx, "tcp", address)
	} else {
		connection, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return err
	}
	defer connection.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = connection.SetDeadline(deadline)
	} else {
		_ = connection.SetDeadline(time.Now().Add(10 * time.Second))
	}
	client, err := smtp.NewClient(connection, cfg.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	if mode == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("SMTP server does not advertise STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{MinVersion: tls.VersionTLS12, ServerName: cfg.Host}); err != nil {
			return err
		}
	}
	if cfg.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)); err != nil {
			return err
		}
	}
	from, _ := mail.ParseAddress(cfg.From)
	if err := client.Mail(from.Address); err != nil {
		return err
	}
	to := make([]string, 0, len(cfg.To))
	for _, raw := range cfg.To {
		recipient, _ := mail.ParseAddress(raw)
		to = append(to, recipient.Address)
		if err := client.Rcpt(recipient.Address); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := io.WriteString(writer, emailBody(from.Address, to, message)); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func emailBody(from string, to []string, message Message) string {
	subject := "[Honeynet][" + strings.ToUpper(message.Alert.Level) + "] " + message.Alert.Title
	encodedSubject := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
	return strings.Join([]string{
		"From: " + from,
		"To: " + strings.Join(to, ", "),
		"Subject: " + encodedSubject,
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		strings.ReplaceAll(message.PlainText(), "\n", "\r\n"),
		"",
	}, "\r\n")
}
