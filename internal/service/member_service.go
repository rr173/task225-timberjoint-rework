package service

import (
	"fmt"
	"time"

	"task225-timberjoint/internal/member"
	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
	"task225-timberjoint/internal/survey"
)

// MemberService 编排木构构件的增删改与状态标记。
type MemberService struct {
	members *store.MemberStore
	batches *store.BatchStore
}

// NewMemberService 构造构件服务。
func NewMemberService(members *store.MemberStore, batches *store.BatchStore) *MemberService {
	return &MemberService{members: members, batches: batches}
}

// MemberAddInput 是新增构件的入参。
type MemberAddInput struct {
	No         string
	MemberType string
	StartX     float64
	StartY     float64
	StartZ     float64
	EndX       float64
	EndY       float64
	EndZ       float64
	TenonDesc  string
}

// Add 在指定批次下新增构件：校验批次可写、端点合法、榫卯描述合法、编号去重。
func (s *MemberService) Add(batchID string, in MemberAddInput) (*model.Member, error) {
	batch, err := s.batches.Get(batchID)
	if err != nil {
		return nil, err
	}
	if !survey.IsMutable(batch.Status) {
		return nil, fmt.Errorf("%w: batch status %s is not writable", model.ErrSealed, batch.Status)
	}
	if _, err := member.ValidateMember(in.No, in.MemberType, in.StartX, in.StartY, in.StartZ, in.EndX, in.EndY, in.EndZ); err != nil {
		return nil, err
	}
	spec, err := member.ParseTenonDesc(in.TenonDesc)
	if err != nil {
		return nil, err
	}
	if err := member.ValidateTenonSpec(spec); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	m := &model.Member{
		ID:         newServiceID("mem"),
		BatchID:    batchID,
		MemberNo:   in.No,
		MemberType: in.MemberType,
		StartX:     in.StartX,
		StartY:     in.StartY,
		StartZ:     in.StartZ,
		EndX:       in.EndX,
		EndY:       in.EndY,
		EndZ:       in.EndZ,
		TenonDesc:  in.TenonDesc,
		Status:     model.MemberUnmatched,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.members.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Get 按 ID 查询构件。
func (s *MemberService) Get(id string) (*model.Member, error) {
	return s.members.Get(id)
}

// ListByBatch 列出某批次的全部构件。
func (s *MemberService) ListByBatch(batchID string) ([]model.Member, error) {
	return s.members.ListByBatch(batchID)
}

// MarkStatus 标记构件状态（matched/direction_conflict/missing/unmatched）。
func (s *MemberService) MarkStatus(id, status string) (*model.Member, error) {
	m, err := s.members.Get(id)
	if err != nil {
		return nil, err
	}
	batch, err := s.batches.Get(m.BatchID)
	if err != nil {
		return nil, err
	}
	if !survey.IsMutable(batch.Status) {
		return nil, fmt.Errorf("%w: batch is not writable", model.ErrSealed)
	}
	switch status {
	case model.MemberUnmatched, model.MemberMatched, model.MemberDirectionConflict, model.MemberMissing:
	default:
		return nil, fmt.Errorf("%w: invalid member status %s", model.ErrInvalidInput, status)
	}
	if err := s.members.UpdateStatus(id, status); err != nil {
		return nil, err
	}
	return s.members.Get(id)
}

// Update 修正构件端点/类型/榫卯描述（用于复核阶段修正方向冲突）。
func (s *MemberService) Update(id string, in MemberAddInput) (*model.Member, error) {
	m, err := s.members.Get(id)
	if err != nil {
		return nil, err
	}
	batch, err := s.batches.Get(m.BatchID)
	if err != nil {
		return nil, err
	}
	if !survey.IsMutable(batch.Status) {
		return nil, fmt.Errorf("%w: batch is not writable", model.ErrSealed)
	}
	if _, err := member.ValidateMember(in.No, in.MemberType, in.StartX, in.StartY, in.StartZ, in.EndX, in.EndY, in.EndZ); err != nil {
		return nil, err
	}
	spec, err := member.ParseTenonDesc(in.TenonDesc)
	if err != nil {
		return nil, err
	}
	if err := member.ValidateTenonSpec(spec); err != nil {
		return nil, err
	}

	m.MemberNo = in.No
	m.MemberType = in.MemberType
	m.StartX = in.StartX
	m.StartY = in.StartY
	m.StartZ = in.StartZ
	m.EndX = in.EndX
	m.EndY = in.EndY
	m.EndZ = in.EndZ
	m.TenonDesc = in.TenonDesc
	// 修正端点后构件不再直接判定为冲突：方向是否合规由下次节点复核决定，
	// 此处回退到已匹配，避免构件状态与复核结论脱节。
	if m.Status == model.MemberDirectionConflict {
		m.Status = model.MemberMatched
	}
	m.UpdatedAt = time.Now().UTC()
	if err := s.members.Update(m); err != nil {
		return nil, err
	}
	return s.members.Get(id)
}
