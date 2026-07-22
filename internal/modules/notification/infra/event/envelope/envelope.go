package envelope

import (
	"encoding/json"
	"strconv"
	"time"

	"auction/internal/modules/notification/domain/event"
	"auction/internal/modules/notification/ports"
)

const (
	schemaVersion      = 1
	subjectPrefix      = "notification.evt."
	subjectStreamRoot  = "notification.evt"
	minSubjectParts    = 3
	userIDSubjectIndex = 2

	// SubjectEmailRequested is the NATS subject an email dispatch consumer
	// subscribes to. It sits under notification.evt.> so the shared outbox relay
	// publishes it to the NOTIFICATION_EVENTS stream.
	SubjectEmailRequested   = "notification.evt.email.requested"
	EmailRequestedEventType = "email_requested"
)

// NotificationCreatedPayload is the JSON body published to notification.evt.{userID}.
// The SSE realtime hub decodes it and streams it to the recipient's clients.
type NotificationCreatedPayload struct {
	EventID        string          `json:"event_id"`
	NotificationID uint64          `json:"notification_id"`
	UserID         uint64          `json:"user_id"`
	Category       string          `json:"category"`
	Type           string          `json:"type"`
	Title          string          `json:"title"`
	Body           string          `json:"body"`
	Payload        json.RawMessage `json:"payload"`
	CreatedAt      string          `json:"created_at"`
}

// BuildSubject renders the per-user subject notification.evt.{userID} so the hub
// can filter and route each event to the owning user's connected SSE clients.
func BuildSubject(userID uint64) string {
	return subjectPrefix + strconv.FormatUint(userID, 10)
}

func ToNotificationCreatedOutboxEvent(domainEvent event.NotificationCreatedEvent) (ports.OutboxEvent, error) {
	rawPayload := domainEvent.Payload()
	if len(rawPayload) == 0 {
		rawPayload = []byte("{}")
	}

	payload := NotificationCreatedPayload{
		EventID:        domainEvent.EventID(),
		NotificationID: domainEvent.NotificationID(),
		UserID:         domainEvent.UserID(),
		Category:       domainEvent.Category(),
		Type:           domainEvent.Type(),
		Title:          domainEvent.Title(),
		Body:           domainEvent.Body(),
		Payload:        rawPayload,
		CreatedAt:      domainEvent.CreatedAt().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ports.OutboxEvent{}, err
	}

	return ports.OutboxEvent{
		EventID:       domainEvent.EventID(),
		EventType:     event.NotificationCreatedEventType,
		SchemaVersion: schemaVersion,
		Subject:       BuildSubject(domainEvent.UserID()),
		Payload:       body,
		OccurredAt:    domainEvent.Timestamp(),
	}, nil
}

// EmailRequestedPayload is the JSON body published to notification.evt.email.requested.
// The email dispatch consumer decodes it and hands the recipient and content to
// the EmailPort implementation.
type EmailRequestedPayload struct {
	EventID string `json:"event_id"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Title   string `json:"title"`
	Body    string `json:"body"`
}

// BuildEmailRequestedOutboxEvent renders an idempotent email-request outbox row.
// The caller is responsible for supplying a deterministic event id
// (source_event_id:user_id:email) so redeliveries collapse onto the same row.
func BuildEmailRequestedOutboxEvent(
	eventID string,
	to string,
	subject string,
	title string,
	body string,
	occurredAt time.Time,
) (ports.OutboxEvent, error) {
	payload := EmailRequestedPayload{
		EventID: eventID,
		To:      to,
		Subject: subject,
		Title:   title,
		Body:    body,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return ports.OutboxEvent{}, err
	}

	return ports.OutboxEvent{
		EventID:       eventID,
		EventType:     EmailRequestedEventType,
		SchemaVersion: schemaVersion,
		Subject:       SubjectEmailRequested,
		Payload:       raw,
		OccurredAt:    occurredAt,
	}, nil
}
