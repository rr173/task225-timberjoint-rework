package service

import (
	"errors"
	"sync"
	"testing"
	"time"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
)

func TestBug01ConcurrentEvidenceUploadIsIdempotent(t *testing.T) {
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
		Name: "concurrent-evidence", CoordinateSystem: "CGCS2000", LengthUnit: "mm",
		ClosureTolerance: 10, DirectionTolDeg: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	point, err := app.Points.Add(batch.ID, PointAddInput{No: "P001", X: 0})
	if err != nil {
		t.Fatal(err)
	}
	joint, err := app.Joints.Create(batch.ID, "N001", []string{point.ID}, nil)
	if err != nil {
		t.Fatal(err)
	}

	const callers = 20
	start := make(chan struct{})
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	duplicates := 0
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := app.Evidences.Add(joint.ID, "node.jpg", "east face", time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC))
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				successes++
			case errors.Is(err, model.ErrDuplicate):
				duplicates++
			default:
				t.Errorf("unexpected upload error: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	items, err := app.Evidences.ListByJoint(joint.ID)
	if err != nil {
		t.Fatal(err)
	}
	if successes != 1 || duplicates != callers-1 || len(items) != 1 {
		t.Fatalf("evidence idempotency broken: successes=%d duplicates=%d rows=%d", successes, duplicates, len(items))
	}
}
