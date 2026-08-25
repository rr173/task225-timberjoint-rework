package service

import (
	"testing"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
)

func TestBug02ContactErrorPointExcludedFromClosure(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer db.Close()
	app, err := New(db)
	if err != nil { t.Fatal(err) }
	batch, err := app.Batches.Create(BatchCreateInput{Name: "contact-exclusion", CoordinateSystem: "CGCS2000", LengthUnit: "mm", ClosureTolerance: 2, DirectionTolDeg: 10})
	if err != nil { t.Fatal(err) }
	pointIDs := make([]string, 0, 3)
	for _, input := range []PointAddInput{
		{No: "P001", X: 0, Y: 0, Z: 0},
		{No: "P002", X: 1, Y: 0, Z: 0},
		{No: "P003", X: 100, Y: 100, Z: 100},
	} {
		point, err := app.Points.Add(batch.ID, input)
		if err != nil { t.Fatal(err) }
		pointIDs = append(pointIDs, point.ID)
	}
	if _, err := app.Points.MarkStatus(pointIDs[2], model.PointContactError); err != nil { t.Fatal(err) }
	joint, err := app.Joints.Create(batch.ID, "N001", pointIDs, nil)
	if err != nil { t.Fatal(err) }
	report, err := app.Joints.Check(joint.ID)
	if err != nil { t.Fatal(err) }
	if !report.Closed || report.ContactErrors != 1 {
		t.Fatalf("contact-error point contaminated closure: closed=%v contact_errors=%d residual=%.2f", report.Closed, report.ContactErrors, report.ClosureResidual)
	}
}
