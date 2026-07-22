package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"auction/internal/modules/notification/ports"
	"auction/tests/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProcessEmailRequest_DecodesAndSends(t *testing.T) {
	t.Parallel()

	emailMock := mocks.NewMockEmailPort(t)
	consumer := NewEmailDispatchConsumer(nil, emailMock, nil)

	payload := emailRequestedPayload{
		EventID: "evt-1",
		To:      "buyer@example.com",
		Subject: "Withdrawal completed",
		Title:   "Withdrawal completed",
		Body:    "Your withdrawal was paid out.",
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	emailMock.EXPECT().Send(mock.Anything, mock.MatchedBy(func(msg ports.EmailMessage) bool {
		return msg.To == "buyer@example.com" &&
			msg.Subject == "Withdrawal completed" &&
			msg.TextBody == "Your withdrawal was paid out." &&
			strings.Contains(msg.HTMLBody, "<h2>Withdrawal completed</h2>")
	})).Return(nil)

	require.NoError(t, consumer.processEmailRequest(context.Background(), data))
}

func TestProcessEmailRequest_EscapesHTML(t *testing.T) {
	t.Parallel()

	emailMock := mocks.NewMockEmailPort(t)
	consumer := NewEmailDispatchConsumer(nil, emailMock, nil)

	payload := emailRequestedPayload{
		EventID: "evt-2",
		To:      "u@example.com",
		Subject: "Note",
		Title:   "<b>Alert</b>",
		Body:    "amount < 100 & > 0",
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	emailMock.EXPECT().Send(mock.Anything, mock.MatchedBy(func(msg ports.EmailMessage) bool {
		// Angle brackets and ampersands must be HTML-escaped, never emitted raw.
		return !strings.Contains(msg.HTMLBody, "<b>Alert</b>") &&
			strings.Contains(msg.HTMLBody, "&lt;b&gt;Alert&lt;/b&gt;") &&
			strings.Contains(msg.HTMLBody, "amount &lt; 100 &amp; &gt; 0")
	})).Return(nil)

	require.NoError(t, consumer.processEmailRequest(context.Background(), data))
}

func TestProcessEmailRequest_InvalidJSON_ReturnsError(t *testing.T) {
	t.Parallel()

	emailMock := mocks.NewMockEmailPort(t)
	consumer := NewEmailDispatchConsumer(nil, emailMock, nil)

	err := consumer.processEmailRequest(context.Background(), []byte("not-json"))
	require.Error(t, err)
	emailMock.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
}
