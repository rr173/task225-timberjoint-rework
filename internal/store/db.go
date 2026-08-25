// Package store 提供基于 SQLite（modernc.org/sqlite，纯 Go 驱动，CGO 无关）的
// 持久化实现：建表迁移与批次/测点/构件/节点关系/证据/版本的 CRUD 及统计查询。
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Open 打开（或创建）SQLite 数据库并执行迁移。
// 支持 ":memory:" 用于测试与冒烟验证。
func Open(path string) (*DB, error) {
	if path == ":memory:" {
		db, err := sql.Open("sqlite", "file::memory:?cache=shared")
		if err != nil {
			return nil, fmt.Errorf("open memory db: %w", err)
		}
		db.SetMaxOpenConns(1)
		d := &DB{db: db, path: path}
		if err := d.migrate(); err != nil {
			db.Close()
			return nil, err
		}
		return d, nil
	}

	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("mkdir db dir: %w", err)
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set wal: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable fk: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}
	d := &DB{db: db, path: path}
	if err := d.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return d, nil
}

// DB 封装 SQLite 连接。
type DB struct {
	db   *sql.DB
	path string
}

// Close 关闭数据库连接。
func (d *DB) Close() error { return d.db.Close() }

// withTx executes a persistence operation atomically on the single SQLite
// connection used by the application.
func (d *DB) withTx(fn func(*sql.Tx) error) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// Path 返回数据库路径（调试用）。
func (d *DB) Path() string { return d.path }

// SQL 暴露底层连接（供 Store 实现使用）。
func (d *DB) SQL() *sql.DB { return d.db }

// migrate 建表：全部业务表 + 唯一约束（防重复）。
func (d *DB) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS batches (
			id                 TEXT PRIMARY KEY,
			name               TEXT NOT NULL,
			description        TEXT NOT NULL DEFAULT '',
			coordinate_system  TEXT NOT NULL,
			length_unit        TEXT NOT NULL,
			closure_tolerance  REAL NOT NULL,
			direction_tol_deg  REAL NOT NULL,
			status             TEXT NOT NULL,
			fingerprint        TEXT NOT NULL,
			created_at         TEXT NOT NULL,
			updated_at         TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_batches_fingerprint ON batches(fingerprint)`,
		`CREATE TABLE IF NOT EXISTS points (
			id                TEXT PRIMARY KEY,
			batch_id          TEXT NOT NULL REFERENCES batches(id),
			point_no          TEXT NOT NULL,
			x                 REAL NOT NULL,
			y                 REAL NOT NULL,
			z                 REAL NOT NULL,
			error_semi_major  REAL NOT NULL,
			error_semi_minor  REAL NOT NULL,
			error_vertical    REAL NOT NULL,
			orientation_azim  REAL NOT NULL,
			status            TEXT NOT NULL,
			fingerprint       TEXT NOT NULL,
			created_at        TEXT NOT NULL,
			updated_at        TEXT NOT NULL,
			UNIQUE(batch_id, point_no)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_points_batch ON points(batch_id)`,
		`CREATE TABLE IF NOT EXISTS members (
			id            TEXT PRIMARY KEY,
			batch_id      TEXT NOT NULL REFERENCES batches(id),
			member_no     TEXT NOT NULL,
			member_type   TEXT NOT NULL,
			start_x       REAL NOT NULL,
			start_y       REAL NOT NULL,
			start_z       REAL NOT NULL,
			end_x         REAL NOT NULL,
			end_y         REAL NOT NULL,
			end_z         REAL NOT NULL,
			tenon_desc    TEXT NOT NULL DEFAULT '',
			status        TEXT NOT NULL,
			created_at    TEXT NOT NULL,
			updated_at    TEXT NOT NULL,
			UNIQUE(batch_id, member_no)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_members_batch ON members(batch_id)`,
		`CREATE TABLE IF NOT EXISTS joints (
			id                TEXT PRIMARY KEY,
			batch_id          TEXT NOT NULL REFERENCES batches(id),
			node_no           TEXT NOT NULL,
			point_ids         TEXT NOT NULL DEFAULT '[]',
			member_ids        TEXT NOT NULL DEFAULT '[]',
			status            TEXT NOT NULL,
			closure_residual  REAL NOT NULL DEFAULT 0,
			direction_spread  REAL NOT NULL DEFAULT 0,
			check_report      TEXT NOT NULL DEFAULT '',
			created_at        TEXT NOT NULL,
			updated_at        TEXT NOT NULL,
			UNIQUE(batch_id, node_no)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_joints_batch ON joints(batch_id)`,
		`CREATE TABLE IF NOT EXISTS evidences (
			id          TEXT PRIMARY KEY,
			joint_id    TEXT NOT NULL REFERENCES joints(id),
			batch_id    TEXT NOT NULL REFERENCES batches(id),
			filename    TEXT NOT NULL,
			caption     TEXT NOT NULL,
			hash        TEXT NOT NULL,
			taken_at    TEXT NOT NULL,
			created_at  TEXT NOT NULL,
			UNIQUE(joint_id, hash)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_evidences_joint ON evidences(joint_id)`,
		`CREATE TABLE IF NOT EXISTS versions (
			id            TEXT PRIMARY KEY,
			batch_id      TEXT NOT NULL REFERENCES batches(id),
			node_no       TEXT NOT NULL,
			version_no    INTEGER NOT NULL,
			status        TEXT NOT NULL,
			snapshot_json TEXT NOT NULL DEFAULT '{}',
			reason        TEXT NOT NULL DEFAULT '',
			created_at    TEXT NOT NULL,
			published_at  TEXT,
			UNIQUE(batch_id, node_no, version_no)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_versions_batch ON versions(batch_id)`,
	}
	for _, s := range stmts {
		if _, err := d.db.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w (%s)", err, firstWords(s, 10))
		}
	}
	return nil
}

func firstWords(s string, n int) string {
	parts := strings.Fields(s)
	if len(parts) > n {
		parts = parts[:n]
	}
	return strings.Join(parts, " ")
}

func nowUTC() time.Time { return time.Now().UTC() }

func ts(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }

func parseTS(s string) (time.Time, error) { return time.Parse(time.RFC3339Nano, s) }

type scanner interface{ Scan(dest ...any) error }
