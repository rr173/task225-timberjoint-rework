package evidence

import (
	"testing"
	"time"
)

func TestContentHashIsStableAndValidationRejectsFutureEvidence(t *testing.T) {
	takenAt := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	first := ContentHash("node.jpg", "east face", takenAt)
	second := ContentHash("node.jpg", "east face", takenAt)
	if first == "" || first != second {
		t.Fatalf("content hash is not stable: %q vs %q", first, second)
	}
	if err := ValidateEvidence("node.jpg", "east face", takenAt); err != nil {
		t.Fatalf("valid evidence rejected: %v", err)
	}
	if err := ValidateEvidence("node.jpg", "east face", time.Now().Add(2*time.Hour)); err == nil {
		t.Fatal("future evidence should be rejected")
	}
}
