package service

import (
	"errors"
	"testing"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
)

// TestArchiveRejectedWithoutFrozenVersions locks the rule: a published batch
// may not be archived while any node still lacks a frozen version. The
// request must be rejected and the batch must stay published.
func TestArchiveRejectedWithoutFrozenVersions(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	app, err := New(db)
	if err != nil {
		t.Fatalf("create app: %v", err)
	}

	batch, err := app.Batches.Create(BatchCreateInput{
		Name:             "archive-guard", CoordinateSystem: "CGCS2000", LengthUnit: "mm",
		ClosureTolerance: 50.0, DirectionTolDeg: 10.0,
	})
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	// 4 个闭合测点。
	p1, err := app.Points.Add(batch.ID, PointAddInput{No: "P001", X: 1000, Y: 1000, Z: 1000, SemiMajor: 5, SemiMinor: 5, Vertical: 5})
	if err != nil {
		t.Fatalf("add point 1: %v", err)
	}
	p2, err := app.Points.Add(batch.ID, PointAddInput{No: "P002", X: 1010, Y: 1000, Z: 1000, SemiMajor: 5, SemiMinor: 5, Vertical: 5})
	if err != nil {
		t.Fatalf("add point 2: %v", err)
	}
	p3, err := app.Points.Add(batch.ID, PointAddInput{No: "P003", X: 1000, Y: 1010, Z: 1000, SemiMajor: 5, SemiMinor: 5, Vertical: 5})
	if err != nil {
		t.Fatalf("add point 3: %v", err)
	}
	p4, err := app.Points.Add(batch.ID, PointAddInput{No: "P004", X: 1000, Y: 1000, Z: 1010, SemiMajor: 5, SemiMinor: 5, Vertical: 5})
	if err != nil {
		t.Fatalf("add point 4: %v", err)
	}
	m1, err := app.Members.Add(batch.ID, MemberAddInput{No: "M001", MemberType: "梁", StartX: 0, StartY: 0, StartZ: 0, EndX: 1000, EndY: 0, EndZ: 0})
	if err != nil {
		t.Fatalf("add member 1: %v", err)
	}
	m2, err := app.Members.Add(batch.ID, MemberAddInput{No: "M002", MemberType: "梁", StartX: 0, StartY: 500, StartZ: 0, EndX: 1000, EndY: 500, EndZ: 0})
	if err != nil {
		t.Fatalf("add member 2: %v", err)
	}

	// 流转到待复核。
	if _, err := app.Batches.Advance(batch.ID); err != nil {
		t.Fatalf("advance to organizing: %v", err)
	}
	if _, err := app.Batches.Advance(batch.ID); err != nil {
		t.Fatalf("advance to reviewing: %v", err)
	}

	jrel, err := app.Joints.Create(batch.ID, "N001", []string{p1.ID, p2.ID, p3.ID, p4.ID}, []string{m1.ID, m2.ID})
	if err != nil {
		t.Fatalf("create joint: %v", err)
	}
	rep, err := app.Joints.Check(jrel.ID)
	if err != nil {
		t.Fatalf("check joint: %v", err)
	}
	if !rep.Closed || rep.DirectionConflict {
		t.Fatalf("joint should close, got closed=%v conflict=%v", rep.Closed, rep.DirectionConflict)
	}
	if _, err := app.Joints.Confirm(jrel.ID); err != nil {
		t.Fatalf("confirm joint: %v", err)
	}

	// 发布批次（仅要求全部节点已确认，不要求冻结版本）。
	if _, err := app.Batches.Advance(batch.ID); err != nil {
		t.Fatalf("advance to published: %v", err)
	}

	// 尚未发布任何冻结版本：封存应被拒绝。
	if _, err := app.Batches.Archive(batch.ID); err == nil {
		t.Fatal("archive should be rejected while nodes have no frozen versions")
	} else if !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("archive rejection should be ErrInvalidState, got %v", err)
	}

	// 批次必须保持已发布状态。
	got, err := app.Batches.Get(batch.ID)
	if err != nil {
		t.Fatalf("get batch after rejected archive: %v", err)
	}
	if got.Status != model.BatchPublished {
		t.Fatalf("batch should stay published after rejected archive, got %s", got.Status)
	}

	// 冻结节点版本后，封存应成功。
	if _, err := app.Versions.Publish(batch.ID, "N001", "复核通过，冻结发布"); err != nil {
		t.Fatalf("publish version: %v", err)
	}
	archived, err := app.Batches.Archive(batch.ID)
	if err != nil {
		t.Fatalf("archive after freezing should succeed: %v", err)
	}
	if archived.Status != model.BatchArchived {
		t.Fatalf("batch should be archived, got %s", archived.Status)
	}
}
