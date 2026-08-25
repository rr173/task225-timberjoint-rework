package geometry

import "testing"

func TestAnalyzeDirectionFindsLargestSpread(t *testing.T) {
	result := AnalyzeDirection([]MemberAxis{
		{Label: "M001", Dir: Vec3{X: 1}},
		{Label: "M002", Dir: Vec3{Y: 1}},
		{Label: "M003", Dir: Vec3{X: 2}},
	}, 45)

	if !result.Conflict {
		t.Fatal("expected perpendicular members to conflict")
	}
	if result.Spread < 89.9 || result.Spread > 90.1 {
		t.Fatalf("expected 90 degree spread, got %.3f", result.Spread)
	}
	if result.MaxPair != [2]string{"M001", "M002"} {
		t.Fatalf("unexpected max pair: %v", result.MaxPair)
	}
}
