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
