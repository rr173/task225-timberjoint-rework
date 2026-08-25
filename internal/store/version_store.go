package store

import (
	"database/sql"
	"fmt"

	"task225-timberjoint/internal/model"
)

// VersionStore 是节点版本的持久化访问。
type VersionStore struct{ db *DB }

// NewVersionStore 构造版本存储。
func NewVersionStore(db *DB) *VersionStore { return &VersionStore{db: db} }

const versionCols = `id, batch_id, node_no, version_no, status, snapshot_json, reason, created_at, published_at`

// Create 插入新版本。
func (s *VersionStore) Create(v *model.NodeVersion) error {
	var publishedAt any
	if v.PublishedAt != nil {
		publishedAt = ts(*v.PublishedAt)
	}
	_, err := s.db.db.Exec(`INSERT INTO versions (`+versionCols+`) VALUES (?,?,?,?,?,?,?,?,?)`,
		v.ID, v.BatchID, v.NodeNo, v.VersionNo, v.Status, v.SnapshotJSON, v.Reason,
		ts(v.CreatedAt), publishedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrDuplicate
		}
		return fmt.Errorf("insert version: %w", err)
	}
	return nil
}

// Get 按 ID 查询版本。
func (s *VersionStore) Get(id string) (*model.NodeVersion, error) {
	row := s.db.db.QueryRow(`SELECT `+versionCols+` FROM versions WHERE id = ?`, id)
	return scanVersion(row)
}

// ListByNode 列出某节点（按批次+节点号）的全部版本（版本号倒序）。
func (s *VersionStore) ListByNode(batchID, nodeNo string) ([]model.NodeVersion, error) {
	rows, err := s.db.db.Query(`SELECT `+versionCols+` FROM versions WHERE batch_id = ? AND node_no = ? ORDER BY version_no DESC`, batchID, nodeNo)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	defer rows.Close()
	var out []model.NodeVersion
	for rows.Next() {
		v, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

// ListByBatch 列出某批次的全部版本。
func (s *VersionStore) ListByBatch(batchID string) ([]model.NodeVersion, error) {
	rows, err := s.db.db.Query(`SELECT `+versionCols+` FROM versions WHERE batch_id = ? ORDER BY node_no, version_no DESC`, batchID)
	if err != nil {
		return nil, fmt.Errorf("list versions by batch: %w", err)
	}
	defer rows.Close()
	var out []model.NodeVersion
	for rows.Next() {
		v, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

// AllNodesFrozen reports whether every node in a batch has at least one
// currently frozen version ready for archival.
func (s *VersionStore) AllNodesFrozen(batchID string) (bool, error) {
	var total, missing int
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM joints WHERE batch_id = ?`, batchID).Scan(&total); err != nil {
		return false, fmt.Errorf("count batch nodes: %w", err)
	}
	if total == 0 {
		return false, nil
	}
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM joints j WHERE j.batch_id = ? AND NOT EXISTS (
		SELECT 1 FROM versions v WHERE v.batch_id = j.batch_id AND v.node_no = j.node_no AND v.status = ?
	)`, batchID, model.VersionFrozen).Scan(&missing); err != nil {
		return false, fmt.Errorf("count nodes without frozen versions: %w", err)
	}
	return missing == 0, nil
}

// MaxVersionNo 返回某节点当前最大版本号（无版本返回 0）。
func (s *VersionStore) MaxVersionNo(batchID, nodeNo string) (int, error) {
	var n sql.NullInt64
	if err := s.db.db.QueryRow(`SELECT MAX(version_no) FROM versions WHERE batch_id = ? AND node_no = ?`, batchID, nodeNo).Scan(&n); err != nil {
		return 0, fmt.Errorf("max version no: %w", err)
	}
	if !n.Valid {
		return 0, nil
	}
	return int(n.Int64), nil
}

// UpdateStatus 更新版本状态与发布时间。
func (s *VersionStore) UpdateStatus(id, status string) error {
	_, err := s.db.db.Exec(`UPDATE versions SET status=?, published_at=? WHERE id=?`,
		status, ts(nowUTC()), id)
	if err != nil {
		return fmt.Errorf("update version status: %w", err)
	}
	return nil
}

// Count 返回版本总数。
func (s *VersionStore) Count() (int, error) {
	var n int
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM versions`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count versions: %w", err)
	}
	return n, nil
}

func scanVersion(sc scanner) (*model.NodeVersion, error) {
	var v model.NodeVersion
	var createdAt, publishedAt sql.NullString
	err := sc.Scan(&v.ID, &v.BatchID, &v.NodeNo, &v.VersionNo, &v.Status,
		&v.SnapshotJSON, &v.Reason, &createdAt, &publishedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan version: %w", err)
	}
	v.CreatedAt, err = parseTS(createdAt.String)
	if err != nil {
		return nil, fmt.Errorf("parse version created_at: %w", err)
	}
	if publishedAt.Valid {
		t, err := parseTS(publishedAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse version published_at: %w", err)
		}
		v.PublishedAt = &t
	}
	return &v, nil
}
