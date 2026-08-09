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
├── cmd/
│   ├── server/                 # HTTP 服务入口
│   ├── scheduler/              # cron 定时投递入口
│   ├── worker/                 # Asynq 任务消费入口
│   ├── migrate/                # 数据库迁移命令
│   └── tool/                   # 运维工具入口
├── config/                     # dev/test/prod 配置文件
├── docs/                       # Swagger 文档与设计记录
├── internal/
│   ├── app/                    # 各 runtime 的初始化、启动与关闭
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
├── migrations/                 # PostgreSQL 迁移脚本
├── scripts/                    # 模板模块名重命名脚本
├── docker-compose.yml          # PG + MinIO + Redis + 应用进程
├── Dockerfile                  # server/scheduler/worker 通用镜像构建
├── Makefile
└── go.mod
```

## 快速开始

### 1. 启动开发基础设施

```bash
make dev-up    # 启动 PostgreSQL + MinIO + Redis
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

请**按文件名顺序执行全部迁移**（`migrations/` 下脚本可能包含破坏性变更，例如用户表结构重建）；不要只执行部分迁移以免与当前模型不一致。

### 4. 启动服务

```bash
make run       # HTTP 服务 http://localhost:8080
```

定时任务可独立运行：

```bash
make run-scheduler
make run-worker
```

生产运行时职责固定为：HTTP 服务处理请求，scheduler 到点后向 Redis 投递任务，worker 执行任务。HTTP 服务不会启动 cron，因此同时部署三个进程不会重复注册定时任务。

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
| `JWT_SECRET` | jwt.secret |
| `TZ` | server.timezone |
| `SNOWFLAKE_NODE_ID` | snowflake.node_id |
| `TASK_QUEUE_ENABLED` | task_queue.enabled |
| `TASK_QUEUE_CONCURRENCY` | task_queue.concurrency |
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
| GET | `/ready` | - | 就绪检查 (DB、Redis ping) |
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
  "message": "success",
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

## License

[MIT](LICENSE)
