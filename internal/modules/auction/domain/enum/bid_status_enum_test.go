package enum

import (
	"testing"
)

func TestNewBidStatusEnum(t *testing.T) {
	t.Run("valid accepted status", func(t *testing.T) {
		status, err := NewBidStatusEnum(EnumBidStatusAccepted)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if status.String() != EnumBidStatusAccepted {
			t.Errorf("expected status %s, got %s", EnumBidStatusAccepted, status.String())
		}
	})

	t.Run("valid rejected status", func(t *testing.T) {
		status, err := NewBidStatusEnum(EnumBidStatusRejected)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if status.String() != EnumBidStatusRejected {
			t.Errorf("expected status %s, got %s", EnumBidStatusRejected, status.String())
		}
	})

	t.Run("valid superseded status", func(t *testing.T) {
		status, err := NewBidStatusEnum(EnumBidStatusSuperseded)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if status.String() != EnumBidStatusSuperseded {
			t.Errorf("expected status %s, got %s", EnumBidStatusSuperseded, status.String())
		}
	})

	t.Run("rejects invalid status", func(t *testing.T) {
		_, err := NewBidStatusEnum("invalid")
		if err == nil {
			t.Fatal("expected error for invalid status, got nil")
		}
		expectedMsg := "invalid bid status: invalid"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects empty status", func(t *testing.T) {
		_, err := NewBidStatusEnum("")
		if err == nil {
			t.Fatal("expected error for empty status, got nil")
		}
		expectedMsg := "invalid bid status: "
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects status with wrong case", func(t *testing.T) {
		_, err := NewBidStatusEnum("ACCEPTED")
		if err == nil {
			t.Fatal("expected error for uppercase status, got nil")
		}
		expectedMsg := "invalid bid status: ACCEPTED"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects status with extra characters", func(t *testing.T) {
		_, err := NewBidStatusEnum("accepted ")
		if err == nil {
			t.Fatal("expected error for status with trailing space, got nil")
		}
		expectedMsg := "invalid bid status: accepted "
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects misspelled status", func(t *testing.T) {
		_, err := NewBidStatusEnum("accept")
		if err == nil {
			t.Fatal("expected error for misspelled status, got nil")
		}
		expectedMsg := "invalid bid status: accept"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects random string", func(t *testing.T) {
		_, err := NewBidStatusEnum("random_status")
		if err == nil {
			t.Fatal("expected error for random string, got nil")
		}
		expectedMsg := "invalid bid status: random_status"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects pending status", func(t *testing.T) {
		_, err := NewBidStatusEnum("pending")
		if err == nil {
			t.Fatal("expected error for non-existent 'pending' status, got nil")
		}
		expectedMsg := "invalid bid status: pending"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestBidStatusEnum_String(t *testing.T) {
	t.Run("returns correct string for accepted", func(t *testing.T) {
		status, _ := NewBidStatusEnum(EnumBidStatusAccepted)
		if status.String() != "accepted" {
			t.Errorf("expected 'accepted', got '%s'", status.String())
		}
	})

	t.Run("returns correct string for rejected", func(t *testing.T) {
		status, _ := NewBidStatusEnum(EnumBidStatusRejected)
		if status.String() != "rejected" {
			t.Errorf("expected 'rejected', got '%s'", status.String())
		}
	})

	t.Run("returns correct string for superseded", func(t *testing.T) {
		status, _ := NewBidStatusEnum(EnumBidStatusSuperseded)
		if status.String() != "superseded" {
			t.Errorf("expected 'superseded', got '%s'", status.String())
		}
	})
}

func TestBidStatusEnum_AllStatuses(t *testing.T) {
	t.Run("all three statuses are valid", func(t *testing.T) {
		statuses := []string{
			EnumBidStatusAccepted,
			EnumBidStatusRejected,
			EnumBidStatusSuperseded,
		}

		for _, statusValue := range statuses {
			_, err := NewBidStatusEnum(statusValue)
			if err != nil {
				t.Errorf("expected status %s to be valid, got error: %v", statusValue, err)
			}
		}
	})
}
