基于 Go 实现的古建筑木构节点测绘复核 Web 项目，一款后端服务，完成几何闭合校验、测绘证据复核与节点版本发布。

# BENZHI 评测说明 — task225-timberjoint

古建筑木构节点测绘复核台：对木构节点现场测绘数据做几何闭合/方向/编号复核，
登记照片证据并发布冻结的节点版本。

## 运行契约

- **入口**：`cmd/task225-timberjoint/main.go`
- **标志**：
  - `--addr :8080` 监听地址
  - `--db ./task225-timberjoint.db` SQLite 路径
  - `--smoke-test` 执行端到端冒烟并退出（退出码 0 = 通过）
- **冒烟判定**：`go run ./cmd/task225-timberjoint --smoke-test` 必须输出 `SMOKE TEST PASSED` 且退出码 0。

## 构建门禁

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
```

## Docker 双架构

```bash
bash build_benzhi_docker.sh my-project linux/amd64
bash build_benzhi_docker.sh my-project linux/arm64

docker run --rm my-project --smoke-test          # 冒烟
docker run --rm -p 8080:8080 my-project --addr :8080  # 服务
```

镜像 ENTRYPOINT 为二进制，CMD 为 `["--smoke-test"]`；运行冒烟时**不要**追加
路径参数，否则会被当作位置参数导致 `--smoke-test` 不生效。

## 关键 API

- 几何复核：`POST /api/joints/{id}/check` 返回闭合残差、方向散布、编号一致性。
- 版本发布：`POST /api/batches/{id}/nodes/{nodeNo}/versions`（body: `{"reason":"..."}`）。
- 统计：`GET /api/stats`；健康：`GET /api/health`。

## 组件版本

- Go 1.26.3
- SQLite 3.46.1（`modernc.org/sqlite v1.52.0`）
