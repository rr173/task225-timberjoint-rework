package store

import (
	"database/sql"
	"fmt"

	"task225-timberjoint/internal/model"
)

// MemberStore 是木构构件的持久化访问。
type MemberStore struct{ db *DB }

// NewMemberStore 构造构件存储。
func NewMemberStore(db *DB) *MemberStore { return &MemberStore{db: db} }

const memberCols = `id, batch_id, member_no, member_type, start_x, start_y, start_z,
	end_x, end_y, end_z, tenon_desc, status, created_at, updated_at`

// Create 插入新构件。
func (s *MemberStore) Create(m *model.Member) error {
	_, err := s.db.db.Exec(`INSERT INTO members (`+memberCols+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		m.ID, m.BatchID, m.MemberNo, m.MemberType,
		m.StartX, m.StartY, m.StartZ, m.EndX, m.EndY, m.EndZ,
		m.TenonDesc, m.Status, ts(m.CreatedAt), ts(m.UpdatedAt))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrDuplicate
		}
		return fmt.Errorf("insert member: %w", err)
	}
	return nil
}

// Get 按 ID 查询构件。
func (s *MemberStore) Get(id string) (*model.Member, error) {
	row := s.db.db.QueryRow(`SELECT `+memberCols+` FROM members WHERE id = ?`, id)
	return scanMember(row)
}

// GetInBatch returns a member only when it belongs to the requested batch.
func (s *MemberStore) GetInBatch(id, batchID string) (*model.Member, error) {
	row := s.db.db.QueryRow(`SELECT `+memberCols+` FROM members WHERE id = ? AND batch_id = ?`, id, batchID)
	return scanMember(row)
}

// ListByBatch 列出某批次的全部构件（按编号排序）。
func (s *MemberStore) ListByBatch(batchID string) ([]model.Member, error) {
	rows, err := s.db.db.Query(`SELECT `+memberCols+` FROM members WHERE batch_id = ? ORDER BY member_no`, batchID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()
	var out []model.Member
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

// Update 更新构件。状态沿用 service 给定的值（修正端点后由复核重新裁定）。
func (s *MemberStore) Update(m *model.Member) error {
	_, err := s.db.db.Exec(`UPDATE members SET member_no=?, member_type=?,
		start_x=?, start_y=?, start_z=?, end_x=?, end_y=?, end_z=?,
		tenon_desc=?, status=?, updated_at=? WHERE id=?`,
		m.MemberNo, m.MemberType,
		m.StartX, m.StartY, m.StartZ, m.EndX, m.EndY, m.EndZ,
		m.TenonDesc, m.Status, ts(m.UpdatedAt), m.ID)
	if err != nil {
		return fmt.Errorf("update member: %w", err)
	}
	return nil
}

// UpdateStatus 更新构件状态（匹配/方向冲突/缺失）。
func (s *MemberStore) UpdateStatus(id, status string) error {
	_, err := s.db.db.Exec(`UPDATE members SET status=?, updated_at=? WHERE id=?`,
		status, ts(nowUTC()), id)
	if err != nil {
		return fmt.Errorf("update member status: %w", err)
	}
	return nil
}

// Count 返回构件总数。
func (s *MemberStore) Count() (int, error) {
	var n int
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM members`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count members: %w", err)
	}
	return n, nil
}

func scanMember(sc scanner) (*model.Member, error) {
	var m model.Member
	var createdAt, updatedAt string
	err := sc.Scan(&m.ID, &m.BatchID, &m.MemberNo, &m.MemberType,
		&m.StartX, &m.StartY, &m.StartZ, &m.EndX, &m.EndY, &m.EndZ,
		&m.TenonDesc, &m.Status, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan member: %w", err)
	}
	m.CreatedAt, err = parseTS(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse member created_at: %w", err)
	}
	m.UpdatedAt, err = parseTS(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse member updated_at: %w", err)
	}
	return &m, nil
}
