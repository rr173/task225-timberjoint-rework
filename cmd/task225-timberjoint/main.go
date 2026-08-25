// Command task225-timberjoint 古建筑木构节点测绘复核台服务入口。
//
// 支持三个标志：
//   - --addr :8080        监听地址（默认 :8080）
//   - --db ./task225-timberjoint.db  SQLite 数据库路径
//   - --smoke-test        执行端到端冒烟：真实创建数据、关闭并重开数据库
//     验证持久化与重启恢复，随后以 0 退出码结束。
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"task225-timberjoint/internal/httpapi"
	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/service"
	"task225-timberjoint/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "./task225-timberjoint.db", "SQLite database path")
	smoke := flag.Bool("smoke-test", false, "run end-to-end smoke test and exit")
	flag.Parse()

	if *smoke {
		if err := runSmokeTest(*dbPath); err != nil {
			fmt.Fprintln(os.Stderr, "SMOKE TEST FAILED:", err)
			os.Exit(1)
		}
		fmt.Println("SMOKE TEST PASSED")
		return
	}

	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	app, err := service.New(db)
	if err != nil {
		log.Fatalf("init services: %v", err)
	}
	srv := httpapi.New(app)
	log.Printf("task225-timberjoint listening on %s (db=%s)", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

// runSmokeTest 执行端到端冒烟：
//  1. 打开数据库 A，创建批次 -> 加测点/构件 -> 建立节点关系 ->
//     标记接触异常 -> 几何复核（检测方向冲突）-> 修正端点 -> 重检闭合 ->
//     确认 -> 登记照片证据 -> 发布节点版本 -> 发布并封存批次；
//  2. 幂等验证：重复测点/证据被跳过；
//  3. 封存后拒绝再写入测点；
//  4. 关闭数据库 A，重开同一路径数据库 B，验证数据仍在（重启恢复）。
func runSmokeTest(dbPath string) error {
	if dbPath != ":memory:" {
		_ = os.Remove(dbPath)
	}

	db, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	app, err := service.New(db)
	if err != nil {
		db.Close()
		return fmt.Errorf("init services: %w", err)
	}

	// --- 步骤 1：创建测绘批次 ---
	batch, err := app.Batches.Create(service.BatchCreateInput{
		Name:             "样例木构节点测绘批次 A",
		Description:      "柱梁枋节点 N001 的现场测绘复核",
		CoordinateSystem: "CGCS2000",
		LengthUnit:       "mm",
		ClosureTolerance: 50.0,
		DirectionTolDeg:  10.0,
	})
	if err != nil {
		db.Close()
		return fmt.Errorf("create batch: %w", err)
	}

	// --- 步骤 2：加 4 个闭合测点 + 1 个接触异常测点 ---
	type ptSpec struct {
		no string
		x  float64
		y  float64
		z  float64
	}
	closedPts := []ptSpec{
		{"P001", 1000, 1000, 1000},
		{"P002", 1010, 1000, 1000},
		{"P003", 1000, 1010, 1000},
		{"P004", 1000, 1000, 1010},
	}
	var pointIDs []string
	for _, p := range closedPts {
		pt, err := app.Points.Add(batch.ID, service.PointAddInput{
			No: p.no, X: p.x, Y: p.y, Z: p.z,
			SemiMajor: 5, SemiMinor: 5, Vertical: 5,
		})
		if err != nil {
			db.Close()
			return fmt.Errorf("add point %s: %w", p.no, err)
		}
		pointIDs = append(pointIDs, pt.ID)
	}
	// 接触异常测点（远点，将标记为 contact_error，不参与闭合）。
	contactPt, err := app.Points.Add(batch.ID, service.PointAddInput{
		No: "P005", X: 3000, Y: 3000, Z: 3000,
		SemiMajor: 5, SemiMinor: 5, Vertical: 5,
	})
	if err != nil {
		db.Close()
		return fmt.Errorf("add contact point: %w", err)
	}
	pointIDs = append(pointIDs, contactPt.ID)

	// 幂等验证：重复加同一测点被跳过。
	if _, err := app.Points.Add(batch.ID, service.PointAddInput{
		No: "P001", X: 1000, Y: 1000, Z: 1000,
		SemiMajor: 5, SemiMinor: 5, Vertical: 5,
	}); err != model.ErrDuplicate {
		db.Close()
		return fmt.Errorf("duplicate point should be rejected, got %v", err)
	}

	// --- 步骤 3：加 3 个构件（M003 初始方向偏斜 45°） ---
	m1, err := app.Members.Add(batch.ID, service.MemberAddInput{
		No: "M001", MemberType: "梁",
		StartX: 0, StartY: 0, StartZ: 0, EndX: 1000, EndY: 0, EndZ: 0,
	})
	if err != nil {
		db.Close()
		return fmt.Errorf("add member M001: %w", err)
	}
	m2, err := app.Members.Add(batch.ID, service.MemberAddInput{
		No: "M002", MemberType: "梁",
		StartX: 0, StartY: 500, StartZ: 0, EndX: 1000, EndY: 500, EndZ: 0,
	})
	if err != nil {
		db.Close()
		return fmt.Errorf("add member M002: %w", err)
	}
	m3, err := app.Members.Add(batch.ID, service.MemberAddInput{
		No: "M003", MemberType: "梁",
		StartX: 0, StartY: 1000, StartZ: 0, EndX: 707, EndY: 1707, EndZ: 0,
	})
	if err != nil {
		db.Close()
		return fmt.Errorf("add member M003: %w", err)
	}
	memberIDs := []string{m1.ID, m2.ID, m3.ID}

	// --- 步骤 4：流转到待复核（collecting -> organizing -> reviewing） ---
	if _, err := app.Batches.Advance(batch.ID); err != nil {
		db.Close()
		return fmt.Errorf("advance to organizing: %w", err)
	}
	if _, err := app.Batches.Advance(batch.ID); err != nil {
		db.Close()
		return fmt.Errorf("advance to reviewing: %w", err)
	}

	// --- 步骤 5：建立节点关系 N001 ---
	jrel, err := app.Joints.Create(batch.ID, "N001", pointIDs, memberIDs)
	if err != nil {
		db.Close()
		return fmt.Errorf("create joint: %w", err)
	}

	// --- 步骤 6：标记接触异常测点 ---
	if _, err := app.Points.MarkStatus(contactPt.ID, model.PointContactError); err != nil {
		db.Close()
		return fmt.Errorf("mark contact error: %w", err)
	}

	// --- 步骤 7：第一次复核 —— 检测到方向冲突 ---
	rep1, err := app.Joints.Check(jrel.ID)
	if err != nil {
		db.Close()
		return fmt.Errorf("check joint (first): %w", err)
	}
	if !rep1.DirectionConflict {
		db.Close()
		return fmt.Errorf("expected direction conflict, got spread=%.2f", rep1.DirectionSpread)
	}
	// 冲突时节点关系应判为断裂。
	jrelAfter, _ := app.Joints.Get(jrel.ID)
	if jrelAfter.Status != model.JointBroken {
		db.Close()
		return fmt.Errorf("joint should be broken on direction conflict, got %s", jrelAfter.Status)
	}

	// --- 步骤 8：修正 M003 端点，使其与其余梁平行 ---
	if _, err := app.Members.Update(m3.ID, service.MemberAddInput{
		No: "M003", MemberType: "梁",
		StartX: 0, StartY: 1000, StartZ: 0, EndX: 1000, EndY: 1000, EndZ: 0,
	}); err != nil {
		db.Close()
		return fmt.Errorf("fix member M003: %w", err)
	}

	// --- 步骤 9：重检 —— 闭合 ---
	rep2, err := app.Joints.Check(jrel.ID)
	if err != nil {
		db.Close()
		return fmt.Errorf("check joint (second): %w", err)
	}
	if !rep2.Closed || rep2.DirectionConflict {
		db.Close()
		return fmt.Errorf("joint should close after fix, got closed=%v conflict=%v", rep2.Closed, rep2.DirectionConflict)
	}
	jrelAfter, _ = app.Joints.Get(jrel.ID)
	if jrelAfter.Status != model.JointClosed {
		db.Close()
		return fmt.Errorf("joint should be closed, got %s", jrelAfter.Status)
	}

	// --- 步骤 10：确认节点关系 ---
	if _, err := app.Joints.Confirm(jrel.ID); err != nil {
		db.Close()
		return fmt.Errorf("confirm joint: %w", err)
	}

	// --- 步骤 11：登记照片证据（幂等去重） ---
	takenAt := time.Now().UTC().Truncate(time.Second)
	ev, err := app.Evidences.Add(jrel.ID, "node_n001_north.jpg", "节点 N001 北侧榫卯连接", takenAt)
	if err != nil {
		db.Close()
		return fmt.Errorf("add evidence: %w", err)
	}
	if _, err := app.Evidences.Add(jrel.ID, "node_n001_north.jpg", "节点 N001 北侧榫卯连接", takenAt); err != model.ErrDuplicate {
		db.Close()
		return fmt.Errorf("duplicate evidence should be rejected, got %v", err)
	}
	if ev.Hash == "" {
		db.Close()
		return fmt.Errorf("evidence hash should not be empty")
	}

	// --- 步骤 12：发布节点版本 ---
	ver, err := app.Versions.Publish(batch.ID, "N001", "方向冲突修正后复核通过，发布供修缮方案引用")
	if err != nil {
		db.Close()
		return fmt.Errorf("publish version: %w", err)
	}
	if ver.Status != model.VersionFrozen {
		db.Close()
		return fmt.Errorf("version should be frozen, got %s", ver.Status)
	}

	// --- 步骤 13：发布并封存批次 ---
	if _, err := app.Batches.Advance(batch.ID); err != nil {
		db.Close()
		return fmt.Errorf("advance to published: %w", err)
	}
	if _, err := app.Batches.Archive(batch.ID); err != nil {
		db.Close()
		return fmt.Errorf("archive batch: %w", err)
	}

	// --- 步骤 14：封存后拒绝再写入测点 ---
	if _, err := app.Points.Add(batch.ID, service.PointAddInput{
		No: "P999", X: 1, Y: 1, Z: 1,
		SemiMajor: 1, SemiMinor: 1, Vertical: 1,
	}); err == nil {
		db.Close()
		return fmt.Errorf("archived batch should reject new point")
	}

	// --- 步骤 15：重启恢复 ---
	db.Close()

	db2, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("reopen db: %w", err)
	}
	defer db2.Close()
	app2, err := service.New(db2)
	if err != nil {
		return fmt.Errorf("reinit services: %w", err)
	}
	b2, err := app2.Batches.Get(batch.ID)
	if err != nil {
		return fmt.Errorf("get batch after reopen: %w", err)
	}
	if b2.Status != model.BatchArchived {
		return fmt.Errorf("batch status after reopen should be archived, got %s", b2.Status)
	}
	vers2, err := app2.Versions.ListByNode(batch.ID, "N001")
	if err != nil {
		return fmt.Errorf("list versions after reopen: %w", err)
	}
	if len(vers2) == 0 || vers2[0].Status != model.VersionFrozen {
		return fmt.Errorf("frozen version missing after reopen")
	}
	evidences2, err := app2.Evidences.ListByJoint(jrel.ID)
	if err != nil {
		return fmt.Errorf("list evidences after reopen: %w", err)
	}
	if len(evidences2) == 0 {
		return fmt.Errorf("evidence missing after reopen")
	}
	j2, err := app2.Joints.Get(jrel.ID)
	if err != nil {
		return fmt.Errorf("get joint after reopen: %w", err)
	}
	if j2.Status != model.JointConfirmed {
		return fmt.Errorf("joint status after reopen should be confirmed, got %s", j2.Status)
	}
	return nil
}
