// Package service 编排层：组装各业务模块并暴露给 httpapi 与 main。
package service

import (
	"sync"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/store"
)

// App 应用编排根：聚合全部模块服务。
type App struct {
	db *store.DB

	Batches   *BatchService
	Points    *PointService
	Members   *MemberService
	Joints    *JointService
	Evidences *EvidenceService
	Versions  *VersionService

	stats *store.StatsStore

	mu sync.Mutex // 生命周期状态流转与版本发布串行
}

// New 组装应用服务。
func New(db *store.DB) (*App, error) {
	app := &App{
		db:    db,
		stats: store.NewStatsStore(db),
	}
	app.Batches = NewBatchService(store.NewBatchStore(db), store.NewJointStore(db), store.NewVersionStore(db))
	app.Points = NewPointService(store.NewPointStore(db), store.NewBatchStore(db))
	app.Members = NewMemberService(store.NewMemberStore(db), store.NewBatchStore(db))
	app.Joints = NewJointService(
		store.NewJointStore(db),
		store.NewPointStore(db),
		store.NewMemberStore(db),
		store.NewBatchStore(db),
	)
	app.Evidences = NewEvidenceService(
		store.NewEvidenceStore(db),
		store.NewJointStore(db),
		store.NewBatchStore(db),
	)
	app.Versions = NewVersionService(
		store.NewVersionStore(db),
		app.Joints,
		store.NewJointStore(db),
		store.NewPointStore(db),
		store.NewMemberStore(db),
		store.NewEvidenceStore(db),
		store.NewBatchStore(db),
	)
	return app, nil
}

// DB 暴露底层连接（自检用）。
func (a *App) DB() *store.DB { return a.db }

// Stats 返回全局统计快照。
func (a *App) Stats() (*model.StatSummary, error) { return a.stats.Global() }

// Lock 串行化关键区（批次流转、版本发布）。
func (a *App) Lock() func() {
	a.mu.Lock()
	return a.mu.Unlock
}

// WithLock serializes a lifecycle operation across all HTTP callers sharing
// this application instance.
func (a *App) WithLock(fn func() error) error {
	unlock := a.Lock()
	defer unlock()
	return fn()
}
