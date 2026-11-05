# Distributed-Seckill

一个面向高并发限时抢购场景的分布式秒杀系统学习项目：Go + Gin + MySQL + Redis，后续按里程碑接入网关、Kafka 异步削峰、限流与可观测性。

## 当前进度

- 里程碑 A（同步直连全链路）进行中：用户注册/登录、商品/活动主数据、商品详情缓存已完成。
- W2 的秒杀接口、订单同步直建、DB 防超卖、一人一单尚未开始。
- 2026-10-30 代码审查发现的存量问题见 [`docs/PROGRESS.md`](docs/PROGRESS.md) 的「代码审查待修复」。

## 文档

| 文件 | 作用 |
|---|---|
| [`高并发分布式秒杀系统-项目计划.md`](高并发分布式秒杀系统-项目计划.md) | 唯一权威执行计划 |
| [`harness_project.md`](harness_project.md) | AI 协作守则与代码规范 |
| [`docs/PROGRESS.md`](docs/PROGRESS.md) | 当前里程碑、任务日志、待修复项 |
| [`docs/DECISIONS.md`](docs/DECISIONS.md) | 架构决策记录（ADR） |
| [`docs/GIT_GUIDE.md`](docs/GIT_GUIDE.md) | Git 日常操作与提交规范 |

## 本地环境

前置：Go 1.25、Docker / Docker Compose。

```bash
# 1. 安装钉死版本的工具链
make tools

# 2. 准备本地配置（deploy/.env 不入库，需自行填写密钥/密码；已存在则不要覆盖）
cp -n deploy/.env.example deploy/.env

# 3. 启动 MySQL 8（宿主 3307）与 Redis 7（宿主 6379）
make up

# 4. 执行迁移
make migrate
```

MySQL 配置：默认由 `MYSQL_HOST/PORT/DATABASE/USER/PASSWORD` 组合，密码缺省回退 `MYSQL_ROOT_PASSWORD`；也可用 `MYSQL_DSN` 整体覆盖（go-sql-driver 格式，不带 `mysql://`，`make migrate` 会自动补前缀）。连接/读写超时可用 `MYSQL_DIAL_TIMEOUT`（默认 3s）、`MYSQL_READ_TIMEOUT` / `MYSQL_WRITE_TIMEOUT`（默认 5s）调整。

常用检查：

```bash
make lint    # gofmt + goimports + go vet + golangci-lint
make test    # go test ./...
make build   # go build ./...
```

服务启动目前需手动执行（`make run` / `make seed` 尚未实现）：

```bash
# user 服务，计划端口 :8081
HTTP_ADDR=:8081 go run ./services/user/cmd/server

# product 服务，计划端口 :8082
HTTP_ADDR=:8082 go run ./services/product/cmd/server
```

> 服务默认读取仓库根目录的 `deploy/.env`，也可用 `ENV_FILE` 指定其他路径；已存在的进程环境变量优先。
