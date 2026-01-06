package enum

import (
	"testing"
)

func TestNewAuctionStateEnum(t *testing.T) {
	t.Run("valid draft state", func(t *testing.T) {
		state, err := NewAuctionStateEnum(EnumAuctionStateDraft)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if state.String() != EnumAuctionStateDraft {
			t.Errorf("expected state %s, got %s", EnumAuctionStateDraft, state.String())
		}
	})

	t.Run("valid active state", func(t *testing.T) {
		state, err := NewAuctionStateEnum(EnumAuctionStateActive)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if state.String() != EnumAuctionStateActive {
			t.Errorf("expected state %s, got %s", EnumAuctionStateActive, state.String())
		}
	})

	t.Run("valid closed state", func(t *testing.T) {
		state, err := NewAuctionStateEnum(EnumAuctionStateClosed)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if state.String() != EnumAuctionStateClosed {
			t.Errorf("expected state %s, got %s", EnumAuctionStateClosed, state.String())
		}
	})

	t.Run("valid cancelled state", func(t *testing.T) {
		state, err := NewAuctionStateEnum(EnumAuctionStateCancelled)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if state.String() != EnumAuctionStateCancelled {
			t.Errorf("expected state %s, got %s", EnumAuctionStateCancelled, state.String())
		}
	})

	t.Run("rejects invalid state", func(t *testing.T) {
		_, err := NewAuctionStateEnum("invalid")
		if err == nil {
			t.Fatal("expected error for invalid state, got nil")
		}
		expectedMsg := "invalid auction state: invalid"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects empty state", func(t *testing.T) {
		_, err := NewAuctionStateEnum("")
		if err == nil {
			t.Fatal("expected error for empty state, got nil")
		}
		expectedMsg := "invalid auction state: "
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects state with wrong case", func(t *testing.T) {
		_, err := NewAuctionStateEnum("DRAFT")
		if err == nil {
			t.Fatal("expected error for uppercase state, got nil")
		}
		expectedMsg := "invalid auction state: DRAFT"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects state with extra characters", func(t *testing.T) {
		_, err := NewAuctionStateEnum("draft ")
		if err == nil {
			t.Fatal("expected error for state with trailing space, got nil")
		}
		expectedMsg := "invalid auction state: draft "
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects misspelled state", func(t *testing.T) {
		_, err := NewAuctionStateEnum("activ")
		if err == nil {
			t.Fatal("expected error for misspelled state, got nil")
		}
		expectedMsg := "invalid auction state: activ"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects random string", func(t *testing.T) {
		_, err := NewAuctionStateEnum("random_state")
		if err == nil {
			t.Fatal("expected error for random string, got nil")
		}
		expectedMsg := "invalid auction state: random_state"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestAuctionStateEnum_String(t *testing.T) {
	t.Run("returns correct string for draft", func(t *testing.T) {
		state, _ := NewAuctionStateEnum(EnumAuctionStateDraft)
		if state.String() != "draft" {
			t.Errorf("expected 'draft', got '%s'", state.String())
		}
	})

	t.Run("returns correct string for active", func(t *testing.T) {
		state, _ := NewAuctionStateEnum(EnumAuctionStateActive)
		if state.String() != "active" {
			t.Errorf("expected 'active', got '%s'", state.String())
		}
	})

	t.Run("returns correct string for closed", func(t *testing.T) {
		state, _ := NewAuctionStateEnum(EnumAuctionStateClosed)
		if state.String() != "closed" {
			t.Errorf("expected 'closed', got '%s'", state.String())
		}
	})

	t.Run("returns correct string for cancelled", func(t *testing.T) {
		state, _ := NewAuctionStateEnum(EnumAuctionStateCancelled)
		if state.String() != "cancelled" {
			t.Errorf("expected 'cancelled', got '%s'", state.String())
		}
	})
}

func TestAuctionStateEnum_AllStates(t *testing.T) {
	t.Run("all four states are valid", func(t *testing.T) {
		states := []string{
			EnumAuctionStateDraft,
			EnumAuctionStateActive,
			EnumAuctionStateClosed,
			EnumAuctionStateCancelled,
		}

		for _, stateValue := range states {
			_, err := NewAuctionStateEnum(stateValue)
			if err != nil {
				t.Errorf("expected state %s to be valid, got error: %v", stateValue, err)
			}
		}
	})
}
