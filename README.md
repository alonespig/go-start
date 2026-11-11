# go-starter

一个通用的 Go 后端脚手架，拿过来即可直接开发。基于分层架构 + 依赖注入，支持多命令启动（HTTP 服务 / 定时任务 / 数据库迁移）。

## 技术栈

- Web 框架：[Gin](https://github.com/gin-gonic/gin)
- ORM / 数据库：[GORM](https://gorm.io) + MySQL
- 缓存：[go-redis v9](https://github.com/redis/go-redis)
- CLI / 配置：[Cobra](https://github.com/spf13/cobra) + [Viper](https://github.com/spf13/viper)
- 日志：[Zap](https://github.com/uber-go/zap)
- 定时任务：[robfig/cron](https://github.com/robfig/cron)
- 鉴权：JWT（golang-jwt）
- 可观测性：Prometheus + Zap
- 数据库迁移：goose（SQL 迁移内嵌进二进制）
- 消息队列：Kafka（kafka-go，默认关闭，基建已就绪）

## 目录结构

```
go-starter/
├── main.go                  # 入口：注册校验规则 → cmd.Execute()
├── config.example.yaml      # 配置模板（复制为 config.yaml 后使用）
├── config.yaml              # 本地配置，已被 gitignore，不提交
├── docs/                    # OpenAPI 文档
├── .golangci.yml            # lint 配置
├── cmd/                     # Cobra 命令层（每个命令一种进程角色）
│   ├── root.go
│   ├── serve.go             # HTTP 服务
│   ├── cron.go              # 定时任务
│   ├── migrate.go           # 数据库迁移
│   └── version.go
├── config/                  # 配置结构 + viper 加载
├── pkg/                     # 可复用公共库
│   ├── log/                 # 全局日志（zap 单例）
│   ├── errcode/             # 业务错误码 + BizError
│   ├── cache/               # 泛型缓存工具（cache-aside）
│   └── slice/               # 泛型集合工具（Map）
└── internal/
    ├── api/                 # HTTP 层（对外接口的一切）
    │   ├── router/          # 路由注册
    │   ├── handler/         # 绑定 dto → 调 service → 包 response
    │   ├── middleware/      # 中间件
    │   ├── dto/             # 请求/响应数据结构
    │   ├── validation/      # 参数校验错误格式化
    │   └── response/        # 统一响应信封
    ├── di/                  # 依赖注入 / 组合根
    ├── service/             # 业务逻辑层
    ├── repository/          # 数据访问层
    ├── infra/               # 基础设施层
    │   ├── db/              # GORM MySQL + goose 迁移
    │   ├── redis/           # Redis
    │   └── kafka/           # Kafka 生产/消费封装
    └── model/               # 数据模型
```

依赖方向（单向）：`cmd → di`，`api → service → repository → infra`；`config`、`pkg` 为公共依赖。

## 快速开始

1. 复制示例配置，并按需修改 MySQL / Redis 连接信息（`config.yaml` 已被 `.gitignore` 忽略，不会提交）：

```bash
cp config.example.yaml config.yaml
```

2. 执行数据库迁移：

```bash
go run . migrate
```

3. 启动 HTTP 服务：

```bash
go run . serve
```

4. 验证：

```bash
curl http://localhost:8080/health
```

## 命令

| 命令 | 说明 |
| --- | --- |
| `go run . serve` | 启动 HTTP 服务 |
| `go run . cron` | 启动定时任务调度器 |
| `go run . migrate` | 执行版本化数据库迁移 |
| `go run . migrate --status` | 查看迁移状态 |
| `go run . migrate --down` | 回滚最近一次迁移 |
| `go run . seed` | 写入开发用示例数据 |
| `go run . version` | 打印版本号 |

通用参数：`--config` 指定配置文件路径（默认 `config.yaml`）。

开发辅助命令（需要 `make`）：

| 命令 | 说明 |
| --- | --- |
| `make fmt` | 格式化代码 |
| `make vet` | 静态检查 |
| `make test` | 运行测试 |
| `make lint` | golangci-lint（需本地安装，版本要支持 Go 1.26） |
| `make build` | 编译到 `bin/go-starter` |

## 配置

- 配置模板见 `config.example.yaml`，复制为 `config.yaml` 后使用；`config.yaml` 不提交。
- 所有配置项都支持 `GO_STARTER_` 前缀的环境变量覆盖，例如 `GO_STARTER_MYSQL_PASSWORD`、`GO_STARTER_REDIS_ADDR`，优先级高于配置文件，适合容器和 CI 注入密码。
- `server` 段控制超时、请求头/请求体大小、可信代理与优雅退出；`cors` 段控制跨域来源；`auth` 段控制 JWT 鉴权（`jwt_secret` 为空表示不启用）；`metrics` / `rate_limit` / `idempotency` / `kafka` 分别控制指标、限流、幂等和消息队列；`log` 段控制级别、输出文件与轮转（`max_size` / `max_backups` / `max_age` / `compress`）。
- 启动时做基础校验（端口、MySQL / Redis 必填项、超时值等），配置错误直接退出，不会带病运行。

## 接口示例

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/health` | 存活检查 |
| GET | `/ready` | 就绪检查（MySQL / Redis） |
| POST | `/api/v1/users` | 创建用户（返回新 ID） |
| GET | `/api/v1/users/:id` | 查询用户（带 Redis 缓存） |
| PUT | `/api/v1/users/:id` | 更新用户（成功后失效缓存） |
| DELETE | `/api/v1/users/:id` | 删除用户（成功后失效缓存） |
| GET | `/api/v1/users?page=1&size=10` | 用户列表（分页） |
| GET | `/api/v1/me` | 当前 token 身份（启用鉴权后注册） |
| GET | `/metrics` | Prometheus 指标 |

统一响应格式：

```json
{ "code": 0, "message": "ok", "data": {} }
```

## 新增一个模块

以「新增一个实体」为例，按依赖方向从内向外加：

1. `model/` 加数据模型（GORM tag）。
2. `repository/` 加接口 + 实现（`NewXxxRepository`）；通用 CRUD 直接嵌入 `BaseRepo[T]`，只有特殊查询才自己写。
3. `service/` 加接口 + 实现（`NewXxxService`）；缓存复用泛型 `cache.GetOrSet`。
4. `api/dto/` 加请求/响应 DTO（`ToModel` / `ToXxxResp`）。
5. `api/handler/` 加 handler（`NewXxxHandler`）。
6. `api/router/` 注册路由。
7. `di/di.go` 的 `Build` 里按顺序装配。
8. `internal/infra/db/migrations/` 加版本化 SQL 迁移（`goose`），本地跑 `make migrate`。

> 新领域建议按模块切：`internal/modules/<domain>/{handler,service,repository,dto,model}.go`，
> 避免所有业务继续堆在全局 layer；现有 user 模块保持分层结构即可，后续可逐步迁移。

## 注意事项

- 参数校验用 `pkg/validation.FormatError` + `response.FailWithData`，错误明细放在 `data` 字段；不要直接把 validator 的英文原文返回给前端。
- 每个请求会生成或透传 `X-Request-ID`，日志里带 `request_id`，排查线上问题时可按它串联一次请求。
- 错误码统一放 `pkg/errcode`，业务层返回 `BizError`，handler 层用 `response.Fail` 翻译成响应。不要把业务错误码放 `api` 层，避免 service 反向依赖 api。
- HTTP 超时在配置的 `server` 段设置（默认 `read_header_timeout: 10s`、`read_timeout: 30s`、`write_timeout: 30s`、`idle_timeout: 120s`）。上 SSE 前把 `write_timeout` 调成 `0`，或在 handler 里用 `http.NewResponseController(c.Writer).SetWriteDeadline(time.Time{})` 单独取消；WebSocket 建议 `read_timeout`、`write_timeout` 都设 `0`（或在升级拿到连接后自行清除 deadline）。
- 优雅退出最多等 `server.shutdown_timeout`（默认 10s）：SSE 是活跃请求，`Shutdown` 会一直等到超时，需要在退出时主动通知 SSE handler 关闭；WebSocket 连接被 hijack 后不受 `Shutdown` 管理，需要自己记录并在退出时关闭。
- Redis 目前定位为缓存（cache-aside 模式），需要分布式锁/计数器时再单独封装。
- 鉴权：`internal/auth.TokenService` 负责签发/校验 JWT，`middleware.Auth` 负责解析 `Authorization: Bearer`，claims 会放进 gin.Context 和 request context。接入登录接口时复用 `TokenService.Issue` 即可。
- 可观测性：`/metrics` 暴露 HTTP 请求量/耗时/在途数与数据库连接池指标；日志带 `request_id`，可直接接到 Prometheus + Grafana / Loki。
- 安全：统一安全响应头、请求体大小限制、可选 Redis 限流；写接口带 `Idempotency-Key` 时启用幂等重放，避免重复提交。
- 模块路径为 `github.com/alonespig/go-start`，迁移到其他仓库时同步修改 `go.mod` 与 import 路径即可。
- 接口文档见 `docs/openapi.yaml`；改动接口时同步更新，便于前端联调和生成 SDK。
- `make lint` 依赖 golangci-lint v2+ 且其构建 Go 版本不低于 1.26；CI 已配置 lint job。

## Docker 部署

1. 构建并启动全部服务（MySQL / Redis / 应用）：

```bash
docker compose up -d
```

2. 首次启动先执行数据库迁移（复用同一个镜像跑 `migrate` 命令）：

```bash
docker compose run --rm app migrate
```

3. 验证：

```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

说明：容器内使用 `config.docker.yaml`（`docker-compose.yml` 已将其挂载为 `/app/config.yaml`），数据库地址指向 compose 服务名（`mysql` / `redis`）。本机直接 `go run . serve` 时用根目录的 `config.yaml`（地址是 `127.0.0.1`），两者互不影响。
