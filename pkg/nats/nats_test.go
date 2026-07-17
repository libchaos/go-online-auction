package nats_test

import (
	"testing"
	"time"

	"auction/pkg/nats"
	"github.com/stretchr/testify/require"
)

func TestNew_InvalidConfig_ReturnsError(t *testing.T) {
	t.Run("empty URL returns error", func(t *testing.T) {
		// Arrange
		cfg := nats.Config{URL: ""}

		// Act
		conn, js, err := nats.New(cfg)

		// Assert
		require.Error(t, err)
		require.Nil(t, conn)
		require.Nil(t, js)
	})

	t.Run("blank URL returns error", func(t *testing.T) {
		// Arrange
		cfg := nats.Config{URL: "   "}

		// Act
		conn, js, err := nats.New(cfg)

		// Assert
		require.Error(t, err)
		require.Nil(t, conn)
		require.Nil(t, js)
	})

	t.Run("negative dedupe window returns error", func(t *testing.T) {
		// Arrange
		cfg := nats.Config{URL: "nats://localhost:4222", DedupeWindow: -1 * time.Second}

		// Act
		conn, js, err := nats.New(cfg)

		// Assert
		require.Error(t, err)
		require.Nil(t, conn)
		require.Nil(t, js)
	})
}
