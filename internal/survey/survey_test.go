package survey

import (
	"testing"

	"task225-timberjoint/internal/model"
)

func TestBatchLifecycleAndValidation(t *testing.T) {
	state := model.BatchCollecting
	for _, want := range []string{model.BatchOrganizing, model.BatchReviewing, model.BatchPublished, model.BatchArchived} {
		if !CanTransition(state, want) || NextBatchState(state) != want {
			t.Fatalf("expected %s -> %s", state, want)
		}
		state = want
	}
	if IsMutable(model.BatchArchived) || NextBatchState(model.BatchArchived) != "" {
		t.Fatal("archived batch must be terminal and immutable")
	}
	if err := ValidateCoordinateSystem(" "); err == nil {
		t.Fatal("blank coordinate system should be rejected")
	}
	if err := ValidateTolerances(10, 180); err != nil {
		t.Fatalf("valid tolerances rejected: %v", err)
	}
}
