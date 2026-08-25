package service

import (
	"encoding/json"
	"fmt"
	"time"

	"task225-timberjoint/internal/joint"
	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
	"task225-timberjoint/internal/version"
)

// VersionService 编排节点版本的发布与快照冻结。
type VersionService struct {
	versions  *store.VersionStore
	joints    *store.JointStore
	points    *store.PointStore
	members   *store.MemberStore
	evidences *store.EvidenceStore
	batches   *store.BatchStore
}

// NewVersionService 构造版本服务。
func NewVersionService(versions *store.VersionStore, joints *store.JointStore, points *store.PointStore, members *store.MemberStore, evidences *store.EvidenceStore, batches *store.BatchStore) *VersionService {
	return &VersionService{
		versions:  versions,
		joints:    joints,
		points:    points,
		members:   members,
		evidences: evidences,
		batches:   batches,
	}
}

// NodeSnapshot 是节点版本冻结的不可变快照。
type NodeSnapshot struct {
	NodeNo    string                `json:"node_no"`
	Joint     *model.JointRelation  `json:"joint"`
	Points    []model.SurveyPoint   `json:"points"`
	Members   []model.Member        `json:"members"`
	Evidences []model.PhotoEvidence `json:"evidences"`
}

// Publish 发布节点版本：组装测点/构件/关系/证据快照并冻结，
// 版本号递增；已存在的冻结版本被标记为替代。
func (s *VersionService) Publish(batchID, nodeNo, reason string) (*model.NodeVersion, error) {
	batch, err := s.batches.Get(batchID)
	if err != nil {
		return nil, err
	}
	if batch.Status != model.BatchReviewing && batch.Status != model.BatchPublished {
		return nil, fmt.Errorf("%w: version publish requires reviewing/published batch, got %s", model.ErrInvalidState, batch.Status)
	}
	if err := version.ValidateReason(reason); err != nil {
		return nil, err
	}

	// 定位该节点的关系（已闭合或已确认）。
	jrel, err := s.findJoint(batchID, nodeNo)
	if err != nil {
		return nil, err
	}
	if jrel.Status != model.JointClosed && jrel.Status != model.JointConfirmed {
		return nil, fmt.Errorf("%w: node %s not closed/confirmed (status %s)", model.ErrInvalidState, nodeNo, jrel.Status)
	}

	// 组装快照。
	snap, err := s.buildSnapshot(jrel)
	if err != nil {
		return nil, err
	}
	snapJSON, err := json.Marshal(snap)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}

	now := time.Now().UTC()
	v := &model.NodeVersion{
		ID:           newServiceID("ver"),
		BatchID:      batchID,
		NodeNo:       nodeNo,
		Status:       model.VersionFrozen,
		SnapshotJSON: string(snapJSON),
		Reason:       reason,
		CreatedAt:    now,
		PublishedAt:  &now,
	}
	if err := s.versions.CreateNext(v); err != nil {
		return nil, err
	}
	return v, nil
}

// Get 按 ID 查询版本。
func (s *VersionService) Get(id string) (*model.NodeVersion, error) {
	return s.versions.Get(id)
}

// ListByNode 列出某节点的全部版本。
func (s *VersionService) ListByNode(batchID, nodeNo string) ([]model.NodeVersion, error) {
	return s.versions.ListByNode(batchID, nodeNo)
}

// ListByBatch 列出某批次的全部版本。
func (s *VersionService) ListByBatch(batchID string) ([]model.NodeVersion, error) {
	return s.versions.ListByBatch(batchID)
}

// Snapshot 解析版本冻结的快照。
func (s *VersionService) Snapshot(v *model.NodeVersion) (*NodeSnapshot, error) {
	var snap NodeSnapshot
	if err := json.Unmarshal([]byte(v.SnapshotJSON), &snap); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}
	return &snap, nil
}

// findJoint 按节点号定位节点关系。
func (s *VersionService) findJoint(batchID, nodeNo string) (*model.JointRelation, error) {
	joints, err := s.joints.ListByBatch(batchID)
	if err != nil {
		return nil, err
	}
	for i := range joints {
		if joints[i].NodeNo == nodeNo {
			return &joints[i], nil
		}
	}
	return nil, fmt.Errorf("%w: node %s not found", model.ErrNotFound, nodeNo)
}

// buildSnapshot 组装节点快照。
func (s *VersionService) buildSnapshot(jrel *model.JointRelation) (*NodeSnapshot, error) {
	snap := &NodeSnapshot{NodeNo: jrel.NodeNo, Joint: jrel}

	pointIDs, err := joint.ParseIDList(jrel.PointIDs)
	if err != nil {
		return nil, err
	}
	for _, pid := range pointIDs {
		p, err := s.points.Get(pid)
		if err != nil {
			return nil, fmt.Errorf("snapshot point %s: %w", pid, err)
		}
		snap.Points = append(snap.Points, *p)
	}

	memberIDs, err := joint.ParseIDList(jrel.MemberIDs)
	if err != nil {
		return nil, err
	}
	for _, mid := range memberIDs {
		m, err := s.members.Get(mid)
		if err != nil {
			return nil, fmt.Errorf("snapshot member %s: %w", mid, err)
		}
		snap.Members = append(snap.Members, *m)
	}

	evidences, err := s.evidences.ListByJoint(jrel.ID)
	if err != nil {
		return nil, err
	}
	snap.Evidences = evidences

	return snap, nil
}

// supersedeFrozen 把该节点现存冻结版本标记为替代。
func (s *VersionService) supersedeFrozen(batchID, nodeNo string) error {
	vs, err := s.versions.ListByNode(batchID, nodeNo)
	if err != nil {
		return err
	}
	for _, v := range vs {
		if v.Status == model.VersionFrozen {
			if err := s.versions.UpdateStatus(v.ID, model.VersionSuperseded); err != nil {
				return err
			}
		}
	}
	return nil
}
