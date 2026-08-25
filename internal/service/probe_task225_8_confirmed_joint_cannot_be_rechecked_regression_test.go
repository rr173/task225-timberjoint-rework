package service

import (
	"errors"
	"testing"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
)

func TestBug08ConfirmedJointCannotBeRechecked(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	b, err := app.Batches.Create(BatchCreateInput{Name: "terminal-recheck", CoordinateSystem: "CGCS2000", LengthUnit: "mm", ClosureTolerance: 2, DirectionTolDeg: 10})
	if err != nil {
		t.Fatal(err)
	}
	p1, err := app.Points.Add(b.ID, PointAddInput{No: "P001", X: 0})
	if err != nil {
		t.Fatal(err)
	}
	p2, err := app.Points.Add(b.ID, PointAddInput{No: "P002", X: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	j, err := app.Joints.Create(b.ID, "N001", []string{p1.ID, p2.ID}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Joints.Check(j.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Joints.Confirm(j.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Joints.Check(j.ID); !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("recheck of confirmed joint error = %v, want invalid state", err)
	}
	got, err := app.Joints.Get(j.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.JointConfirmed {
		t.Fatalf("joint status after rejected recheck = %s, want confirmed", got.Status)
	}
}
