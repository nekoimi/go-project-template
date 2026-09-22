# Go Template

基于 Go 的后端项目模板，集成常用组件，开箱即用。

## 技术栈

| 组件 | 说明 |
|---|---|
| [Gin](https://github.com/gin-gonic/gin) | HTTP 框架 |
| [GORM](https://gorm.io/) | ORM |
| [PostgreSQL](https://www.postgresql.org/) | 数据库 |
| [Zap](https://github.com/uber-go/zap) | 结构化日志 |
| [Viper](https://github.com/spf13/viper) | 配置管理 |
| [golang-jwt](https://github.com/golang-jwt/jwt) | JWT 认证 |
| [gorilla/websocket](https://github.com/gorilla/websocket) | WebSocket |
| [robfig/cron](https://github.com/robfig/cron) | 定时任务 |
| [Asynq](https://github.com/hibiken/asynq) | Redis 后台任务队列 |
| [Redis](https://redis.io/) | 任务队列存储 |
| [snowflake](https://github.com/bwmarrin/snowflake) | 分布式 ID |
| [gin-swagger](https://github.com/swaggo/gin-swagger) | API 文档 |
| [MinIO](https://min.io/) | 对象存储 (可选) |

## 项目结构

```
├── .github/
│   ├── workflows/              # lint、单测、竞态、集成测试与镜像发布
│   └── dependabot.yml          # Go 与 GitHub Actions 依赖更新
├── cmd/
│   ├── server/                 # HTTP 服务入口
│   ├── scheduler/              # cron 定时投递入口
│   ├── worker/                 # Asynq 任务消费入口
│   ├── all/                    # HTTP + scheduler + worker 一体化入口
│   ├── migrate/                # 数据库迁移命令
│   ├── tool/                   # 运维工具入口
│   └── version/                # 构建版本信息
├── config/                     # dev/test/prod 配置文件
├── docs/                       # Swagger 文档与设计记录
├── integration/                # PostgreSQL、Redis 集成测试
├── internal/
│   ├── app/                    # 各 runtime 的初始化、启动与关闭
│   ├── buildinfo/              # 版本、commit、构建时间
│   ├── config/                 # 配置模型、加载、校验与默认值
│   ├── domain/user/            # 用户领域实体
│   ├── framework/              # 模块注册、scope、生命周期、事件与健康检查
│   ├── middleware/             # CORS、JWT、限流、日志、Recovery、RequestID
│   ├── modules/                # 按业务能力组织的模块
│   │   ├── auth/               # 注册、登录与认证事件
│   │   ├── user/               # 用户资料
│   │   ├── upload/             # 文件上传
│   │   ├── websocket/          # WebSocket 模块注册
│   │   └── examplejob/         # cron 投递与 Asynq Handler 示例
│   ├── pkg/                    # 数据库、日志、错误码、响应、ID、JWT、时间工具
│   ├── repository/             # GORM 数据访问层
│   ├── scheduler/              # robfig/cron 调度器封装
│   ├── storage/                # Local / S3 文件存储及工厂
│   ├── taskqueue/              # Asynq Client、Worker 与通用任务协议
│   ├── transport/              # 对外传输层
│   │   ├── http/               # Gin 路由与 HTTP 基础端点
│   │   └── grpc/               # gRPC transport 扩展位置
│   └── websocket/              # WebSocket 连接与消息管理
├── migrations/                 # PostgreSQL 初始基线及后续迁移
├── scripts/                    # Bash/PowerShell 模块路径重命名脚本
├── .dockerignore               # Docker 构建上下文排除规则
├── .gitattributes              # Shell 脚本 LF 换行规则
├── docker-compose.yml          # PG + MinIO + Redis + 应用进程
├── Dockerfile                  # 非 root、多命令通用镜像
├── Makefile                    # 本地开发、构建和部署快捷命令
└── go.mod                      # Go 模块与依赖
```

## 使用模板

创建新项目后，先修改 Go module path。脚本会同步更新 `go.mod`、Go 源码 import
以及 Dockerfile 中用于注入构建信息的包路径。

Linux/macOS：

```bash
./scripts/rename-module.sh github.com/your-org/your-project
```

Windows PowerShell：

```powershell
.\scripts\rename-module.ps1 github.com/your-org/your-project
```

## 快速开始

### 1. 启动开发基础设施

```bash
make dev-up    # 仅启动 PostgreSQL + MinIO + Redis
```

### 2. 运行数据库迁移

```bash
make migrate-up
```

也可以直接使用内置迁移命令：

```bash
go run ./cmd/migrate --config config/config.dev.yaml up
go run ./cmd/migrate --config config/config.dev.yaml version
go run ./cmd/migrate --config config/config.dev.yaml down 1
```

支持 `up`、`down`、`version`、`goto` 和 `force`。也可以通过
`MIGRATE_DATABASE_URL` 或 `--database-url` 覆盖配置文件中的数据库连接。

### 3. 使用运维工具创建用户

模板当前没有角色/RBAC 字段，因此工具提供通用用户创建命令：

```bash
go run ./cmd/tool user create \
  --config config/config.dev.yaml \
  --username admin \
  --email admin@example.com \
  --password 'change-me-now'
```

也支持环境变量 `APP_CONFIG`、`APP_USER_USERNAME`、`APP_USER_EMAIL` 和
`APP_USER_PASSWORD`。后续引入角色模型后，可在此命令基础上增加管理员提升命令。

当前模板从一份全新的 BIGINT 用户表基线迁移开始。复制模板创建新项目后，后续结构
变更应继续追加新的迁移文件，不要修改已经在环境中执行过的迁移。

### 4. 启动服务

```bash
make run       # HTTP 服务 http://localhost:8080
```

定时任务可独立运行：

```bash
make run-scheduler
make run-worker
```

也可以在一个进程中启动全部运行时：

```bash
make run-all
```

项目提供三种部署方式：`server` 仅运行 HTTP；`server`、`scheduler`、`worker`
可拆分为独立进程；`all` 在一个进程中同时运行三者。拆分部署适合独立扩缩容，
一体化模式适合本地开发和小规模部署。scheduler 默认通过 Redis 进行 leader
election，因此多个 `all` 或 scheduler 实例中只有 leader 会运行 cron；仍不建议
在同一部署中混用 `all` 与独立 scheduler，以免增加不必要的运行复杂度。
`scheduler.leader_ttl` 应大于 cron 回调本身的最长执行时间；cron 回调应只负责
快速投递队列任务，耗时工作交给 worker。

HTTP-only 模式不会创建任务队列客户端，也不会把 Redis 纳入 `/ready` 检查。
`scheduler` 到点后向 Redis 投递任务，`worker` 负责消费任务。

Asynq 使用至少一次投递语义，任务处理器必须幂等。Payload 只应包含 ID 和小型参数；文件、长文本和模型上下文应存入 PostgreSQL 或对象存储，任务中只传引用。默认队列为 `critical`、`default` 和 `ai`，可以分别配置优先级与 worker 总并发数。

## 配置

配置文件位于 `config/`，通过 `--config` 参数指定。支持环境变量覆盖：

| 环境变量 | 对应配置 |
|---|---|
| `DATABASE_HOST` | database.host |
| `DATABASE_PORT` | database.port |
| `DATABASE_USER` | database.user |
| `DATABASE_PASSWORD` | database.password |
| `DATABASE_NAME` | database.dbname |
| `DATABASE_CONNECT_TIMEOUT` | database.connect_timeout |
| `JWT_SECRET` | jwt.secret |
| `TZ` | server.timezone |
| `SERVER_READ_TIMEOUT` | server.read_timeout |
| `SERVER_READ_HEADER_TIMEOUT` | server.read_header_timeout |
| `SERVER_WRITE_TIMEOUT` | server.write_timeout |
| `SERVER_IDLE_TIMEOUT` | server.idle_timeout |
| `SNOWFLAKE_NODE_ID` | snowflake.node_id |
| `TASK_QUEUE_ENABLED` | task_queue.enabled |
| `TASK_QUEUE_CONCURRENCY` | task_queue.concurrency |
| `SCHEDULER_LEADER_ELECTION` | scheduler.leader_election |
| `SCHEDULER_LEADER_KEY` | scheduler.leader_key |
| `SCHEDULER_LEADER_TTL` | scheduler.leader_ttl |
| `REDIS_ADDR` | task_queue.redis.addr |
| `REDIS_PASSWORD` | task_queue.redis.password |
| `REDIS_DB` | task_queue.redis.db |
| `S3_ACCESS_KEY` | storage.s3.access_key |
| `S3_SECRET_KEY` | storage.s3.secret_key |
| `S3_ENDPOINT` | storage.s3.endpoint |
| `S3_PUBLIC_URL` | storage.s3.public_url |
| `S3_BUCKET` | storage.s3.bucket |
| `S3_REGION` | storage.s3.region |
| `S3_USE_SSL` | storage.s3.use_ssl |
| `S3_FORCE_PATH_STYLE` | storage.s3.force_path_style |
| `S3_CREATE_BUCKET` | storage.s3.create_bucket |
| `S3_STARTUP_TIMEOUT` | storage.s3.startup_timeout |

生产使用的 `config/config.prod.yaml` 中等占位符（如 `${S3_ACCESS_KEY}`）不会被自动展开，需通过上表环境变量覆盖，或在 YAML 中直接写最终值。为平滑迁移，`MINIO_ACCESS_KEY`、`MINIO_SECRET_KEY`、`MINIO_ENDPOINT`、`MINIO_PUBLIC_URL`、`MINIO_BUCKET` 仍可作为对应 `S3_*` 环境变量的兼容别名。

多实例部署时，请为每个实例设置不同的 `SNOWFLAKE_NODE_ID`，避免雪花 ID 冲突。

数据库主键仍为 `bigint`，但 **JSON API 中的用户 ID 一律为十进制字符串**（DTO 字段类型为 `string`），避免 JavaScript `Number` 对大整数精度丢失；前端请按字符串传递与展示，不要 `parseInt` / `Number()` 后再回传。

设置 `websocket.enabled: true` 后才会注册 `/ws/v1/chat` 并启动 WebSocket 管理循环；同一配置块中的 buffer、读写超时、`max_message_size`、ping 间隔会应用于连接。

当配置了 `server.allowed_origins` 时，**未携带 `Origin` 头的请求**（如 curl / 服务端调用）不会因 CORS 白名单被拦成 403；浏览器跨站请求仍会按白名单校验。

完整配置项见 `config/config.dev.yaml`。

## API

启动后访问 Swagger UI：`http://localhost:8080/swagger/index.html`

| 方法 | 路径 | 认证 | 说明 |
|---|---|---|---|
| GET | `/health` | - | 存活检查 |
| GET | `/ready` | - | 就绪检查（仅检查当前运行模式实际使用的依赖） |
| POST | `/v1/auth/register` | - | 用户注册 |
| POST | `/v1/auth/login` | - | 用户登录 |
| GET | `/v1/users/profile` | JWT | 获取当前用户信息 |
| POST | `/v1/upload/single` | JWT | 上传单个文件 |
| POST | `/v1/upload/multiple` | JWT | 上传多个文件 |
| GET | `/ws/v1/chat` | JWT | WebSocket；推荐：`Sec-WebSocket-Protocol: access_token, <jwt>`（与 `new WebSocket(url, ['access_token', token])` 一致）；兼容查询参数 `?token=` |

### 统一响应格式

```json
{
  "code": 0,
  "msg": "success",
  "data": {},
  "error": null
}
```

错误时 `code` 为业务错误码（如 40100=未授权，40401=用户不存在），`error` 包含错误详情。

## 构建与部署

```bash
make build         # 编译到 bin/
make swagger       # 重新生成 Swagger 文档
make test          # 运行测试
make lint          # golangci-lint（需已安装：go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest）

make docker-build  # 构建 Docker 镜像
make docker-up     # 启动完整部署 (app + scheduler + worker + PG + MinIO + Redis)
make docker-down   # 停止
```

不带 `full` profile 时，Compose 只启动 PostgreSQL、MinIO 和 Redis，适合本地运行
Go 进程；`make docker-up` 会启用 `full` profile，同时启动 app、scheduler 和
worker。应用容器提供 `/health` 健康检查，镜像使用非 root 用户运行。

Docker 容器支持在 `server` 或 `all` 启动前自动执行 `migrate up`，默认关闭；
`scheduler`、`worker`、`migrate` 和 `tool` 不会触发自动迁移。使用
Compose 时可通过环境变量开启：

```bash
AUTO_MIGRATE=true docker compose --profile full up -d
```

也可以在部署平台中为 `app` 容器设置 `AUTO_MIGRATE=true`。迁移失败时容器会
直接退出，不会继续启动业务进程。`MIGRATE_CONFIG` 和 `MIGRATE_PATH` 可分别覆盖
迁移使用的配置文件（默认 `config/config.prod.yaml`）和 SQL 目录（默认
`migrations`）；数据库连接仍可通过 `MIGRATE_DATABASE_URL` 覆盖。多副本部署时，
建议只为一个启动实例开启自动迁移，或在发布流程中使用独立迁移任务。

同一个镜像包含 `server`、`scheduler`、`worker`、`all`、`migrate`、`tool` 和
`version` 七个命令，默认执行 `server`。例如：

```bash
docker compose --profile full run --rm app migrate --config config/config.prod.yaml version
docker run --rm go-project-template:local tool --help
docker run --rm go-project-template:local all --config config/config.prod.yaml
docker run --rm go-project-template:local version
docker compose --profile full run --rm worker
```

通过 Compose 启动时，`app`、`scheduler` 和 `worker` 也会复用
`go-project-template:local` 这一个镜像。

发布流水线会通过 ldflags 注入版本、commit 和构建时间，可用 `version` 查看：

```text
version=v1.2.3 commit=abc123 build_time=2026-09-22T12:00:00Z
```

## 测试与 CI

本地单元测试和静态检查：

```bash
go test ./...
go test -race ./...
go vet ./...
```

集成测试需要 PostgreSQL 和 Redis。首次执行时创建独立测试数据库：

```bash
make dev-up
docker compose exec postgres createdb -U postgres go_template_test
go test -tags=integration -count=1 ./integration
```

默认连接 `localhost:5432/go_template_test` 和 `localhost:6379`，可分别通过
`INTEGRATION_DATABASE_URL`、`INTEGRATION_REDIS_ADDR` 覆盖。CI 会自动启动依赖，
验证 migration、数据库唯一约束、任务投递/消费/关闭、竞态检测、Docker 构建及
镜像内各命令的 smoke test。

## License

[MIT](LICENSE)
