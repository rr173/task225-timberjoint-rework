package service

import (
	"errors"
	"testing"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
)

func TestBug06JointCannotReferenceEntitiesFromAnotherBatch(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	b1, err := app.Batches.Create(BatchCreateInput{Name: "batch-a", CoordinateSystem: "CGCS2000", LengthUnit: "mm", ClosureTolerance: 2, DirectionTolDeg: 10})
	if err != nil {
		t.Fatal(err)
	}
	b2, err := app.Batches.Create(BatchCreateInput{Name: "batch-b", CoordinateSystem: "CGCS2000", LengthUnit: "mm", ClosureTolerance: 2, DirectionTolDeg: 10})
	if err != nil {
		t.Fatal(err)
	}
	p, err := app.Points.Add(b2.ID, PointAddInput{No: "P-FOREIGN", X: 0})
	if err != nil {
		t.Fatal(err)
	}
	m, err := app.Members.Add(b2.ID, MemberAddInput{No: "M-FOREIGN", MemberType: "梁", StartX: 0, EndX: 1})
	if err != nil {
		t.Fatal(err)
	}
	_, err = app.Joints.Create(b1.ID, "N001", []string{p.ID}, []string{m.ID})
	if !errors.Is(err, model.ErrConflict) {
		t.Fatalf("cross-batch joint creation error = %v, want conflict", err)
	}
}
