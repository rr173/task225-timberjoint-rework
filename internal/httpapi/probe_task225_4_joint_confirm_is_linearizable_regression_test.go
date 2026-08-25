package httpapi

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"task225-timberjoint/internal/service"
	"task225-timberjoint/internal/store"
)

func TestBug04ConcurrentJointConfirmHasOneLinearizedWinner(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app, err := service.New(db)
	if err != nil {
		t.Fatal(err)
	}
	b, err := app.Batches.Create(service.BatchCreateInput{Name: "confirm-race", CoordinateSystem: "CGCS2000", LengthUnit: "mm", ClosureTolerance: 2, DirectionTolDeg: 10})
	if err != nil {
		t.Fatal(err)
	}
	p1, err := app.Points.Add(b.ID, service.PointAddInput{No: "P001", X: 0})
	if err != nil {
		t.Fatal(err)
	}
	p2, err := app.Points.Add(b.ID, service.PointAddInput{No: "P002", X: 0.5})
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
	h := New(app).Handler()
	statuses := make(chan int, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := httptest.NewRequest(http.MethodPost, "/api/joints/"+j.ID+"/confirm", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			statuses <- w.Code
		}()
	}
	wg.Wait()
	close(statuses)
	var ok, conflict int
	for code := range statuses {
		switch code {
		case http.StatusOK:
			ok++
		case http.StatusConflict:
			conflict++
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("concurrent confirm statuses = ok:%d conflict:%d, want one each", ok, conflict)
	}
	got, err := app.Joints.Get(j.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "confirmed" {
		t.Fatalf("joint status = %s, want confirmed", got.Status)
	}
}
