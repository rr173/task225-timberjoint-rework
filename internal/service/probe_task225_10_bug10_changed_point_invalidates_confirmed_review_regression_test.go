package service

import "testing"

import "task225-timberjoint/internal/store"

func TestBug10ChangedPointInvalidatesConfirmedReview(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := app.Batches.Create(BatchCreateInput{
		Name: "stale-confirmed-review", CoordinateSystem: "CGCS2000", LengthUnit: "mm",
		ClosureTolerance: 2, DirectionTolDeg: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	p1, err := app.Points.Add(batch.ID, PointAddInput{No: "P001", X: 0})
	if err != nil {
		t.Fatal(err)
	}
	p2, err := app.Points.Add(batch.ID, PointAddInput{No: "P002", X: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	j, err := app.Joints.Create(batch.ID, "N001", []string{p1.ID, p2.ID}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Joints.Check(j.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Joints.Confirm(j.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Points.MarkStatus(p2.ID, "contact_error"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Batches.Advance(batch.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Batches.Advance(batch.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Versions.Publish(batch.ID, "N001", "stale confirmed review"); err == nil {
		t.Fatal("changed point left a confirmed review publishable")
	}
}
