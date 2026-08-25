package service

import (
	"errors"
	"testing"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
)

// newMutBatchWithMemberAndPoint 构造一个处于可写状态的批次，
// 并在其中登记一个测点和一个构件，返回它们的 ID。
func newMutBatchWithMemberAndPoint(t *testing.T, app *App, name string) (batchID, pointID, memberID string) {
	t.Helper()
	batch, err := app.Batches.Create(BatchCreateInput{
		Name:             name,
		CoordinateSystem: "CGCS2000",
		LengthUnit:       "mm",
		ClosureTolerance: 10,
		DirectionTolDeg:  10,
	})
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	p, err := app.Points.Add(batch.ID, PointAddInput{
		No: "P001", X: 1, Y: 2, Z: 3,
		SemiMajor: 1, SemiMinor: 1, Vertical: 1,
	})
	if err != nil {
		t.Fatalf("add point: %v", err)
	}
	m, err := app.Members.Add(batch.ID, MemberAddInput{
		No: "M001", MemberType: "梁",
		StartX: 0, StartY: 0, StartZ: 0, EndX: 1000, EndY: 0, EndZ: 0,
	})
	if err != nil {
		t.Fatalf("add member: %v", err)
	}
	return batch.ID, p.ID, m.ID
}

// TestJointCreateRejectsCrossBatchReferences 确保：建立节点关系时，
// 不能把其他测绘批次的测点或构件挂到当前批次；跨批次引用应被拒绝，
// 不生成节点关系。
func TestJointCreateRejectsCrossBatchReferences(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	app, err := New(db)
	if err != nil {
		t.Fatalf("create app: %v", err)
	}

	// 两个独立批次，各自持有一个测点和一个构件。
	batchA, ptA, memA := newMutBatchWithMemberAndPoint(t, app, "batch-cross-ref-A")
	batchB, ptB, memB := newMutBatchWithMemberAndPoint(t, app, "batch-cross-ref-B")

	// 前置条件：A、B 批次的测点/构件 ID 互不相同。
	if ptA == ptB || memA == memB || batchA == batchB {
		t.Fatalf("test setup invariant: ids must differ across batches")
	}

	// 同批次引用应成功。
	jSame, err := app.Joints.Create(batchA, "N001", []string{ptA}, []string{memA})
	if err != nil {
		t.Fatalf("same-batch joint should be created, got %v", err)
	}
	if jSame.BatchID != batchA {
		t.Fatalf("joint should belong to batch A, got %s", jSame.BatchID)
	}

	// 跨批次引用测点：B 批次的测点挂到 A 批次的节点，应被拒绝。
	if _, err := app.Joints.Create(batchA, "N002", []string{ptB}, nil); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("cross-batch point reference must be rejected with ErrInvalidInput, got %v", err)
	}

	// 跨批次引用构件：B 批次的构件挂到 A 批次的节点，应被拒绝。
	if _, err := app.Joints.Create(batchA, "N003", nil, []string{memB}); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("cross-batch member reference must be rejected with ErrInvalidInput, got %v", err)
	}

	// 混合引用（一个本批次、一个跨批次）同样应被拒绝。
	if _, err := app.Joints.Create(batchA, "N004", []string{ptA, ptB}, []string{memA, memB}); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("mixed cross-batch reference must be rejected with ErrInvalidInput, got %v", err)
	}

	// 不存在的测点 ID 也应被拒绝（视作不属于本批次）。
	if _, err := app.Joints.Create(batchA, "N005", []string{"pt_nonexistent"}, nil); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("nonexistent point reference must be rejected with ErrInvalidInput, got %v", err)
	}

	// 被拒绝后，A 批次只能查到同批次创建的那一条节点关系。
	joints, err := app.Joints.ListByBatch(batchA)
	if err != nil {
		t.Fatalf("list joints: %v", err)
	}
	if len(joints) != 1 || joints[0].ID != jSame.ID {
		t.Fatalf("rejected cross-batch joints must not be persisted, got %d joints", len(joints))
	}
}
