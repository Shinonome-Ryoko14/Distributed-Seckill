# 项目进度

> 规则：每完成一个任务必须更新本文件。AI 每次开工**先读本页顶部「当前状态」**，再按 `harness_project.md` §17 的权重规则读「任务日志」。

## 当前状态（最高读取权重 · 开工锚点）

- **当前里程碑**: A — 同步直连全链路（每域一服务，W1–W3）
- **当前任务**: product 活动主数据接口已落地（`services/product/*`：model/repo/service + Gin handler/router/main，活动域错误码 20003/20004，service/handler 单测，导出符号注释齐全）；`make lint && make test && make build` 全绿
- **上次会话结束点**: 2026-10-17 完成 product 活动主数据接口（POST/GET activities、RFC3339 时间边界解析、先查 product 再校验、补 13 处导出注释）；此前（2026-10-10）product 商品接口、（2026-10-04）user 注册/登录全链路
- **下一步**: 进入 W2：实现 product 商品详情缓存（Cache Aside + TTL + 空值防穿透）
- **范围基线**: 里程碑 C 采用「最小可信集」（ADR-0011）；A/B 做透，目标 **2026-11 底–12 月中收口**，2027-01–06 做第二个 Agent 项目
- **阻塞**: 无

## 影响分级图例

| 级别 | 含义 |
|---|---|
| `L` 大 | 跨模块/跨服务、数据模型或行为变更、架构调整 |
| `M` 中 | 单模块内行为变更 |
| `S` 小 | 文档/格式/无行为变更 |

## 里程碑

- [ ] **A 同步直连全链路（每域一服务，W1–W3）**：数据模型、用户/商品/活动、DB 防超卖、一人一单、超时取消+回补、v0 基线压测
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

## 已知问题 / 待办池

- 本机原生 MySQL 占用 3306，compose 使用 3307（ADR-0007）。
- 工具链已装（`make tools`）：golangci-lint `v2.12.2` / goimports `v0.49.0` / golang-migrate `v4.19.1`。**注意**：golangci-lint `v2.13+` 要求 Go ≥1.26，当前 Go 为 1.25.0，故钉在 v2.12.2；k6/wrk 到里程碑 C。
- MySQL 的 `--` 行注释**后面必须跟空格**（`--bcrypt...` 会报 1064 语法错）。
- 应用配置走环境变量（见 `internal/config`）：`HTTP_ADDR`（默认 `:8080`，各服务需自行覆盖，如 user `:8081`）、`MYSQL_HOST/PORT/DATABASE/USER/PASSWORD`（密码缺省回退 `MYSQL_ROOT_PASSWORD`，亦可用 `MYSQL_DSN` 整体覆盖）、`JWT_SECRET`（必填）/`JWT_TTL`/`JWT_ISSUER`、`LOG_FORMAT`/`LOG_LEVEL`。`config.Load()` 默认 best-effort 加载 `deploy/.env`（`ENV_FILE` 可改），已有环境变量优先；本地 `deploy/.env` 已补 `JWT_SECRET`。
- 迁移失败会把 `schema_migrations` 置为 dirty：若失败发生在首个 DDL 且未建任何业务表，直接 `DROP TABLE schema_migrations;` 后重跑 `make migrate`；`make migrate-force V=<上一版本>` 可用于已知版本的恢复（勿用 `V=0`，它会被当成真实版本）。
- repository 层暂无自动化测试（单测使用 fake repo），真实 SQL 依赖 E2E 冒烟；活动接口的「预热触发」按计划后置到里程碑 B，活动 status 与 start/end 时间自洽校验暂未做。
