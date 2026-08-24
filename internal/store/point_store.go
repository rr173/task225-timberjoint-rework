package store

import (
	"database/sql"
	"fmt"

	"task225-timberjoint/internal/model"
)

// PointStore 是测点的持久化访问。
type PointStore struct{ db *DB }

// NewPointStore 构造测点存储。
func NewPointStore(db *DB) *PointStore { return &PointStore{db: db} }

const pointCols = `id, batch_id, point_no, x, y, z, error_semi_major, error_semi_minor,
	error_vertical, orientation_azim, status, fingerprint, created_at, updated_at`

// Create 插入新测点。
func (s *PointStore) Create(p *model.SurveyPoint) error {
	_, err := s.db.db.Exec(`INSERT INTO points (`+pointCols+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		p.ID, p.BatchID, p.PointNo, p.X, p.Y, p.Z,
		p.ErrorSemiMajor, p.ErrorSemiMinor, p.ErrorVertical, p.OrientationAzim,
		p.Status, p.Fingerprint, ts(p.CreatedAt), ts(p.UpdatedAt))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrDuplicate
		}
		return fmt.Errorf("insert point: %w", err)
	}
	return nil
}

// Get 按 ID 查询测点。
func (s *PointStore) Get(id string) (*model.SurveyPoint, error) {
	row := s.db.db.QueryRow(`SELECT `+pointCols+` FROM points WHERE id = ?`, id)
	return scanPoint(row)
}

// ListByBatch 列出某批次的全部测点（按编号排序）。
func (s *PointStore) ListByBatch(batchID string) ([]model.SurveyPoint, error) {
	rows, err := s.db.db.Query(`SELECT `+pointCols+` FROM points WHERE batch_id = ? ORDER BY point_no`, batchID)
	if err != nil {
		return nil, fmt.Errorf("list points: %w", err)
	}
	defer rows.Close()
	var out []model.SurveyPoint
	for rows.Next() {
		p, err := scanPoint(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// Update 更新测点（坐标/误差椭球/状态）。
func (s *PointStore) Update(p *model.SurveyPoint) error {
	_, err := s.db.db.Exec(`UPDATE points SET point_no=?, x=?, y=?, z=?,
		error_semi_major=?, error_semi_minor=?, error_vertical=?, orientation_azim=?,
		status=?, fingerprint=?, updated_at=? WHERE id=?`,
		p.PointNo, p.X, p.Y, p.Z,
		p.ErrorSemiMajor, p.ErrorSemiMinor, p.ErrorVertical, p.OrientationAzim,
		p.Status, p.Fingerprint, ts(p.UpdatedAt), p.ID)
	if err != nil {
		return fmt.Errorf("update point: %w", err)
	}
	return nil
}

// UpdateStatus 更新测点状态（标记接触异常/缺口）。
func (s *PointStore) UpdateStatus(id, status string) error {
	_, err := s.db.db.Exec(`UPDATE points SET status=?, updated_at=? WHERE id=?`,
		status, ts(nowUTC()), id)
	if err != nil {
		return fmt.Errorf("update point status: %w", err)
	}
	return nil
}

// CountByBatch 统计某批次测点数。
func (s *PointStore) CountByBatch(batchID string) (int, error) {
	var n int
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM points WHERE batch_id = ?`, batchID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count points: %w", err)
	}
	return n, nil
}

// Count 返回测点总数。
func (s *PointStore) Count() (int, error) {
	var n int
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM points`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count points: %w", err)
	}
	return n, nil
}

func scanPoint(sc scanner) (*model.SurveyPoint, error) {
	var p model.SurveyPoint
	var createdAt, updatedAt string
	err := sc.Scan(&p.ID, &p.BatchID, &p.PointNo, &p.X, &p.Y, &p.Z,
		&p.ErrorSemiMajor, &p.ErrorSemiMinor, &p.ErrorVertical, &p.OrientationAzim,
		&p.Status, &p.Fingerprint, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan point: %w", err)
	}
	p.CreatedAt, err = parseTS(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse point created_at: %w", err)
	}
	p.UpdatedAt, err = parseTS(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse point updated_at: %w", err)
	}
	return &p, nil
}
