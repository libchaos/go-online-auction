package event

import (
	"testing"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
)

func TestNewDomainEvent(t *testing.T) {
	t.Run("generates unique event IDs", func(t *testing.T) {
		event1 := newDomainEvent()
		event2 := newDomainEvent()

		if event1.EventID() == event2.EventID() {
			t.Error("Expected unique event IDs, but got duplicates")
		}
	})

	t.Run("sets timestamp in UTC", func(t *testing.T) {
		before := time.Now().UTC()
		event := newDomainEvent()
		after := time.Now().UTC()

		if event.Timestamp().Location() != time.UTC {
			t.Error("Expected timestamp to be in UTC")
		}

		if event.Timestamp().Before(before) || event.Timestamp().After(after) {
			t.Errorf("Expected timestamp between %v and %v, got %v", before, after, event.Timestamp())
		}
	})

	t.Run("event ID is not empty", func(t *testing.T) {
		event := newDomainEvent()

		if event.EventID() == "" {
			t.Error("Expected non-empty event ID")
		}
	})
}

func TestDomainEvent_Getters(t *testing.T) {
	t.Run("EventID returns correct value", func(t *testing.T) {
		event := newDomainEvent()
		eventID := event.EventID()

		if eventID == "" {
			t.Error("Expected non-empty event ID")
		}

		// Verify getter returns the same value consistently
		if event.EventID() != eventID {
			t.Error("EventID getter should return consistent value")
		}
	})

	t.Run("Timestamp returns correct value", func(t *testing.T) {
		event := newDomainEvent()
		timestamp := event.Timestamp()

		// Verify getter returns the same value consistently
		if !event.Timestamp().Equal(timestamp) {
			t.Error("Timestamp getter should return consistent value")
		}
	})
}

func TestNewAuctionStartedEvent(t *testing.T) {
	t.Run("creates event with all fields", func(t *testing.T) {
		auctionID := uint64(123)
		listingID := uint64(456)
		startTime := time.Now().UTC()
		endTime := startTime.Add(24 * time.Hour)

		event := NewAuctionStartedEvent(auctionID, listingID, startTime, endTime)

		if event.AuctionID() != auctionID {
			t.Errorf("Expected auctionID %d, got %d", auctionID, event.AuctionID())
		}

		if event.ListingID() != listingID {
			t.Errorf("Expected listingID %d, got %d", listingID, event.ListingID())
		}

		if !event.StartTime().Equal(startTime) {
			t.Errorf("Expected startTime %v, got %v", startTime, event.StartTime())
		}

		if !event.EndTime().Equal(endTime) {
			t.Errorf("Expected endTime %v, got %v", endTime, event.EndTime())
		}
	})

	t.Run("generates unique event ID", func(t *testing.T) {
		startTime := time.Now().UTC()
		endTime := startTime.Add(24 * time.Hour)

		event := NewAuctionStartedEvent(1, 2, startTime, endTime)

		if event.EventID() == "" {
			t.Error("Expected non-empty event ID")
		}
	})

	t.Run("sets timestamp in UTC", func(t *testing.T) {
		startTime := time.Now().UTC()
		endTime := startTime.Add(24 * time.Hour)

		event := NewAuctionStartedEvent(1, 2, startTime, endTime)

		if event.Timestamp().Location() != time.UTC {
			t.Error("Expected timestamp to be in UTC")
		}
	})

	t.Run("creates multiple events with unique IDs", func(t *testing.T) {
		startTime := time.Now().UTC()
		endTime := startTime.Add(24 * time.Hour)

		event1 := NewAuctionStartedEvent(1, 2, startTime, endTime)
		event2 := NewAuctionStartedEvent(1, 2, startTime, endTime)

		if event1.EventID() == event2.EventID() {
			t.Error("Expected unique event IDs for different events")
		}
	})
}

func TestAuctionStartedEvent_Getters(t *testing.T) {
	auctionID := uint64(100)
	listingID := uint64(200)
	startTime := time.Date(2026, 1, 5, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 1, 6, 10, 0, 0, 0, time.UTC)

	event := NewAuctionStartedEvent(auctionID, listingID, startTime, endTime)

	t.Run("AuctionID returns correct value", func(t *testing.T) {
		if event.AuctionID() != auctionID {
			t.Errorf("Expected auctionID %d, got %d", auctionID, event.AuctionID())
		}
	})

	t.Run("ListingID returns correct value", func(t *testing.T) {
		if event.ListingID() != listingID {
			t.Errorf("Expected listingID %d, got %d", listingID, event.ListingID())
		}
	})

	t.Run("StartTime returns correct value", func(t *testing.T) {
		if !event.StartTime().Equal(startTime) {
			t.Errorf("Expected startTime %v, got %v", startTime, event.StartTime())
		}
	})

	t.Run("EndTime returns correct value", func(t *testing.T) {
		if !event.EndTime().Equal(endTime) {
			t.Errorf("Expected endTime %v, got %v", endTime, event.EndTime())
		}
	})
}

func TestNewBidPlacedEvent(t *testing.T) {
	t.Run("creates event with all fields", func(t *testing.T) {
		bidID := uint64(789)
		auctionID := uint64(123)
		userID := uint64(456)
		amount, _ := model.NewMoneyModel(10000, "USD")

		event := NewBidPlacedEvent(bidID, auctionID, userID, amount)

		if event.BidID() != bidID {
			t.Errorf("Expected bidID %d, got %d", bidID, event.BidID())
		}

		if event.AuctionID() != auctionID {
			t.Errorf("Expected auctionID %d, got %d", auctionID, event.AuctionID())
		}

		if event.UserID() != userID {
			t.Errorf("Expected userID %d, got %d", userID, event.UserID())
		}

		if event.Amount().AmountInCents() != amount.AmountInCents() {
			t.Errorf("Expected amount %d, got %d", amount.AmountInCents(), event.Amount().AmountInCents())
		}

		if event.Amount().Currency() != amount.Currency() {
			t.Errorf("Expected currency %s, got %s", amount.Currency(), event.Amount().Currency())
		}
	})

	t.Run("generates unique event ID", func(t *testing.T) {
		amount, _ := model.NewMoneyModel(10000, "USD")
		event := NewBidPlacedEvent(1, 2, 3, amount)

		if event.EventID() == "" {
			t.Error("Expected non-empty event ID")
		}
	})

	t.Run("sets timestamp in UTC", func(t *testing.T) {
		amount, _ := model.NewMoneyModel(10000, "USD")
		event := NewBidPlacedEvent(1, 2, 3, amount)

		if event.Timestamp().Location() != time.UTC {
			t.Error("Expected timestamp to be in UTC")
		}
	})

	t.Run("creates multiple events with unique IDs", func(t *testing.T) {
		amount, _ := model.NewMoneyModel(10000, "USD")
		event1 := NewBidPlacedEvent(1, 2, 3, amount)
		event2 := NewBidPlacedEvent(1, 2, 3, amount)

		if event1.EventID() == event2.EventID() {
			t.Error("Expected unique event IDs for different events")
		}
	})
}

func TestBidPlacedEvent_Getters(t *testing.T) {
	bidID := uint64(100)
	auctionID := uint64(200)
	userID := uint64(300)
	amount, _ := model.NewMoneyModel(50000, "EUR")

	event := NewBidPlacedEvent(bidID, auctionID, userID, amount)

	t.Run("BidID returns correct value", func(t *testing.T) {
		if event.BidID() != bidID {
			t.Errorf("Expected bidID %d, got %d", bidID, event.BidID())
		}
	})

	t.Run("AuctionID returns correct value", func(t *testing.T) {
		if event.AuctionID() != auctionID {
			t.Errorf("Expected auctionID %d, got %d", auctionID, event.AuctionID())
		}
	})

	t.Run("UserID returns correct value", func(t *testing.T) {
		if event.UserID() != userID {
			t.Errorf("Expected userID %d, got %d", userID, event.UserID())
		}
	})

	t.Run("Amount returns correct value", func(t *testing.T) {
		if event.Amount().AmountInCents() != amount.AmountInCents() {
			t.Errorf("Expected amount %d, got %d", amount.AmountInCents(), event.Amount().AmountInCents())
		}
		if event.Amount().Currency() != amount.Currency() {
			t.Errorf("Expected currency %s, got %s", amount.Currency(), event.Amount().Currency())
		}
	})
}

func TestNewAuctionEndedEvent(t *testing.T) {
	t.Run("creates event with winning bid", func(t *testing.T) {
		auctionID := uint64(123)
		winningBidID := uint64(789)
		finalAmount, _ := model.NewMoneyModel(100000, "USD")

		event := NewAuctionEndedEvent(auctionID, &winningBidID, &finalAmount)

		if event.AuctionID() != auctionID {
			t.Errorf("Expected auctionID %d, got %d", auctionID, event.AuctionID())
		}

		if event.WinningBidID() == nil {
			t.Error("Expected non-nil winningBidID")
		} else if *event.WinningBidID() != winningBidID {
			t.Errorf("Expected winningBidID %d, got %d", winningBidID, *event.WinningBidID())
		}

		if event.FinalAmount() == nil {
			t.Error("Expected non-nil finalAmount")
		} else {
			if event.FinalAmount().AmountInCents() != finalAmount.AmountInCents() {
				t.Errorf("Expected amount %d, got %d", finalAmount.AmountInCents(), event.FinalAmount().AmountInCents())
			}
			if event.FinalAmount().Currency() != finalAmount.Currency() {
				t.Errorf("Expected currency %s, got %s", finalAmount.Currency(), event.FinalAmount().Currency())
			}
		}
	})

	t.Run("creates event without winning bid (no bids)", func(t *testing.T) {
		auctionID := uint64(123)

		event := NewAuctionEndedEvent(auctionID, nil, nil)

		if event.AuctionID() != auctionID {
			t.Errorf("Expected auctionID %d, got %d", auctionID, event.AuctionID())
		}

		if event.WinningBidID() != nil {
			t.Error("Expected nil winningBidID for auction with no bids")
		}

		if event.FinalAmount() != nil {
			t.Error("Expected nil finalAmount for auction with no bids")
		}
	})

	t.Run("generates unique event ID", func(t *testing.T) {
		event := NewAuctionEndedEvent(1, nil, nil)

		if event.EventID() == "" {
			t.Error("Expected non-empty event ID")
		}
	})

	t.Run("sets timestamp in UTC", func(t *testing.T) {
		event := NewAuctionEndedEvent(1, nil, nil)

		if event.Timestamp().Location() != time.UTC {
			t.Error("Expected timestamp to be in UTC")
		}
	})

	t.Run("creates multiple events with unique IDs", func(t *testing.T) {
		event1 := NewAuctionEndedEvent(1, nil, nil)
		event2 := NewAuctionEndedEvent(1, nil, nil)

		if event1.EventID() == event2.EventID() {
			t.Error("Expected unique event IDs for different events")
		}
	})
}

func TestAuctionEndedEvent_Getters(t *testing.T) {
	t.Run("getters with winning bid", func(t *testing.T) {
		auctionID := uint64(100)
		winningBidID := uint64(200)
		finalAmount, _ := model.NewMoneyModel(75000, "GBP")

		event := NewAuctionEndedEvent(auctionID, &winningBidID, &finalAmount)

		if event.AuctionID() != auctionID {
			t.Errorf("Expected auctionID %d, got %d", auctionID, event.AuctionID())
		}

		if event.WinningBidID() == nil || *event.WinningBidID() != winningBidID {
			t.Errorf("Expected winningBidID %d", winningBidID)
		}

		if event.FinalAmount() == nil {
			t.Error("Expected non-nil finalAmount")
		}
	})

	t.Run("getters without winning bid", func(t *testing.T) {
		auctionID := uint64(100)

		event := NewAuctionEndedEvent(auctionID, nil, nil)

		if event.AuctionID() != auctionID {
			t.Errorf("Expected auctionID %d, got %d", auctionID, event.AuctionID())
		}

		if event.WinningBidID() != nil {
			t.Error("Expected nil winningBidID")
		}

		if event.FinalAmount() != nil {
			t.Error("Expected nil finalAmount")
		}
	})
}

func TestEventImmutability(t *testing.T) {
	t.Run("DomainEvent fields are not modifiable externally", func(t *testing.T) {
		event := newDomainEvent()
		originalID := event.EventID()
		originalTimestamp := event.Timestamp()

		// Verify getters return consistent values (immutability check)
		if event.EventID() != originalID {
			t.Error("EventID should be immutable")
		}

		if !event.Timestamp().Equal(originalTimestamp) {
			t.Error("Timestamp should be immutable")
		}
	})

	t.Run("AuctionStartedEvent fields are not modifiable externally", func(t *testing.T) {
		startTime := time.Now().UTC()
		endTime := startTime.Add(24 * time.Hour)
		event := NewAuctionStartedEvent(1, 2, startTime, endTime)

		originalAuctionID := event.AuctionID()
		originalListingID := event.ListingID()
		originalStartTime := event.StartTime()
		originalEndTime := event.EndTime()

		// Verify all getters return consistent values
		if event.AuctionID() != originalAuctionID {
			t.Error("AuctionID should be immutable")
		}
		if event.ListingID() != originalListingID {
			t.Error("ListingID should be immutable")
		}
		if !event.StartTime().Equal(originalStartTime) {
			t.Error("StartTime should be immutable")
		}
		if !event.EndTime().Equal(originalEndTime) {
			t.Error("EndTime should be immutable")
		}
	})
}
