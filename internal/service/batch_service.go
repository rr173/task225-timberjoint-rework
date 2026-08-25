package service

import (
	"fmt"
	"time"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
	"task225-timberjoint/internal/survey"
)

// BatchService 编排测绘批次的生命周期。
type BatchService struct {
	batches  *store.BatchStore
	joints   *store.JointStore
	versions *store.VersionStore
}

// NewBatchService 构造批次服务。
func NewBatchService(batches *store.BatchStore, joints *store.JointStore, versions *store.VersionStore) *BatchService {
	return &BatchService{batches: batches, joints: joints, versions: versions}
}

// BatchCreateInput 是创建批次的入参。
type BatchCreateInput struct {
	Name             string
	Description      string
	CoordinateSystem string
	LengthUnit       string
	ClosureTolerance float64
	DirectionTolDeg  float64
}

// Create 创建测绘批次：校验坐标系/单位/容差，幂等指纹防重复。
func (s *BatchService) Create(in BatchCreateInput) (*model.SurveyBatch, error) {
	if err := survey.ValidateCoordinateSystem(in.CoordinateSystem); err != nil {
		return nil, err
	}
	if err := survey.ValidateLengthUnit(in.LengthUnit); err != nil {
		return nil, err
	}
	if err := survey.ValidateTolerances(in.ClosureTolerance, in.DirectionTolDeg); err != nil {
		return nil, err
	}
	if in.Name == "" {
		return nil, fmt.Errorf("%w: batch name required", model.ErrInvalidInput)
	}

	fp := fingerprint("batch", in.Name, in.CoordinateSystem, in.LengthUnit)
	if existing, err := s.batches.GetByFingerprint(fp); err == nil && existing != nil {
		return nil, model.ErrDuplicate
	}

	now := time.Now().UTC()
	b := &model.SurveyBatch{
		ID:               newServiceID("bat"),
		Name:             in.Name,
		Description:      in.Description,
		CoordinateSystem: in.CoordinateSystem,
		LengthUnit:       in.LengthUnit,
		ClosureTolerance: in.ClosureTolerance,
		DirectionTolDeg:  in.DirectionTolDeg,
		Status:           model.BatchCollecting,
		Fingerprint:      fp,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.batches.Create(b); err != nil {
		return nil, err
	}
	return b, nil
}

// Get 按 ID 查询批次。
func (s *BatchService) Get(id string) (*model.SurveyBatch, error) {
	return s.batches.Get(id)
}

// List 列出全部批次。
func (s *BatchService) List() ([]model.SurveyBatch, error) {
	return s.batches.List()
}

// Advance 把批次流转到下一状态（采集中→整理中→待复核→已发布）。
// 封存须显式调用 Archive。
func (s *BatchService) Advance(id string) (*model.SurveyBatch, error) {
	b, err := s.batches.Get(id)
	if err != nil {
		return nil, err
	}
	if b.Status == model.BatchArchived {
		return nil, fmt.Errorf("%w: archived batch cannot advance", model.ErrSealed)
	}
	next := survey.NextBatchState(b.Status)
	if next == "" {
		return nil, fmt.Errorf("%w: batch already at terminal state %s", model.ErrInvalidState, b.Status)
	}
	if err := s.batches.UpdateStatus(id, next); err != nil {
		return nil, err
	}
	return s.batches.Get(id)
}

// Archive 封存批次（终态）。
func (s *BatchService) Archive(id string) (*model.SurveyBatch, error) {
	b, err := s.batches.Get(id)
	if err != nil {
		return nil, err
	}
	if b.Status == model.BatchArchived {
		return nil, fmt.Errorf("%w: batch already archived", model.ErrInvalidState)
	}
	if b.Status != model.BatchPublished {
		return nil, fmt.Errorf("%w: only published batch can be archived, got %s", model.ErrInvalidState, b.Status)
	}
	ready, err := s.versions.AllNodesFrozen(id)
	if err != nil {
		return nil, err
	}
	if !ready {
		return nil, fmt.Errorf("%w: all batch nodes need frozen versions before archival", model.ErrInvalidState)
	}
	if err := s.batches.UpdateStatus(id, model.BatchArchived); err != nil {
		return nil, err
	}
	return s.batches.Get(id)
}
