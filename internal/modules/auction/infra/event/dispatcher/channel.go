package dispatcher

import "fmt"

// BuildAuctionEventChannel builds the Redis Pub/Sub channel name for auction events.
// Channel pattern: auction:{auctionID}:events
func BuildAuctionEventChannel(auctionID uint64) string {
	return fmt.Sprintf("auction:%d:events", auctionID)
}
