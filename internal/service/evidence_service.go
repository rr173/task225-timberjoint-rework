package service

import (
	"fmt"
	"time"

	"task225-timberjoint/internal/evidence"
	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
	"task225-timberjoint/internal/survey"
)

// EvidenceService 编排现场照片证据的登记。
type EvidenceService struct {
	evidences *store.EvidenceStore
	joints    *store.JointStore
	batches   *store.BatchStore
}

// NewEvidenceService 构造证据服务。
func NewEvidenceService(evidences *store.EvidenceStore, joints *store.JointStore, batches *store.BatchStore) *EvidenceService {
	return &EvidenceService{evidences: evidences, joints: joints, batches: batches}
}

// Add 在指定节点关系下登记照片证据：校验批次可写、内容合法、幂等去重。
func (s *EvidenceService) Add(jointID, filename, caption string, takenAt time.Time) (*model.PhotoEvidence, error) {
	jrel, err := s.joints.Get(jointID)
	if err != nil {
		return nil, err
	}
	batch, err := s.batches.Get(jrel.BatchID)
	if err != nil {
		return nil, err
	}
	if !survey.IsMutable(batch.Status) {
		return nil, fmt.Errorf("%w: batch is not writable", model.ErrSealed)
	}
	if err := evidence.ValidateEvidence(filename, caption, takenAt); err != nil {
		return nil, err
	}

	hash := evidence.ContentHash(filename, caption, takenAt)
	if existing, err := s.evidences.GetByJointHash(jointID, hash); err == nil && existing != nil {
		return nil, model.ErrDuplicate
	}

	now := time.Now().UTC()
	e := &model.PhotoEvidence{
		ID:        newServiceID("evi"),
		JointID:   jointID,
		BatchID:   jrel.BatchID,
		Filename:  filename,
		Caption:   caption,
		Hash:      hash,
		TakenAt:   takenAt,
		CreatedAt: now,
	}
	if err := s.evidences.Create(e); err != nil {
		return nil, err
	}
	return e, nil
}

// ListByJoint 列出某节点关系的全部证据。
func (s *EvidenceService) ListByJoint(jointID string) ([]model.PhotoEvidence, error) {
	return s.evidences.ListByJoint(jointID)
}
