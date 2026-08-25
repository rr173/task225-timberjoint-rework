package service

import (
	"fmt"
	"time"

	"task225-timberjoint/internal/geometry"
	"task225-timberjoint/internal/joint"
	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
	"task225-timberjoint/internal/survey"
)

// JointService 编排节点关系的建立、几何复核与流转。
type JointService struct {
	joints  *store.JointStore
	points  *store.PointStore
	members *store.MemberStore
	batches *store.BatchStore
}

// NewJointService 构造节点关系服务。
func NewJointService(joints *store.JointStore, points *store.PointStore, members *store.MemberStore, batches *store.BatchStore) *JointService {
	return &JointService{joints: joints, points: points, members: members, batches: batches}
}

// Create 建立节点关系：把一组测点与构件端点关联到某节点。
func (s *JointService) Create(batchID, nodeNo string, pointIDs, memberIDs []string) (*model.JointRelation, error) {
	batch, err := s.batches.Get(batchID)
	if err != nil {
		return nil, err
	}
	if !survey.IsMutable(batch.Status) {
		return nil, fmt.Errorf("%w: batch status %s is not writable", model.ErrSealed, batch.Status)
	}
	if nodeNo == "" {
		return nil, fmt.Errorf("%w: node number required", model.ErrInvalidInput)
	}
	if len(pointIDs) == 0 && len(memberIDs) == 0 {
		return nil, fmt.Errorf("%w: joint must reference at least one point or member", model.ErrInvalidInput)
	}
	for _, pid := range pointIDs {
		if _, err := s.points.GetInBatch(pid, batchID); err != nil {
			return nil, fmt.Errorf("%w: point %s does not belong to batch %s", model.ErrConflict, pid, batchID)
		}
	}
	for _, mid := range memberIDs {
		if _, err := s.members.GetInBatch(mid, batchID); err != nil {
			return nil, fmt.Errorf("%w: member %s does not belong to batch %s", model.ErrConflict, mid, batchID)
		}
	}

	now := time.Now().UTC()
	j := &model.JointRelation{
		ID:        newServiceID("jnt"),
		BatchID:   batchID,
		NodeNo:    nodeNo,
		PointIDs:  joint.EncodeIDList(pointIDs),
		MemberIDs: joint.EncodeIDList(memberIDs),
		Status:    model.JointCandidate,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.joints.Create(j); err != nil {
		return nil, err
	}
	return j, nil
}

// Get 按 ID 查询节点关系。
func (s *JointService) Get(id string) (*model.JointRelation, error) {
	return s.joints.Get(id)
}

// ListByBatch 列出某批次的全部节点关系。
func (s *JointService) ListByBatch(batchID string) ([]model.JointRelation, error) {
	return s.joints.ListByBatch(batchID)
}

// ValidateCurrent rejects publishing a review result after any referenced
// point or member has changed since the review was persisted.
func (s *JointService) ValidateCurrent(j *model.JointRelation) error {
	pointIDs, err := joint.ParseIDList(j.PointIDs)
	if err != nil {
		return err
	}
	for _, pid := range pointIDs {
		p, err := s.points.GetInBatch(pid, j.BatchID)
		if err != nil {
			return fmt.Errorf("resolve point %s for freshness: %w", pid, err)
		}
		if p.UpdatedAt.After(j.UpdatedAt) {
			return fmt.Errorf("%w: point %s changed after joint review", model.ErrInvalidState, pid)
		}
	}
	memberIDs, err := joint.ParseIDList(j.MemberIDs)
	if err != nil {
		return err
	}
	for _, mid := range memberIDs {
		m, err := s.members.GetInBatch(mid, j.BatchID)
		if err != nil {
			return fmt.Errorf("resolve member %s for freshness: %w", mid, err)
		}
		if m.UpdatedAt.After(j.UpdatedAt) {
			return fmt.Errorf("%w: member %s changed after joint review", model.ErrInvalidState, mid)
		}
	}
	return nil
}

// Check 执行节点几何复核：组装测点/构件，做闭合 + 方向 + 编号校验，
// 把结论落库（closed/broken），并联动更新构件方向冲突状态。
func (s *JointService) Check(id string) (*joint.CheckReport, error) {
	j, err := s.joints.Get(id)
	if err != nil {
		return nil, err
	}
	if !joint.CanCheck(j.Status) {
		return nil, fmt.Errorf("%w: joint status %s cannot be rechecked", model.ErrInvalidState, j.Status)
	}
	previousStatus := j.Status
	batch, err := s.batches.Get(j.BatchID)
	if err != nil {
		return nil, err
	}
	if !survey.IsMutable(batch.Status) {
		return nil, fmt.Errorf("%w: batch is not writable", model.ErrSealed)
	}

	// 组装测点输入。
	pointIDs, err := joint.ParseIDList(j.PointIDs)
	if err != nil {
		return nil, err
	}
	points := make([]joint.PointInfo, 0, len(pointIDs))
	for _, pid := range pointIDs {
		p, err := s.points.GetInBatch(pid, j.BatchID)
		if err != nil {
			return nil, fmt.Errorf("resolve point %s in batch %s: %w", pid, j.BatchID, model.ErrConflict)
		}
		ell := geometry.NewErrorEllipsoid(p.ErrorSemiMajor, p.ErrorSemiMinor, p.ErrorVertical, p.OrientationAzim)
		points = append(points, joint.PointInfo{
			No:       p.PointNo,
			Pos:      geometry.Vec3{X: p.X, Y: p.Y, Z: p.Z},
			Radius:   ell.EffectiveRadius(),
			Excluded: survey.PointExcluded(p.Status),
		})
	}

	// 组装构件输入。
	memberIDs, err := joint.ParseIDList(j.MemberIDs)
	if err != nil {
		return nil, err
	}
	members := make([]joint.MemberInfo, 0, len(memberIDs))
	for _, mid := range memberIDs {
		m, err := s.members.GetInBatch(mid, j.BatchID)
		if err != nil {
			return nil, fmt.Errorf("resolve member %s in batch %s: %w", mid, j.BatchID, model.ErrConflict)
		}
		dir := geometry.Vec3{X: m.EndX - m.StartX, Y: m.EndY - m.StartY, Z: m.EndZ - m.StartZ}
		members = append(members, joint.MemberInfo{
			No:      m.MemberNo,
			Dir:     dir,
			Missing: m.Status == model.MemberMissing,
		})
	}

	rep := joint.CheckNode(points, members, batch.ClosureTolerance, batch.DirectionTolDeg)

	// 结论落库：闭合不满足、编号冲突或方向冲突均判为断裂（需修正后重检）。
	status := model.JointClosed
	if !rep.Closed || !rep.NumberingOK || rep.DirectionConflict {
		status = model.JointBroken
	}
	j.Status = status
	j.ClosureResidual = rep.ClosureResidual
	j.DirectionSpread = rep.DirectionSpread
	j.CheckReport = joint.MarshalReport(rep)
	j.UpdatedAt = time.Now().UTC()
	if err := s.joints.UpdateCheck(j, previousStatus); err != nil {
		return nil, err
	}

	// 联动：方向冲突构件标记。
	if rep.DirectionConflict {
		for _, mid := range memberIDs {
			m, err := s.members.Get(mid)
			if err == nil && m.Status == model.MemberUnmatched {
				_ = s.members.UpdateStatus(mid, model.MemberDirectionConflict)
			}
		}
	}
	return &rep, nil
}

// Confirm 确认节点关系（闭合结论经复核人员确认）。
func (s *JointService) Confirm(id string) (*model.JointRelation, error) {
	j, err := s.joints.Get(id)
	if err != nil {
		return nil, err
	}
	if !joint.CanTransition(j.Status, model.JointConfirmed) {
		return nil, fmt.Errorf("%w: cannot confirm joint in status %s", model.ErrInvalidState, j.Status)
	}
	if err := s.joints.UpdateStatus(id, model.JointConfirmed); err != nil {
		return nil, err
	}
	return s.joints.Get(id)
}

// Reject 否决节点关系。
func (s *JointService) Reject(id string) (*model.JointRelation, error) {
	j, err := s.joints.Get(id)
	if err != nil {
		return nil, err
	}
	if !joint.CanTransition(j.Status, model.JointRejected) {
		return nil, fmt.Errorf("%w: cannot reject joint in status %s", model.ErrInvalidState, j.Status)
	}
	if err := s.joints.UpdateStatus(id, model.JointRejected); err != nil {
		return nil, err
	}
	return s.joints.Get(id)
}
