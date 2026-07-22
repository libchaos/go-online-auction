package ports

import "context"

// EmailMessage is the payload handed to an EmailPort implementation. The
// notification module builds it from a notification request and lets the
// adapter handle the actual transport (SMTP, SES, SendGrid, ...).
type EmailMessage struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
}

// EmailPort abstracts the external action of sending a single email. The
// notification module depends only on this interface so the concrete transport
// (SMTPEmailAdapter today, a managed provider tomorrow) can change without
// touching the domain or application layers.
type EmailPort interface {
	Send(ctx context.Context, msg EmailMessage) error
}
