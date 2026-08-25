package service

import (
	"testing"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
)

func TestBug03CorrectedMemberLeavesDirectionConflictState(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer db.Close()
	app, err := New(db)
	if err != nil { t.Fatal(err) }
	batch, err := app.Batches.Create(BatchCreateInput{Name: "member-recovery", CoordinateSystem: "CGCS2000", LengthUnit: "mm", ClosureTolerance: 10, DirectionTolDeg: 10})
	if err != nil { t.Fatal(err) }
	p1, err := app.Points.Add(batch.ID, PointAddInput{No: "P001", X: 0})
	if err != nil { t.Fatal(err) }
	p2, err := app.Points.Add(batch.ID, PointAddInput{No: "P002", X: 1})
	if err != nil { t.Fatal(err) }
	members := make([]*model.Member, 0, 3)
	for _, input := range []MemberAddInput{
		{No: "M001", MemberType: "梁", EndX: 10},
		{No: "M002", MemberType: "梁", EndX: 20},
		{No: "M003", MemberType: "梁", EndX: 10, EndY: 10},
	} {
		member, err := app.Members.Add(batch.ID, input)
		if err != nil { t.Fatal(err) }
		members = append(members, member)
	}
	joint, err := app.Joints.Create(batch.ID, "N001", []string{p1.ID, p2.ID}, []string{members[0].ID, members[1].ID, members[2].ID})
	if err != nil { t.Fatal(err) }
	first, err := app.Joints.Check(joint.ID)
	if err != nil || !first.DirectionConflict { t.Fatalf("expected initial direction conflict: %+v, %v", first, err) }
	if _, err := app.Members.Update(members[2].ID, MemberAddInput{No: "M003", MemberType: "梁", EndX: 10}); err != nil { t.Fatal(err) }
	second, err := app.Joints.Check(joint.ID)
	if err != nil || second.DirectionConflict || !second.Closed { t.Fatalf("corrected node did not recover: %+v, %v", second, err) }
	restored, err := app.Members.Get(members[2].ID)
	if err != nil { t.Fatal(err) }
	if restored.Status == model.MemberDirectionConflict { t.Fatalf("corrected member retained conflict state: %s", restored.Status) }
}
