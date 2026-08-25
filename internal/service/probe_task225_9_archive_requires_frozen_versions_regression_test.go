package service

import (
	"errors"
	"testing"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
)

func TestBug09BatchCannotBeArchivedBeforeNodeVersionsFreeze(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	b, err := app.Batches.Create(BatchCreateInput{Name: "archive-gate", CoordinateSystem: "CGCS2000", LengthUnit: "mm", ClosureTolerance: 2, DirectionTolDeg: 10})
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
	for i := 0; i < 3; i++ {
		if _, err := app.Batches.Advance(b.ID); err != nil {
			t.Fatal(err)
		}
	}
	_, err = app.Batches.Archive(b.ID)
	if !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("archive without frozen version error = %v, want invalid state", err)
	}
	got, err := app.Batches.Get(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.BatchPublished {
		t.Fatalf("batch status after rejected archive = %s, want published", got.Status)
	}
}
