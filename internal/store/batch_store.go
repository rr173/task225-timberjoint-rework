package store

import (
	"database/sql"
	"fmt"

	"task225-timberjoint/internal/model"
)

// BatchStore 是测绘批次的持久化访问。
type BatchStore struct{ db *DB }

// NewBatchStore 构造批次存储。
func NewBatchStore(db *DB) *BatchStore { return &BatchStore{db: db} }

const batchCols = `id, name, description, coordinate_system, length_unit,
	closure_tolerance, direction_tol_deg, status, fingerprint, created_at, updated_at`

// Create 插入新批次。
func (s *BatchStore) Create(b *model.SurveyBatch) error {
	_, err := s.db.db.Exec(`INSERT INTO batches (`+batchCols+`) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		b.ID, b.Name, b.Description, b.CoordinateSystem, b.LengthUnit,
		b.ClosureTolerance, b.DirectionTolDeg, b.Status, b.Fingerprint,
		ts(b.CreatedAt), ts(b.UpdatedAt))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrDuplicate
		}
		return fmt.Errorf("insert batch: %w", err)
	}
	return nil
}

// Get 按 ID 查询批次。
func (s *BatchStore) Get(id string) (*model.SurveyBatch, error) {
	row := s.db.db.QueryRow(`SELECT `+batchCols+` FROM batches WHERE id = ?`, id)
	return scanBatch(row)
}

// GetByFingerprint 按幂等指纹查询批次（存在即重复）。
func (s *BatchStore) GetByFingerprint(fp string) (*model.SurveyBatch, error) {
	row := s.db.db.QueryRow(`SELECT `+batchCols+` FROM batches WHERE fingerprint = ?`, fp)
	return scanBatch(row)
}

// List 列出全部批次（按创建时间倒序）。
func (s *BatchStore) List() ([]model.SurveyBatch, error) {
	rows, err := s.db.db.Query(`SELECT ` + batchCols + ` FROM batches ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list batches: %w", err)
	}
	defer rows.Close()
	var out []model.SurveyBatch
	for rows.Next() {
		b, err := scanBatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *b)
	}
	return out, rows.Err()
}

// Update 更新批次字段（整体覆盖非 ID 字段）。
func (s *BatchStore) Update(b *model.SurveyBatch) error {
	_, err := s.db.db.Exec(`UPDATE batches SET name=?, description=?, coordinate_system=?,
		length_unit=?, closure_tolerance=?, direction_tol_deg=?, status=?, fingerprint=?,
		updated_at=? WHERE id=?`,
		b.Name, b.Description, b.CoordinateSystem, b.LengthUnit,
		b.ClosureTolerance, b.DirectionTolDeg, b.Status, b.Fingerprint,
		ts(b.UpdatedAt), b.ID)
	if err != nil {
		return fmt.Errorf("update batch: %w", err)
	}
	return nil
}

// UpdateStatus 更新批次状态（状态机流转由 service 校验）。
func (s *BatchStore) UpdateStatus(id, status string) error {
	_, err := s.db.db.Exec(`UPDATE batches SET status=?, updated_at=? WHERE id=?`,
		status, ts(nowUTC()), id)
	if err != nil {
		return fmt.Errorf("update batch status: %w", err)
	}
	return nil
}

// Count 返回批次总数。
func (s *BatchStore) Count() (int, error) {
	var n int
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM batches`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count batches: %w", err)
	}
	return n, nil
}

func scanBatch(sc scanner) (*model.SurveyBatch, error) {
	var b model.SurveyBatch
	var createdAt, updatedAt string
	err := sc.Scan(&b.ID, &b.Name, &b.Description, &b.CoordinateSystem, &b.LengthUnit,
		&b.ClosureTolerance, &b.DirectionTolDeg, &b.Status, &b.Fingerprint,
		&createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan batch: %w", err)
	}
	b.CreatedAt, err = parseTS(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse batch created_at: %w", err)
	}
	b.UpdatedAt, err = parseTS(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse batch updated_at: %w", err)
	}
	return &b, nil
}
