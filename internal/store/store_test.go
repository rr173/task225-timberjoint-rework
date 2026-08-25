package store

import (
	"path/filepath"
	"testing"
	"time"

	"task225-timberjoint/internal/model"
)

func TestBatchPersistsAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "survey.db")
	created := time.Now().UTC().Truncate(time.Microsecond)
	batch := &model.SurveyBatch{
		ID: "bat-test", Name: "reopen", CoordinateSystem: "CGCS2000", LengthUnit: "mm",
		ClosureTolerance: 10, DirectionTolDeg: 10, Status: model.BatchArchived,
		Fingerprint: "fp-test", CreatedAt: created, UpdatedAt: created,
	}

	db, err := Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := NewBatchStore(db).Create(batch); err != nil {
		db.Close()
		t.Fatalf("create batch: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	db, err = Open(path)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	defer db.Close()
	restored, err := NewBatchStore(db).Get(batch.ID)
	if err != nil {
		t.Fatalf("restore batch: %v", err)
	}
	if restored.Status != model.BatchArchived || restored.Fingerprint != batch.Fingerprint {
		t.Fatalf("restored batch lost state: %+v", restored)
	}
}

// TestAllNodesFrozen guards the archive precondition: a published batch may
// only be archived once every node owns a frozen version.
func TestAllNodesFrozen(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Microsecond)
	batch := &model.SurveyBatch{
		ID: "bat-fz", Name: "freeze-guard", CoordinateSystem: "CGCS2000", LengthUnit: "mm",
		ClosureTolerance: 10, DirectionTolDeg: 10, Status: model.BatchPublished,
		Fingerprint: "fp-fz", CreatedAt: now, UpdatedAt: now,
	}
	if err := NewBatchStore(db).Create(batch); err != nil {
		t.Fatalf("create batch: %v", err)
	}
	joints := NewJointStore(db)
	versions := NewVersionStore(db)

	// 两个节点 N001、N002。
	j1 := &model.JointRelation{ID: "jnt-1", BatchID: batch.ID, NodeNo: "N001",
		PointIDs: "[]", MemberIDs: "[]", Status: model.JointConfirmed, CreatedAt: now, UpdatedAt: now}
	j2 := &model.JointRelation{ID: "jnt-2", BatchID: batch.ID, NodeNo: "N002",
		PointIDs: "[]", MemberIDs: "[]", Status: model.JointConfirmed, CreatedAt: now, UpdatedAt: now}
	if err := joints.Create(j1); err != nil {
		t.Fatalf("create joint 1: %v", err)
	}
	if err := joints.Create(j2); err != nil {
		t.Fatalf("create joint 2: %v", err)
	}

	// 尚无任何冻结版本：不可封存。
	if ok, err := versions.AllNodesFrozen(batch.ID); err != nil {
		t.Fatalf("all frozen (none): %v", err)
	} else if ok {
		t.Fatal("expected not frozen before any version published")
	}

	// 仅冻结 N001：仍不可封存（N002 未冻结）。
	v1 := &model.NodeVersion{ID: "ver-1", BatchID: batch.ID, NodeNo: "N001", VersionNo: 1,
		Status: model.VersionFrozen, Reason: "publish", CreatedAt: now, PublishedAt: &now}
	if err := versions.CreateNext(v1); err != nil {
		t.Fatalf("create frozen version 1: %v", err)
	}
	if ok, err := versions.AllNodesFrozen(batch.ID); err != nil {
		t.Fatalf("all frozen (one): %v", err)
	} else if ok {
		t.Fatal("expected not frozen with N002 missing a frozen version")
	}

	// 再冻结 N002：可封存。
	v2 := &model.NodeVersion{ID: "ver-2", BatchID: batch.ID, NodeNo: "N002", VersionNo: 1,
		Status: model.VersionFrozen, Reason: "publish", CreatedAt: now, PublishedAt: &now}
	if err := versions.CreateNext(v2); err != nil {
		t.Fatalf("create frozen version 2: %v", err)
	}
	if ok, err := versions.AllNodesFrozen(batch.ID); err != nil {
		t.Fatalf("all frozen (both): %v", err)
	} else if !ok {
		t.Fatal("expected frozen once every node has a frozen version")
	}
}
