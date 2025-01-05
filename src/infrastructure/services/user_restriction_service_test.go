package services

import (
	"goproxy/domain/aggregates"
	"log"
	"testing"
)

func TestUserRestrictionService(t *testing.T) {
	cache := NewMockTTLCache[bool]()
	userRestrictionService, userRestrictionServiceErr := NewUserRestrictionService().
		UseCache(cache).
		Build()

	if userRestrictionServiceErr != nil {
		log.Fatal(userRestrictionServiceErr)
	}

	user, newUserErr := aggregates.NewUser(1, "testuser", make([]byte, 32), make([]byte, 32))
	if newUserErr != nil {
		log.Fatal(newUserErr)
	}

	t.Run("IsRestricted returns false for unrestricted user", func(t *testing.T) {
		restricted := userRestrictionService.IsRestricted(user)
		if restricted {
			t.Errorf("expected user to be unrestricted")
		}
	})

	t.Run("AddToRestrictionList adds user to restrictions", func(t *testing.T) {
		err := userRestrictionService.AddToRestrictionList(user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		restricted := userRestrictionService.IsRestricted(user)
		if !restricted {
			t.Errorf("expected user to be restricted")
		}
	})

	t.Run("RemoveFromRestrictionList removes user from restrictions", func(t *testing.T) {
		err := userRestrictionService.RemoveFromRestrictionList(user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		restricted := userRestrictionService.IsRestricted(user)
		if restricted {
			t.Errorf("expected user to be unrestricted")
		}
	})
}
