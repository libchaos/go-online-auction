package messaging_test

import (
	"testing"

	"auction/internal/modules/auction/infra/messaging"
	"github.com/stretchr/testify/require"
)

func TestBuildBidCommandSubject(t *testing.T) {
	t.Run("builds dotted command subject with auction id", func(t *testing.T) {
		// Arrange
		auctionID := uint64(123)

		// Act
		subject := messaging.BuildBidCommandSubject(auctionID)

		// Assert
		require.Equal(t, "auction.cmd.bid.123", subject)
	})
}
