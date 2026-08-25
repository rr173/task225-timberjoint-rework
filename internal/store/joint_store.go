package store

import (
	"database/sql"
	"fmt"

	"task225-timberjoint/internal/model"
)

// JointStore 是节点关系的持久化访问。
type JointStore struct{ db *DB }

// NewJointStore 构造节点关系存储。
func NewJointStore(db *DB) *JointStore { return &JointStore{db: db} }

const jointCols = `id, batch_id, node_no, point_ids, member_ids, status,
	closure_residual, direction_spread, check_report, created_at, updated_at`

// Create 插入新节点关系。
func (s *JointStore) Create(j *model.JointRelation) error {
	_, err := s.db.db.Exec(`INSERT INTO joints (`+jointCols+`) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		j.ID, j.BatchID, j.NodeNo, j.PointIDs, j.MemberIDs, j.Status,
		j.ClosureResidual, j.DirectionSpread, j.CheckReport, ts(j.CreatedAt), ts(j.UpdatedAt))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrDuplicate
		}
		return fmt.Errorf("insert joint: %w", err)
	}
	return nil
}

// Get 按 ID 查询节点关系。
func (s *JointStore) Get(id string) (*model.JointRelation, error) {
	row := s.db.db.QueryRow(`SELECT `+jointCols+` FROM joints WHERE id = ?`, id)
	return scanJoint(row)
}

// ListByBatch 列出某批次的全部节点关系（按编号排序）。
func (s *JointStore) ListByBatch(batchID string) ([]model.JointRelation, error) {
	rows, err := s.db.db.Query(`SELECT `+jointCols+` FROM joints WHERE batch_id = ? ORDER BY node_no`, batchID)
	if err != nil {
		return nil, fmt.Errorf("list joints: %w", err)
	}
	defer rows.Close()
	var out []model.JointRelation
	for rows.Next() {
		j, err := scanJoint(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *j)
	}
	return out, rows.Err()
}

// AllConfirmed reports whether a batch has at least one joint and every joint
// has reached the confirmed state required for batch publication.
func (s *JointStore) AllConfirmed(batchID string) (bool, error) {
	var total, unresolved int
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM joints WHERE batch_id = ?`, batchID).Scan(&total); err != nil {
		return false, fmt.Errorf("count batch joints: %w", err)
	}
	if total == 0 {
		return false, nil
	}
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM joints WHERE batch_id = ? AND status <> ?`, batchID, model.JointConfirmed).Scan(&unresolved); err != nil {
		return false, fmt.Errorf("count unresolved joints: %w", err)
	}
	return unresolved == 0, nil
}

// Update 更新节点关系（几何指标与报告）。
func (s *JointStore) Update(j *model.JointRelation) error {
	_, err := s.db.db.Exec(`UPDATE joints SET node_no=?, point_ids=?, member_ids=?,
		status=?, closure_residual=?, direction_spread=?, check_report=?, updated_at=? WHERE id=?`,
		j.NodeNo, j.PointIDs, j.MemberIDs, j.Status,
		j.ClosureResidual, j.DirectionSpread, j.CheckReport, ts(j.UpdatedAt), j.ID)
	if err != nil {
		return fmt.Errorf("update joint: %w", err)
	}
	return nil
}

// UpdateStatus 更新节点关系状态。
func (s *JointStore) UpdateStatus(id, status string) error {
	_, err := s.db.db.Exec(`UPDATE joints SET status=?, updated_at=? WHERE id=?`,
		status, ts(nowUTC()), id)
	if err != nil {
		return fmt.Errorf("update joint status: %w", err)
	}
	return nil
}

// UpdateStatusFrom performs a compare-and-set state transition.
func (s *JointStore) UpdateStatusFrom(id, from, to string) error {
	res, err := s.db.db.Exec(`UPDATE joints SET status=?, updated_at=? WHERE id=? AND status=?`,
		to, ts(nowUTC()), id, from)
	if err != nil {
		return fmt.Errorf("update joint status: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check joint status update: %w", err)
	}
	if n == 0 {
		return model.ErrConflict
	}
	return nil
}

// Count 返回节点关系总数。
func (s *JointStore) Count() (int, error) {
	var n int
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM joints`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count joints: %w", err)
	}
	return n, nil
}

func scanJoint(sc scanner) (*model.JointRelation, error) {
	var j model.JointRelation
	var createdAt, updatedAt string
	err := sc.Scan(&j.ID, &j.BatchID, &j.NodeNo, &j.PointIDs, &j.MemberIDs, &j.Status,
		&j.ClosureResidual, &j.DirectionSpread, &j.CheckReport, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan joint: %w", err)
	}
	j.CreatedAt, err = parseTS(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse joint created_at: %w", err)
	}
	j.UpdatedAt, err = parseTS(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse joint updated_at: %w", err)
	}
	return &j, nil
}
