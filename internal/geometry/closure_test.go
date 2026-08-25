package geometry

import "testing"

func TestAnalyzeClosureExcludesContactErrors(t *testing.T) {
	result := AnalyzeClosure([]PointSample{
		{Label: "P001", Pos: Vec3{X: 0, Y: 0, Z: 0}},
		{Label: "P002", Pos: Vec3{X: 1, Y: 0, Z: 0}},
		{Label: "P999", Pos: Vec3{X: 100, Y: 100, Z: 100}, Exclude: true},
	}, 1)

	if !result.Closed {
		t.Fatalf("expected valid points to close, got residual %.3f", result.Residual)
	}
	if result.PointCount != 2 || result.ContactErrors != 1 {
		t.Fatalf("unexpected participation counts: points=%d contact_errors=%d", result.PointCount, result.ContactErrors)
	}
}
