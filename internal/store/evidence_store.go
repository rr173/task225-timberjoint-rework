package store

import (
	"database/sql"
	"fmt"

	"task225-timberjoint/internal/model"
)

// EvidenceStore 是照片证据的持久化访问。
type EvidenceStore struct{ db *DB }

// NewEvidenceStore 构造证据存储。
func NewEvidenceStore(db *DB) *EvidenceStore { return &EvidenceStore{db: db} }

const evidenceCols = `id, joint_id, batch_id, filename, caption, hash, taken_at, created_at`

// Create 插入新证据。
func (s *EvidenceStore) Create(e *model.PhotoEvidence) error {
	_, err := s.db.db.Exec(`INSERT INTO evidences (`+evidenceCols+`) VALUES (?,?,?,?,?,?,?,?)`,
		e.ID, e.JointID, e.BatchID, e.Filename, e.Caption, e.Hash, ts(e.TakenAt), ts(e.CreatedAt))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrDuplicate
		}
		if isFKViolation(err) {
			return model.ErrNotFound
		}
		return fmt.Errorf("insert evidence: %w", err)
	}
	return nil
}

// GetByJointHash 按节点关系 + 内容哈希查询证据（幂等去重）。
func (s *EvidenceStore) GetByJointHash(jointID, hash string) (*model.PhotoEvidence, error) {
	row := s.db.db.QueryRow(`SELECT `+evidenceCols+` FROM evidences WHERE joint_id = ? AND hash = ?`, jointID, hash)
	return scanEvidence(row)
}

// ListByJoint 列出某节点关系的全部证据（按拍摄时间倒序）。
func (s *EvidenceStore) ListByJoint(jointID string) ([]model.PhotoEvidence, error) {
	rows, err := s.db.db.Query(`SELECT `+evidenceCols+` FROM evidences WHERE joint_id = ? ORDER BY taken_at DESC`, jointID)
	if err != nil {
		return nil, fmt.Errorf("list evidences: %w", err)
	}
	defer rows.Close()
	var out []model.PhotoEvidence
	for rows.Next() {
		e, err := scanEvidence(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *e)
	}
	return out, rows.Err()
}

// ListByBatch 列出某批次的全部证据。
func (s *EvidenceStore) ListByBatch(batchID string) ([]model.PhotoEvidence, error) {
	rows, err := s.db.db.Query(`SELECT `+evidenceCols+` FROM evidences WHERE batch_id = ? ORDER BY created_at DESC`, batchID)
	if err != nil {
		return nil, fmt.Errorf("list evidences by batch: %w", err)
	}
	defer rows.Close()
	var out []model.PhotoEvidence
	for rows.Next() {
		e, err := scanEvidence(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *e)
	}
	return out, rows.Err()
}

// Count 返回证据总数。
func (s *EvidenceStore) Count() (int, error) {
	var n int
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM evidences`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count evidences: %w", err)
	}
	return n, nil
}

func scanEvidence(sc scanner) (*model.PhotoEvidence, error) {
	var e model.PhotoEvidence
	var takenAt, createdAt string
	err := sc.Scan(&e.ID, &e.JointID, &e.BatchID, &e.Filename, &e.Caption, &e.Hash, &takenAt, &createdAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan evidence: %w", err)
	}
	e.TakenAt, err = parseTS(takenAt)
	if err != nil {
		return nil, fmt.Errorf("parse evidence taken_at: %w", err)
	}
	e.CreatedAt, err = parseTS(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse evidence created_at: %w", err)
	}
	return &e, nil
}
