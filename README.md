# task225-timberjoint — 古建筑木构节点测绘复核台

对古建筑木构节点（柱/梁/枋/斗/拱）的现场测绘数据进行几何复核：检查测点闭合、
构件方向冲突与编号一致性，登记现场照片证据，并发布不可变的节点版本供修缮方案引用。

## 业务闭环

1. 创建测绘批次，声明坐标系、长度单位与闭合/方向容差；
2. 上传测点（三维坐标 + 误差椭球三半轴）与构件（两端点 + 榫卯描述）；
3. 建立节点关系，把一组测点与构件端点关联到某节点；
4. 几何复核：闭合检测（质心 + 最大残差）、方向冲突（构件方向散布角）、编号一致性；
5. 标记接触异常/缺口测点，修正构件端点后重检；
6. 登记现场照片证据（幂等去重）；
7. 发布节点版本（冻结快照），发布并封存批次。

## 标准命令

```bash
# 构建 / 静态检查 / 测试
CGO_ENABLED=0 go build ./...
CGO_ENABLED=0 go vet   ./...
CGO_ENABLED=0 go test  ./...

# 端到端冒烟（真实创建数据 + 关闭重开数据库验证恢复）
go run ./cmd/task225-timberjoint --smoke-test

# 启动服务
go run ./cmd/task225-timberjoint --addr :8080 --db ./task225-timberjoint.db
```

## API 入口（前缀 /api）

| 能力 | 入口 |
| --- | --- |
| 批次生命周期 | `POST /api/batches`、`GET /api/batches`、`GET /api/batches/{id}`、`POST /api/batches/{id}/advance`、`POST /api/batches/{id}/archive` |
| 测点 | `POST /api/batches/{id}/points`、`GET /api/batches/{id}/points`、`GET /api/points/{id}`、`PATCH /api/points/{id}/status` |
| 构件 | `POST /api/batches/{id}/members`、`GET /api/batches/{id}/members`、`GET /api/members/{id}`、`PATCH /api/members/{id}/status` |
| 节点关系与复核 | `POST /api/batches/{id}/joints`、`GET /api/batches/{id}/joints`、`GET /api/joints/{id}`、`POST /api/joints/{id}/check`、`POST /api/joints/{id}/confirm`、`POST /api/joints/{id}/reject` |
| 照片证据 | `POST /api/joints/{id}/evidences`、`GET /api/joints/{id}/evidences` |
| 节点版本 | `POST /api/batches/{id}/nodes/{nodeNo}/versions`、`GET /api/batches/{id}/nodes/{nodeNo}/versions`、`GET /api/versions/{id}` |
| 统计与健康 | `GET /api/stats`、`GET /api/health` |

## 持久化

SQLite（`modernc.org/sqlite`，纯 Go 驱动，CGO 无关）。批次、测点、构件、节点关系、
照片证据、节点版本六类实体全部落库；批次/测点/构件带幂等指纹唯一约束，版本带
`(batch, node, version_no)` 唯一约束。服务关闭重开同一数据库路径即可恢复全部数据。

## 目录结构

```
cmd/task225-timberjoint/main.go
internal/
  model/     实体、状态机、领域错误
  geometry/  三维向量、误差椭球、闭合、方向、编号
  survey/    批次与测点规则
  member/    构件与榫卯规则
  joint/     节点关系与几何复核
  evidence/  照片证据规则
  version/   节点版本规则
  store/     SQLite 持久化
  service/   编排层
  httpapi/   HTTP 层
```
