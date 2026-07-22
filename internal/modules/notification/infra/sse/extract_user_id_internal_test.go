package sse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractUserID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		subject  string
		expected uint64
	}{
		{name: "valid subject", subject: "notification.evt.42", expected: 42},
		{name: "valid subject with trailing token", subject: "notification.evt.7.extra", expected: 7},
		{name: "non numeric user id", subject: "notification.evt.abc", expected: 0},
		{name: "too few parts", subject: "notification.evt", expected: 0},
		{name: "empty subject", subject: "", expected: 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Act
			userID := extractUserID(testCase.subject)

			// Assert
			require.Equal(t, testCase.expected, userID)
		})
	}
}
