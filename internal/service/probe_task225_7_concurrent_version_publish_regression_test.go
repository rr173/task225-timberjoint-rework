package service

import (
	"sync"
	"testing"

	"task225-timberjoint/internal/store"
)

func TestBug07ConcurrentVersionPublishesAllocateDistinctVersions(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	b, err := app.Batches.Create(BatchCreateInput{Name: "version-race", CoordinateSystem: "CGCS2000", LengthUnit: "mm", ClosureTolerance: 2, DirectionTolDeg: 10})
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
	for i := 0; i < 2; i++ {
		if _, err := app.Batches.Advance(b.ID); err != nil {
			t.Fatal(err)
		}
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, reason := range []string{"review-a", "review-b"} {
		wg.Add(1)
		go func(reason string) {
			defer wg.Done()
			_, err := app.Versions.Publish(b.ID, "N001", reason)
			errs <- err
		}(reason)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent publish failed: %v", err)
		}
	}
	versions, err := app.Versions.ListByNode(b.ID, "N001")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 || versions[0].VersionNo != 2 || versions[1].VersionNo != 1 {
		t.Fatalf("versions = %#v, want version numbers 2 and 1", versions)
	}
}
