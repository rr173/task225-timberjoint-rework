package service

import (
	"fmt"
	"time"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
	"task225-timberjoint/internal/survey"
)

// PointService 编排测点的增删改与状态标记。
type PointService struct {
	points  *store.PointStore
	batches *store.BatchStore
}

// NewPointService 构造测点服务。
func NewPointService(points *store.PointStore, batches *store.BatchStore) *PointService {
	return &PointService{points: points, batches: batches}
}

// PointAddInput 是新增测点的入参。
type PointAddInput struct {
	No             string
	X, Y, Z        float64
	SemiMajor      float64
	SemiMinor      float64
	Vertical       float64
	OrientationAzim float64
}

// Add 在指定批次下新增测点：校验批次可写、测点坐标与误差椭球合法、幂等去重。
func (s *PointService) Add(batchID string, in PointAddInput) (*model.SurveyPoint, error) {
	batch, err := s.batches.Get(batchID)
	if err != nil {
		return nil, err
	}
	if !survey.IsMutable(batch.Status) {
		return nil, fmt.Errorf("%w: batch status %s is not writable", model.ErrSealed, batch.Status)
	}
	if _, err := survey.ValidatePoint(in.No, in.X, in.Y, in.Z, in.SemiMajor, in.SemiMinor, in.Vertical, in.OrientationAzim); err != nil {
		return nil, err
	}

	fp := fingerprint("point", batchID, in.No, coordKey(in.X, in.Y, in.Z))
	now := time.Now().UTC()
	p := &model.SurveyPoint{
		ID:              newServiceID("pt"),
		BatchID:         batchID,
		PointNo:         in.No,
		X:               in.X,
		Y:               in.Y,
		Z:               in.Z,
		ErrorSemiMajor:  in.SemiMajor,
		ErrorSemiMinor:  in.SemiMinor,
		ErrorVertical:   in.Vertical,
		OrientationAzim: in.OrientationAzim,
		Status:          survey.DefaultPointStatus(),
		Fingerprint:     fp,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.points.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

// Get 按 ID 查询测点。
func (s *PointService) Get(id string) (*model.SurveyPoint, error) {
	return s.points.Get(id)
}

// ListByBatch 列出某批次的全部测点。
func (s *PointService) ListByBatch(batchID string) ([]model.SurveyPoint, error) {
	return s.points.ListByBatch(batchID)
}

// MarkStatus 标记测点状态（valid/contact_error/gap），封存批次拒绝。
func (s *PointService) MarkStatus(id, status string) (*model.SurveyPoint, error) {
	p, err := s.points.Get(id)
	if err != nil {
		return nil, err
	}
	batch, err := s.batches.Get(p.BatchID)
	if err != nil {
		return nil, err
	}
	if !survey.IsMutable(batch.Status) {
		return nil, fmt.Errorf("%w: batch is not writable", model.ErrSealed)
	}
	switch status {
	case model.PointValid, model.PointContactError, model.PointGap, model.PointPending:
	default:
		return nil, fmt.Errorf("%w: invalid point status %s", model.ErrInvalidInput, status)
	}
	if err := s.points.UpdateStatus(id, status); err != nil {
		return nil, err
	}
	return s.points.Get(id)
}
