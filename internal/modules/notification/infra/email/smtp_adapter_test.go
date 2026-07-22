package email

import (
	"strings"
	"testing"

	"auction/internal/modules/notification/ports"
	"github.com/stretchr/testify/require"
)

func TestBuildEmailMessage_HeadersAndMultipart(t *testing.T) {
	t.Parallel()

	msg := ports.EmailMessage{
		To:       "buyer@example.com",
		Subject:  "Withdrawal completed",
		HTMLBody: "<h2>Hi</h2><p>Done.</p>",
		TextBody: "Hi\nDone.",
	}

	raw, err := buildEmailMessage(msg, "no-reply@example.com", "Auction")
	require.NoError(t, err)

	content := string(raw)
	require.Contains(t, content, "<no-reply@example.com>")
	require.Contains(t, content, "To: buyer@example.com")
	require.Contains(t, content, "Subject: Withdrawal completed")
	require.Contains(t, content, "Content-Type: multipart/alternative; boundary=")
	require.Contains(t, content, "Content-Type: text/plain; charset=UTF-8")
	require.Contains(t, content, "Content-Type: text/html; charset=UTF-8")
}

func TestBuildEmailMessage_UTF8SubjectIsEncoded(t *testing.T) {
	t.Parallel()

	msg := ports.EmailMessage{
		To:       "u@example.com",
		Subject:  "充值成功",
		HTMLBody: "<p>充值成功</p>",
		TextBody: "充值成功",
	}

	raw, err := buildEmailMessage(msg, "no-reply@example.com", "")
	require.NoError(t, err)

	content := string(raw)
	// The UTF-8 subject must be RFC 2047 encoded, never emitted raw.
	require.NotContains(t, content, "Subject: 充值成功")
	lower := strings.ToLower(content)
	require.True(t, strings.Contains(lower, "=?utf-8?q?") || strings.Contains(lower, "=?utf-8?b?"))
}

