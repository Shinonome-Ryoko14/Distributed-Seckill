# 项目进度

> 规则：每完成一个任务必须更新本文件。AI 每次开工**先读本页顶部「当前状态」**，再按 `harness_project.md` §17 的权重规则读「任务日志」。

## 当前状态（最高读取权重 · 开工锚点）

- **当前里程碑**: A — 单体全链路（W1–W3）
- **当前任务**: 环境搭建（docker-compose: mysql + redis）
- **上次会话结束点**: 2026-09-13 制定每日开工/收工流程（harness §17–§18 + DAILY_KICKOFF）
- **下一步**: 按下方任务日志中状态为「待办」的首项执行
- **阻塞**: 无

## 影响分级图例

| 级别 | 含义 |
|---|---|
| `L` 大 | 跨模块/跨服务、数据模型或行为变更、架构调整 |
| `M` 中 | 单模块内行为变更 |
| `S` 小 | 文档/格式/无行为变更 |

## 里程碑

- [ ] **A 单体全链路（W1–W3）**：数据模型、用户/商品/活动、DB 防超卖、一人一单、超时取消+回补、v0 基线压测
- [ ] **B 拆服务+Kafka（W4–W6）**：gateway+4 服务、Lua 原子门+预热、异步建单、幂等/DLQ、限流
- [ ] **C 工程化+压测（W7–W8）**：docker-compose、Prometheus/Grafana、CI、k6 优化、故障注入、报告

## 任务日志

| 日期 | 里程碑 | 任务 | 状态 | 影响 | 涉及文件 |
|---|---|---|---|---|---|
| 2026-09-13 | A | 制定工程规范并写入 harness §10–§16 | 完成 | L | `harness_project.md` |
| 2026-09-13 | A | 建立 docs/DECISIONS.md、docs/PROGRESS.md（9 条 ADR） | 完成 | M | `docs/DECISIONS.md`、`docs/PROGRESS.md` |
| 2026-09-13 | A | 制定每日开工/收工流程 | 完成 | M | `harness_project.md`、`docs/DAILY_KICKOFF.md`、`docs/PROGRESS.md` |
| - | A | `deploy/docker-compose.yml`（mysql:3307 + redis:6379） | 待办 | L | `deploy/*` |
| - | A | `Makefile` 与 `.golangci.yml`、`make tools` | 待办 | M | `Makefile`、`.golangci.yml` |
| - | A | `migrations/0001_init.sql`（users/products/seckill_activities） | 待办 | L | `migrations/0001_init.sql` |
| - | A | 用户注册/登录 + JWT | 待办 | L | `services/user/*`、`internal/jwt` |

## 已知问题 / 待办池

- 本机原生 MySQL 占用 3306，compose 使用 3307（ADR-0007）。
- 工具链待安装：golangci-lint、goimports、golang-migrate（`make tools`）；k6/wrk 到里程碑 C。
