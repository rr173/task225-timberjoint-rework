package member

import (
	"math"
	"testing"
)

func TestValidateMemberAndTenonDescription(t *testing.T) {
	axis, err := ValidateMember("M001", "梁", 0, 0, 0, 3, 4, 0)
	if err != nil {
		t.Fatalf("valid member rejected: %v", err)
	}
	if got := MemberDirection(0, 0, 0, 3, 4, 0); math.Abs(got.X-0.6) > 1e-12 || math.Abs(got.Y-0.8) > 1e-12 {
		t.Fatalf("unexpected unit direction: %+v", got)
	}
	if axis.Label != "M001" {
		t.Fatalf("unexpected axis label: %s", axis.Label)
	}
	if _, err := ValidateMember("M002", "梁", 0, 0, 0, 0, 0, 0); err == nil {
		t.Fatal("coincident endpoints should be rejected")
	}
	spec, err := ParseTenonDesc(`{"tenon_type":"直榫","mortise":"直卯","width":20,"depth":30}`)
	if err != nil || spec.TenonType != "直榫" || ValidateTenonSpec(spec) != nil {
		t.Fatalf("valid tenon description rejected: %+v, %v", spec, err)
	}
}
