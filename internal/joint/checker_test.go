package joint

import (
	"testing"

	"task225-timberjoint/internal/geometry"
)

func TestCheckNodeCombinesClosureDirectionAndNumbering(t *testing.T) {
	result := CheckNode([]PointInfo{
		{No: "P001", Pos: geometry.Vec3{X: 0, Y: 0, Z: 0}},
		{No: "P002", Pos: geometry.Vec3{X: 1, Y: 0, Z: 0}},
	}, []MemberInfo{
		{No: "M001", Dir: geometry.Vec3{X: 1}},
		{No: "M002", Dir: geometry.Vec3{X: 2}},
	}, 1, 10)

	if !result.Closed || result.DirectionConflict || !result.NumberingOK {
		t.Fatalf("unexpected healthy report: %+v", result)
	}

	broken := CheckNode([]PointInfo{
		{No: "P001", Pos: geometry.Vec3{X: 0, Y: 0, Z: 0}},
		{No: "P001", Pos: geometry.Vec3{X: 1, Y: 0, Z: 0}},
	}, nil, 1, 10)
	if broken.NumberingOK || len(broken.Duplicates) != 1 {
		t.Fatalf("expected duplicate numbering to fail: %+v", broken)
	}
}
