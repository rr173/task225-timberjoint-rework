package store

import (
	"fmt"

	"task225-timberjoint/internal/model"
)

// StatsStore 提供全局统计查询。
type StatsStore struct{ db *DB }

// NewStatsStore 构造统计存储。
func NewStatsStore(db *DB) *StatsStore { return &StatsStore{db: db} }

// Global 返回全局统计快照。
func (s *StatsStore) Global() (*model.StatSummary, error) {
	st := &model.StatSummary{}
	counts := []struct {
		name string
		dst  *int
	}{
		{"batches", &st.Batches},
		{"points", &st.Points},
		{"members", &st.Members},
		{"joints", &st.Joints},
		{"evidences", &st.Evidences},
		{"versions", &st.Versions},
	}
	for _, c := range counts {
		var n int
		if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM ` + c.name).Scan(&n); err != nil {
			return nil, fmt.Errorf("count %s: %w", c.name, err)
		}
		*c.dst = n
	}
	// 未封存批次（采集/整理/待复核/已发布）。
	if err := s.db.db.QueryRow(`SELECT COUNT(*) FROM batches WHERE status != ?`, model.BatchArchived).Scan(&st.OpenBatches); err != nil {
		return nil, fmt.Errorf("count open batches: %w", err)
	}
	return st, nil
}
