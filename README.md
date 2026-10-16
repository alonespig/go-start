# go-starter

一个通用的 Go 后端脚手架，拿过来即可直接开发。基于分层架构 + 依赖注入，支持多命令启动（HTTP 服务 / 定时任务 / 数据库迁移）。

## 技术栈

- Web 框架：[Gin](https://github.com/gin-gonic/gin)
- ORM / 数据库：[GORM](https://gorm.io) + MySQL
- 缓存：[go-redis v9](https://github.com/redis/go-redis)
- 文档数据库：[mongo-driver](https://github.com/mongodb/mongo-go-driver)
- CLI / 配置：[Cobra](https://github.com/spf13/cobra) + [Viper](https://github.com/spf13/viper)
- 日志：[Zap](https://github.com/uber-go/zap)
- 定时任务：[robfig/cron](https://github.com/robfig/cron)

## 目录结构

```
go-starter/
├── main.go                  # 入口，仅调用 cmd.Execute()
├── config.yaml              # 配置文件
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
    │   └── response/        # 统一响应信封
    ├── di/                  # 依赖注入 / 组合根
    ├── service/             # 业务逻辑层
    ├── repository/          # 数据访问层
    ├── infra/               # 基础设施层
    │   ├── db/              # GORM MySQL
    │   ├── redis/           # Redis
    │   └── mongo/           # MongoDB
    └── model/               # 数据模型
```

依赖方向（单向）：`cmd → config → di → infra → repository → service → api`

## 快速开始

1. 修改根目录 `config.yaml` 里的 MySQL / Redis / Mongo 连接信息。
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
| `go run . migrate` | 执行数据库迁移 |
| `go run . version` | 打印版本号 |

通用参数：`--config` 指定配置文件路径（默认 `config.yaml`）。

## 接口示例

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/health` | 健康检查 |
| POST | `/api/users` | 创建用户 |
| GET | `/api/users/:id` | 查询用户（带 Redis 缓存） |
| GET | `/api/users?page=1&size=10` | 用户列表（分页） |

统一响应格式：

```json
{ "code": 0, "message": "ok", "data": {} }
```

## 新增一个模块

以「新增一个实体」为例，按依赖方向从内向外加：

1. `model/` 加数据模型（GORM / bson tag）。
2. `repository/` 加接口 + 实现（`NewXxxRepository`）；通用 CRUD 直接嵌入 `BaseRepo[T]`，只有特殊查询才自己写。
3. `service/` 加接口 + 实现（`NewXxxService`）；缓存复用泛型 `cache.GetOrSet`。
4. `api/dto/` 加请求/响应 DTO（`ToModel` / `ToXxxResp`）。
5. `api/handler/` 加 handler（`NewXxxHandler`）。
6. `api/router/` 注册路由。
7. `di/di.go` 的 `Build` 里按顺序装配。
8. `cmd/migrate.go` 里把新 model 加入 `AutoMigrate`。

## 注意事项

- 错误码统一放 `pkg/errcode`，业务层返回 `BizError`，handler 层用 `response.Fail` 翻译成响应。不要把业务错误码放 `api` 层，避免 service 反向依赖 api。
- Redis 目前定位为缓存（cache-aside 模式），需要分布式锁/计数器时再单独封装。
- 模块名当前为 `go-starter`，如需发布可改成 `github.com/yourname/go-starter` 并全局替换 import 路径。

## Docker 部署

1. 构建并启动全部服务（MySQL / Redis / Mongo / 应用）：

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
```

说明：容器内使用 `config.docker.yaml`（`docker-compose.yml` 已将其挂载为 `/app/config.yaml`），数据库地址指向 compose 服务名（`mysql` / `redis` / `mongo`）。本机直接 `go run . serve` 时用根目录的 `config.yaml`（地址是 `127.0.0.1`），两者互不影响。
