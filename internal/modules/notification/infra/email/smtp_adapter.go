package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/smtp"
	"net/textproto"
	"time"

	"auction/internal/modules/notification/ports"
	"auction/internal/shared/modules/config"
)

// SMTPEmailAdapter is the default EmailPort implementation. It speaks SMTP
// (STARTTLS on 587 or implicit TLS on 465) and renders a multipart/alternative
// message with a plain-text and an HTML body. When the configuration is empty
// (SMTPHost or FromAddress missing) Send is a no-op so local dev and tests do
// not need a reachable mail server.
type SMTPEmailAdapter struct {
	cfg config.Email
}

func NewSMTPEmailAdapter(cfg config.Email) *SMTPEmailAdapter {
	return &SMTPEmailAdapter{cfg: cfg}
}

var _ ports.EmailPort = (*SMTPEmailAdapter)(nil)

// tlsConfig returns a TLS configuration hardened to TLS 1.2 minimum, used for
// both implicit TLS dials and STARTTLS upgrades.
func (adapter *SMTPEmailAdapter) tlsConfig() *tls.Config {
	return &tls.Config{
		ServerName: adapter.cfg.SMTPHost,
		MinVersion: tls.VersionTLS12,
	}
}

// Disabled reports whether the adapter should skip sending. A missing host or
// sender means the operator has not configured email yet.
func (adapter *SMTPEmailAdapter) Disabled() bool {
	return adapter.cfg.SMTPHost == "" || adapter.cfg.FromAddress == ""
}

func (adapter *SMTPEmailAdapter) Send(_ context.Context, msg ports.EmailMessage) error {
	if adapter.Disabled() {
		return nil
	}

	raw, buildErr := buildEmailMessage(msg, adapter.cfg.FromAddress, adapter.cfg.FromName)
	if buildErr != nil {
		return buildErr
	}

	var auth smtp.Auth
	if adapter.cfg.SMTPUsername != "" {
		auth = smtp.PlainAuth("", adapter.cfg.SMTPUsername, adapter.cfg.SMTPPassword, adapter.cfg.SMTPHost)
	}

	addr := fmt.Sprintf("%s:%d", adapter.cfg.SMTPHost, adapter.cfg.SMTPPort)

	if adapter.cfg.UseTLS {
		conn, dialErr := tls.Dial("tcp", addr, adapter.tlsConfig())
		if dialErr != nil {
			return fmt.Errorf("email: tls dial: %w", dialErr)
		}
		defer conn.Close()

		client, clientErr := smtp.NewClient(conn, adapter.cfg.SMTPHost)
		if clientErr != nil {
			return fmt.Errorf("email: smtp client: %w", clientErr)
		}

		return adapter.deliver(client, auth, msg, raw)
	}

	client, dialErr := smtp.Dial(addr)
	if dialErr != nil {
		return fmt.Errorf("email: smtp dial: %w", dialErr)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if tlsErr := client.StartTLS(adapter.tlsConfig()); tlsErr != nil {
			return fmt.Errorf("email: starttls: %w", tlsErr)
		}
	}

	return adapter.deliver(client, auth, msg, raw)
}

func (adapter *SMTPEmailAdapter) deliver(
	client *smtp.Client,
	auth smtp.Auth,
	msg ports.EmailMessage,
	raw []byte,
) error {
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("email: auth: %w", err)
		}
	}

	if err := client.Mail(adapter.cfg.FromAddress); err != nil {
		return fmt.Errorf("email: mail from: %w", err)
	}

	if err := client.Rcpt(msg.To); err != nil {
		return fmt.Errorf("email: rcpt: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("email: data: %w", err)
	}

	if _, err = writer.Write(raw); err != nil {
		return fmt.Errorf("email: write: %w", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("email: close: %w", err)
	}

	return nil
}

// buildEmailMessage renders a multipart/alternative RFC 5322 message with a
// quoted-printable-safe UTF-8 subject and both a plain-text and an HTML part.
func buildEmailMessage(msg ports.EmailMessage, fromAddress, fromName string) ([]byte, error) {
	from := fromAddress
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("UTF-8", fromName), fromAddress)
	}

	var buf bytes.Buffer
	buf.WriteString("From: " + from + "\r\n")
	buf.WriteString("To: " + msg.To + "\r\n")
	buf.WriteString("Subject: " + mime.QEncoding.Encode("UTF-8", msg.Subject) + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Date: " + time.Now().UTC().Format(time.RFC1123Z) + "\r\n")

	writer := multipart.NewWriter(&buf)
	buf.WriteString("Content-Type: multipart/alternative; boundary=" + writer.Boundary() + "\r\n")
	buf.WriteString("\r\n")

	if err := writePart(writer, "text/plain; charset=UTF-8", msg.TextBody); err != nil {
		return nil, err
	}

	if err := writePart(writer, "text/html; charset=UTF-8", msg.HTMLBody); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writePart(writer *multipart.Writer, contentType, body string) error {
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", contentType)
	header.Set("Content-Transfer-Encoding", "quoted-printable")

	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}

	encoder := quotedprintable.NewWriter(part)
	if _, err = encoder.Write([]byte(body)); err != nil {
		return err
	}

	return encoder.Close()
}
