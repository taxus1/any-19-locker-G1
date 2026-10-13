# go-gin-m · 基座（Go 1.22 / Gin + GORM + MySQL + Redis）

> **M 档 · Go 技术栈**。镜像 `adminfather/benzhi-claude-code:go`（Go 1.22.10，见
> `runtime/docker/Dockerfile.benzhi-go`）。验收用例用 **Go 原生 `go test`**（`internal/acceptance`）。

## 一键起环境
```bash
docker compose up -d          # MySQL 8.2 + Redis 7.2（宿主机执行）
```

## 构建 / 跑验收
```bash
go build ./...                # 编译
go test ./...                 # 跑全部测试（验收在 internal/acceptance）
```
> 容器/宿主两侧都读环境变量：`SERVER_PORT` / `DB_DSN` / `REDIS_ADDR` / `REDIS_PASSWORD`。
> 不要在代码里写死库名或地址。

## 目录（DDD-lite）
```
main.go                    # 入口：读 config → 起 Gin
internal/config            # 环境变量配置 + DB/Redis 连接
internal/common            # Result 统一返回
internal/model             # 领域对象 + GORM 钩子（审计/操作人 context）
internal/repository        # GORM 数据访问（只认 model）
internal/service           # 应用服务（领域不变量、编排）
internal/handler           # Gin handler（协议适配）
internal/middleware        # OperatorAuth（Basic 登录 → context）
internal/router            # 路由装配（验收用 gin.Routes() 扫）
internal/acceptance        # 验收基座（contracts.md 的 S1 契约按此探测）
doc/schema/demo.sql        # 底座示例表
base.conf / contracts.md / docker-compose.yml
```

## 与 solo 流水线
- 建题：`bin/solo-new <代号> --base go-gin-m`（基座字段固化进工作区 `.solo-base`，`IMAGE` 随之生效）。
- 验收：Go 用例放题库 `cases/<迭代>/`（`package acceptance`），共享基座放 `cases/common/`；
  `host_verify.py` 的 Go 分支编译/起产物、跑 `go test`、按用例记通过。

## 依赖缓存（免每次下载，2026-10-08）
- 宿主持久化缓存放 **`solo/go-cache/`**：`mod/`（GOMODCACHE）+ `build/`（GOCACHE），已预热 gin/gorm/go-redis/mysql 等。
- 做题容器启动时由 `runtime/claude-docker.sh` 自动挂进 `/home/node/go/pkg/mod` 与 `/home/node/.cache/go-build`
  （`host_verify.py` / `product_run.sh` 的 Go 分支同用这份缓存）。**离线**（`GOPROXY=off`）即可 `go build`。
- 关掉挂载：`GO_CACHE_DIR=off`；换路径：`GO_CACHE_DIR=<dir>`。
- 重新预热（基座依赖变了时）：用 `:go` 镜像 `go mod download all` 写进 `solo/go-cache/mod`。
- `go-cache/` 已在 `.gitignore`（勿入库）。
