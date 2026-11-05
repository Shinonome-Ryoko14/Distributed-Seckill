# 项目进度

> 规则：每完成一个任务必须更新本文件。AI 每次开工**先读本页顶部「当前状态」**，再按 `harness_project.md` §17 的权重规则读「任务日志」。

## 当前状态（最高读取权重 · 开工锚点）

- **当前里程碑**: A — 同步直连全链路（每域一服务，W1–W3）；W1 基础接口与 W2 商品缓存子项已完成，W2 秒杀/订单部分尚未开始
- **当前任务**: MySQL 配置/DSN 统一改造完成（ADR-0014）；`make lint && make test && make build` 全绿
- **上次会话结束点**: 2026-11-05 完成 MySQL 配置/DSN 统一改造（Makefile、`internal/config`、两个 `main.go`、单测），已通过 lint/test/build，待提交
- **下一步**: 处理 Redis `MaxRetries`/快速失败；随后做 `POST /products`、`POST /activities` 鉴权，最后进入 W2 秒杀接口 + 订单同步直建 + DB 条件更新防超卖
- **范围基线**: 里程碑 C 采用「最小可信集」（ADR-0011）；A/B 做透，目标 **2026-11 底–12 月中收口**，2027-01–06 做第二个 Agent 项目
- **阻塞**: 无硬阻塞；有代码审查待修复项（见下）

## 影响分级图例

| 级别 | 含义 |
|---|---|
| `L` 大 | 跨模块/跨服务、数据模型或行为变更、架构调整 |
| `M` 中 | 单模块内行为变更 |
| `S` 小 | 文档/格式/无行为变更 |

## 里程碑

- [ ] **A 同步直连全链路（每域一服务，W1–W3）**：数据模型、用户/商品/活动、DB 防超卖、一人一单、超时取消+回补、v0 基线压测
  - 已完成：W1 用户/商品/活动基础接口；W2 商品详情缓存（ADR-0012）
  - 待完成：W2 秒杀接口/订单同步直建/DB 条件更新防超卖/一人一单；W3 超时取消+回补、v0 基线压测
- [ ] **B 网关+异步削峰（W4–W6）**：gateway、Lua 原子门+预热、Kafka 异步建单、幂等/DLQ、限流
- [ ] **C 工程化+压测（W7–W8）**：docker-compose、Prometheus/Grafana、CI、k6 优化、故障注入、报告
  - 交付范围按 **ADR-0011 最小可信集**收口（Grafana 多面板/多轮调优/面试题清单后置）

## 任务日志

| 日期 | 里程碑 | 任务 | 状态 | 影响 | 涉及文件 |
|---|---|---|---|---|---|
| 2026-09-13 | A | 制定工程规范并写入 harness §10–§16 | 完成 | L | `harness_project.md` |
| 2026-09-13 | A | 建立 docs/DECISIONS.md、docs/PROGRESS.md（9 条 ADR） | 完成 | M | `docs/DECISIONS.md`、`docs/PROGRESS.md` |
| 2026-09-13 | A | 制定每日开工/收工流程 | 完成 | M | `harness_project.md`、`docs/DAILY_KICKOFF.md`、`docs/PROGRESS.md` |
| 2026-09-14 | A | 验证 docker-compose 环境；修复 Redis 健康检查（`PING`→`redis-cli ping`） | 完成 | M | `deploy/docker-compose.yml`、`deploy/.env` |
| 2026-09-14 | A | 新增 harness §19.1 骨架填空式协作约定（hint 分级/TODO 可验收/讲解后置） | 完成 | S | `harness_project.md` |
| 2026-09-15 | A | 新增 `Makefile`（tools/up/down/migrate/lint/test/build/fmt 等）与 `.golangci.yml`（v2 配置，§11 linter 集） | 完成 | M | `Makefile`、`.golangci.yml` |
| 2026-09-15 | A | `make tools` 安装钉死版本工具链；`make lint/test/build` 通过（无需 Go 包时跳过） | 完成 | S | `Makefile` |
| 2026-09-15 | A | 给出 `migrations/0001_init.{up,down}.sql` 骨架与任务卡（DDL 待用户填写） | 完成 | L | `migrations/0001_init.up.sql`、`migrations/0001_init.down.sql` |
| 2026-09-17 | A | 用户完成 0001 三表 DDL；`make migrate` / `migrate-down` / 再 `migrate` 验收通过（唯一索引、外键 RESTRICT、中文 COMMENT 编码正确） | 完成 | L | `migrations/0001_init.{up,down}.sql` |
| 2026-09-17 | A | 修复 Makefile `migrate` 密码回显、新增 `migrate-force`；修复迁移注释 MySQL `--` 需空格语法 | 完成 | S | `Makefile`、`migrations/0001_init.up.sql` |
| 2026-09-17 | A | 落地共享基础设施 `internal/{config,log,db,errs,response,jwt,ctxkey}`；config/jwt 单测；引入 `jwt/v5`、`go-sql-driver/mysql`；新增 ADR-0010（响应契约/错误码/JWT） | 完成 | L | `internal/*`、`go.mod`、`deploy/.env(.example)`、`docs/DECISIONS.md` |
| 2026-09-17 | A | 确定工期与范围基线：里程碑 C 收敛为最小可信集（ADR-0011），目标 2026-11 底–12 月中收口 | 完成 | L | `docs/DECISIONS.md`、`docs/PROGRESS.md` |
| 2026-09-21 | A | 回滚 Go 语法练习与所有权补课痕迹；恢复 harness §19 教学模式；补充 `internal/errs` 单测 | 完成 | M | `internal/gopatch/*`（删）、`docs/OWNERSHIP.md`（删）、`harness_project.md`、`internal/errs/errs_test.go`、`docs/PROGRESS.md` |
| - | A | `deploy/docker-compose.yml`（mysql:3307 + redis:6379） | 完成 | L | `deploy/*` |
| 2026-10-04 | A | 用户注册/登录 + JWT：model/repo/service + Gin handler/router/main 装配、handler/service 单测、E2E curl 冒烟 | 完成 | L | `services/user/**`、`go.mod`、`go.sum`、`docs/PROGRESS.md` |
| 2026-10-10 | A | product 商品主数据接口：model/repo/service + Gin router/main、分页、商品域错误码 20001/20002、handler/service 单测、E2E 冒烟 | 完成 | L | `services/product/**`、`internal/errs/errs.go`、`docs/PROGRESS.md` |
| 2026-10-17 | A | product 活动主数据接口：POST/GET activities、活动域错误码 20003/20004、RFC3339 时间解析、先查 product 再校验、handler/service 单测、补 13 处导出注释 | 完成 | L | `services/product/**`、`internal/errs/errs.go`、`docs/PROGRESS.md` |
| 2026-10-20–26 | A | Redis 客户端 + product 商品详情缓存：Cache Aside、负缓存、Redis 故障降级、TTL 配置、单测与 main 装配 | 完成 | L | `internal/rediscli/**`、`internal/config/**`、`services/product/**`、`go.mod`、`go.sum`、`deploy/.env.example` |
| 2026-10-30 | A | 代码审查（缓存更新范围外）：记录活动校验/错误码/配置/Redis 重试等待修复项，未改业务代码 | 完成 | S | `docs/PROGRESS.md` |
| 2026-10-30 | A | 文档收口：补 ADR-0012（缓存）与 ADR-0013（库存写入边界）、更新 PROGRESS 状态/日志、补 README 最小启动与文档索引、补服务端口注释 | 完成 | M | `docs/PROGRESS.md`、`docs/DECISIONS.md`、`README.md`、`harness_project.md`、`deploy/.env.example` |
| 2026-11-02 | A | 活动创建校验：商品必须上架、`status` 与 `start/end` 时间自洽（pending/active/completed 边界），注入时钟依赖并补 service 单测 | 完成 | M | `services/product/internal/service/activity.go`、`activity_test.go`、`services/product/cmd/server/main.go` |
| 2026-11-02 | A | 错误码归域：product/activity handler 的 JSON/路径参数错误改用商品域 `20002`、活动域 `20004`，更新单测与 ADR-0010 | 完成 | M | `services/product/internal/handler/{product,activity}.go`、`{product,activity}_test.go`、`internal/errs/errs.go`、`docs/DECISIONS.md` |
| 2026-11-05 | A | MySQL 配置/DSN 统一：`MYSQL_DATABASE`、单一 `MYSQL_DSN`、驱动生成 DSN、连接/读写超时、启动 Ping deadline；ADR-0014 | 完成 | M | `Makefile`、`internal/config/config.go`、`config_test.go`、`services/{user,product}/cmd/server/main.go`、`deploy/.env.example`、`README.md`、`docs/DECISIONS.md` |

## 已知问题 / 待办池

- 本机原生 MySQL 占用 3306，compose 使用 3307（ADR-0007）。
- 工具链已装（`make tools`）：golangci-lint `v2.12.2` / goimports `v0.49.0` / golang-migrate `v4.19.1`。**注意**：golangci-lint `v2.13+` 要求 Go ≥1.26，当前 Go 为 1.25.0，故钉在 v2.12.2；k6/wrk 到里程碑 C。
- MySQL 的 `--` 行注释**后面必须跟空格**（`--bcrypt...` 会报 1064 语法错）。
- 应用配置走环境变量（见 `internal/config`）：`HTTP_ADDR`（默认 `:8080`，各服务需自行覆盖，如 user `:8081`）、`MYSQL_HOST/PORT/DATABASE/USER/PASSWORD`（密码缺省回退 `MYSQL_ROOT_PASSWORD`；`MYSQL_DSN` 可整体覆盖，格式为 go-sql-driver DSN、不带 `mysql://`，`make migrate` 会自动补前缀）、`MYSQL_DIAL_TIMEOUT`/`MYSQL_READ_TIMEOUT`/`MYSQL_WRITE_TIMEOUT`（默认 `3s`/`5s`/`5s`）、`JWT_SECRET`（当前共享加载必填，product 也被迫要求，待解耦）/`JWT_TTL`/`JWT_ISSUER`、`LOG_FORMAT`/`LOG_LEVEL`、`PRODUCT_CACHE_TTL`/`PRODUCT_CACHE_NULL_TTL`。`config.Load()` 默认 best-effort 加载 `deploy/.env`（`ENV_FILE` 可改），已有环境变量优先；本地 `deploy/.env` 已补 `JWT_SECRET`。
- 迁移失败会把 `schema_migrations` 置为 dirty：若失败发生在首个 DDL 且未建任何业务表，直接 `DROP TABLE schema_migrations;` 后重跑 `make migrate`；`make migrate-force V=<上一版本>` 可用于已知版本的恢复（勿用 `V=0`，它会被当成真实版本）。
- product repository 已有 Cache Aside 装饰器单测（fake cache），真实 SQL 与真实 Redis 序列化仍依赖 E2E，尚未接入 testcontainers/miniredis；活动接口的「预热触发」按计划后置到里程碑 B。
- 缓存只完成了「穿透」防护（负缓存）；「击穿」（singleflight/逻辑过期）与「雪崩」（TTL 抖动）计划在压测前补齐，见 ADR-0012「影响与后续」。
- Redis 客户端当前为 300ms IO 超时，但未显式设置 `MaxRetries`；go-redis 默认 3 次重试，故障时可能放大请求延迟，待加 `MaxRetries`/快速失败/熔断。

### 代码审查待修复（2026-10-30，2026-11-05 更新）

> 2026-11-02 已修复活动创建校验与错误码跨域；2026-11-05 MySQL 配置/DSN 改造已完成并通过 `make lint && make test && make build`；以下为其他剩余项。

- **高**：`POST /products`、`POST /activities` 当前无鉴权（`services/product/internal/handler/router.go`）；活动创建校验已于 2026-11-02 修复。
- **中**：登录对「用户不存在/密码错误」返回不同错误码，存在用户名枚举；`User.PasswordHash` 未加 `json:"-"`。
- **中**：`gin.Recovery`/未知路由未走统一响应；`response.Fail` 对内部错误只返回 500、不记录 cause。
- **中**：配置校验不足：`JWT_TTL<=0`、连接池 `<=0`、`JWT_SECRET` 最小长度未校验。
- **中**：product 服务因共享 `config.Load()` 被迫要求 `JWT_SECRET`，即使自身不签发/校验 JWT。
- **中**：ADR-0002「只有 order 写 `seckill_stock`」与 product 创建活动时写入初始库存存在语义冲突，已在 **ADR-0013** 明确「product 只在创建时写初始值，order 只负责运行时扣减/回补」；代码注释/表注释按此口径补齐。
- **低**：分页 `(page-1)*size` 可能溢出；`loc=Local` 依赖宿主机时区；`_ = err` 若干处未按规范注释；`Makefile` 的 `migrate-down` 错误提示仍写 `MYSQL_ROOT_PASSWORD`（实际检查 `MYSQL_PASSWORD`），不影响功能；README 已补最小启动/文档索引，API 文档仍待后续补。
