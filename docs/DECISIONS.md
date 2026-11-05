# 架构决策记录（ADR）

> 规则：任何"为什么用 X 不用 Y"的决策都记在这里。新增决策追加编号，不修改历史条目（若要推翻，新开一条并把旧条目标为"已废弃"）。
> 格式：编号 / 日期 / 状态 / 背景 / 决策 / 理由 / 被否方案。

---

## ADR-0001: 用 Redis+Lua 闸门 + Kafka 异步削峰替代"同步直连 MySQL"

- **日期**: 2026-09-13
- **状态**: 已接受
- **背景**: 10 万请求瞬时打到 `UPDATE stock`，导致连接池耗尽、CPU 飙升、超卖。核心命题是把高并发流量挡在数据库之外，并把建单从抢购请求里解耦。
- **决策**: 抢购请求先经过 Redis + Lua 原子闸门（查重→查库存→DECR→SADD 一次完成），只放行有限名额；抢到的请求写 Kafka `seckill-order`，由 order 消费者异步落库。
- **理由**: Lua 在 Redis 单线程内原子执行，天然无锁且并发正确；Kafka 持久化、可重放、可横向扩展，实现削峰。进程内 queue/channel 不持久、不可扩展，不适用于分布式场景。
- **被否方案**: ①进程内 channel/slice 缓冲（重启丢失、单机瓶颈）；②用分布式锁串行扣减（阻塞、性能差，原子计数可替代）；③同步直连 MySQL（V0 基线，仅用于留存对比数据）。

## ADR-0002: 库存单一事实源为 MySQL，Redis 仅作闸门

- **日期**: 2026-09-13
- **状态**: 已接受
- **背景**: 若 Redis 与 MySQL 各维护一套库存，会引入"两套账需要强一致"的复杂度。
- **决策**: `seckill_activities.seckill_stock` 是唯一事实源，只有 order 服务写该列；Redis `seckill:stock:{id}` 只是预热闸门值，允许与 DB 有偏差，靠预热刷新 + 回补 + 事后对账收敛（最终一致）。
- **理由**: 避免强一致分布式事务，符合单人 8 周的投入产出比；DB 唯一索引与条件更新作为防超卖/防重复的最后防线。
- **被否方案**: Redis 作为库存事实源（需解决持久化与一致性，复杂度高）；引入独立 Inventory 服务（两套账，已列入刻意不做清单）。

## ADR-0003: MySQL 访问层用 database/sql + 手写 SQL

- **日期**: 2026-09-13
- **状态**: 已接受
- **背景**: 需要选择 repository 层的数据访问方式。
- **决策**: 使用标准库 `database/sql` 配合手写 SQL 与 `golang-migrate` 迁移。
- **理由**: SQL 完全可控，便于做条件更新、事务与索引优化，也便于面试讲清 EXPLAIN/索引/MVCC；避免 ORM 隐藏查询。
- **被否方案**: GORM（CRUD 快但事务/锁不透明，易产生隐藏查询）；sqlc（类型安全但多一套 codegen 流程）。

## ADR-0004: Kafka 客户端用 franz-go

- **日期**: 2026-09-13
- **状态**: 已接受
- **背景**: 里程碑 B 需要 Kafka 生产者/消费者，涉及消费者组、重试、DLQ、幂等。
- **决策**: 使用 `github.com/twmb/franz-go`。
- **理由**: 纯 Go、性能优、功能完整（消费者组、事务、精确一次语义），社区活跃。
- **被否方案**: `segmentio/kafka-go`（API 简洁但高级特性较弱）；`IBM/Shopify Sarama`（API 繁琐、维护一度停滞）。

## ADR-0005: 结构化日志用标准库 log/slog

- **日期**: 2026-09-13
- **状态**: 已接受
- **背景**: 需要结构化日志，并在全链路透传 trace_id。
- **决策**: 使用 Go 1.21+ 内置 `log/slog`，JSON/Text 可切换。
- **理由**: 零额外依赖，符合依赖克制原则，功能足以覆盖本项目。
- **被否方案**: zap / zerolog（性能更高但引入额外依赖，本项目无必要）。

## ADR-0006: 数据库迁移用 golang-migrate

- **日期**: 2026-09-13
- **状态**: 已接受
- **背景**: `migrations/*.sql` 需要一个执行与版本管理工具。
- **决策**: 使用 `golang-migrate`，SQL 文件按顺序编号，通过 `make migrate` 执行。
- **理由**: 行业通用，支持 up/down 与版本表，SQL 文件保持纯 SQL 便于审计。
- **被否方案**: goose（可行但知名度略低）；自写 migrate 脚本（重复造轮子，易出边界问题）。

## ADR-0007: 开发环境 MySQL 宿主端口映射为 3307

- **日期**: 2026-09-13
- **状态**: 已接受
- **背景**: 本机原生 MySQL 服务占用 3306，且 docker-compose 需要一键起环境。
- **决策**: compose 中 MySQL 容器内仍为 3306，映射到宿主机 3307；不动本机原生 MySQL。
- **理由**: 保证 compose 环境可复现、可一键起，同时不影响本机现有服务。
- **被否方案**: 停掉原生 MySQL 改用 3306（影响本机其他依赖）；里程碑 A 直接用原生 MySQL（可复现性差）。

## ADR-0008: 代码质量门使用 gofmt + goimports + golangci-lint

- **日期**: 2026-09-13
- **状态**: 已接受
- **背景**: 需要统一代码风格并自动拦截常见问题（未处理错误、资源泄漏、缺失 ctx 等）。
- **决策**: `gofmt` 格式化、`goimports` 管理 import 分组、`golangci-lint` 静态检查，配置固化于 `.golangci.yml`。
- **理由**: 标准且够严，不引入过度约束，`make lint` 一键执行。
- **被否方案**: gofumpt 严格模式（学习成本略高）；仅 gofmt+go vet（覆盖不足）。

## ADR-0009: 用 ADR + PROGRESS 双文档防止上下文遗忘

- **日期**: 2026-09-13
- **状态**: 已接受
- **背景**: 单人长期开发，项目决策与进度若无沉淀，AI 与开发者都会反复遗忘、重复讨论。
- **决策**: `docs/DECISIONS.md` 记录架构决策（ADR），`docs/PROGRESS.md` 记录里程碑与任务日志，任务完成必须更新。
- **理由**: 决策理由与当前状态分离，便于检索与交接，AI 每次开工先读 PROGRESS。
- **被否方案**: 只维护 harness 里程碑清单（决策理由无处沉淀）；全部塞进单一 AGENTS.md（会越来越长、难检索）。

## ADR-0010: 统一响应契约、错误码分段与 JWT 规范

- **日期**: 2026-09-17
- **状态**: 已接受
- **背景**: 各服务即将各自暴露 HTTP 接口并需要鉴权，若无统一约定会出现响应格式/错误语义/token 载荷不一致，网关与前端难以对接。
- **决策**:
  - HTTP 统一响应体 `{code, message, data}`：成功 `code=0, message=ok`，`data` 为空时省略；失败由 `internal/errs.AppError` 的 `Code/Message/HTTPStatus` 映射（未知错误统一 `50000/500`，不泄漏内部 cause）。
  - 错误码分段：`0` 成功；`1xxxx` 用户域（`10001–10004`）；`2xxxx` 商品域（商品 `20001–20002`、活动子域 `20003–20004`）；`5xxxx` 系统域（`50000` 内部错误）。新增服务在业务域内分配独立号段。
  - 修订（2026-11-02）：handler 的 JSON 解析/路径参数错误也必须使用本域错误码；product 服务中商品接口使用 `20002`、活动接口使用 `20004`，不得复用用户域 `CodeInvalidParam(10001)`。
  - JWT：HS256；载荷含自定义 `uid` 与标准 `sub/iss/iat/exp`；默认 TTL `24h`、issuer `distributed-seckill`；客户端以 `Authorization: Bearer <token>` 传递；服务端校验签名、算法白名单（仅 HS256）、issuer、有效期。
  - `JWT_SECRET` 必填，仅经环境变量注入，不入库；本地开发密钥放 `deploy/.env`（已 gitignore），`deploy/.env.example` 保留空占位。
- **理由**: 统一契约便于网关/前端与错误定位；错误码分段可扩展且可读；HS256 单密钥满足"网关统一签发、各服务校验"的形态；显式算法白名单防 alg 混淆类攻击。
- **被否方案**: 直接返回裸 JSON + HTTP 状态码（错误语义不统一）；RS256 非对称（密钥管理复杂，本项目单网关签发无必要）；把 secret 写死默认值（违反密钥不入库）。

## ADR-0011: 里程碑 C 收敛为「最小可信集」以控制工期

- **日期**: 2026-09-17
- **状态**: 已接受
- **背景**: 项目目标是"学习 + 2027-06 前至少两个拿得出手的项目"。面试价值主要集中在 A（正确性）与 B（Kafka/Lua/幂等），C（工程化/压测）边际收益递减；按当前 10–15h/周，完整 C 会挤占第二个 Agent 项目的时间。
- **决策**: 里程碑 C 只交付「最小可信集」：
  1. docker-compose 一键起 5 服务 + mysql/redis/kafka/prometheus
  2. Gin Prometheus 基础 metrics（QPS/延迟/consumer lag 等关键指标）
  3. k6 压测 1k→10k 至少一轮，产出"优化前后对比"数据
  4. 一次故障注入（kill redis 或 kafka）并记录行为与恢复
  5. GitHub Actions：lint + unit test + build
  6. README + 架构图 + 启动文档 + 简历素材三句话
  后置/可砍：Grafana 花哨仪表盘（保留 1–2 个核心 panel）、多轮精细调优、完整面试题自测清单、独立缓冲周。
- **理由**: 保住 A/B 的深度（面试主战场），用最小成本拿到"可观测性 + CI + 压测"的工程化叙事；把释放的时间投入第二个 Agent 项目，形成"高并发后端 + AI Agent"的互补组合。
- **被否方案**: 完整 C（时间换边际收益，挤压第二项目）；完全砍掉 C（失去工程化叙事）；无期限打磨到完美（与"两个项目"节奏冲突）。

## ADR-0012: product 商品详情缓存采用 Cache Aside + 负缓存 + Redis 故障降级

- **日期**: 2026-10-30
- **状态**: 已接受（实现提交 `b534087`，2026-10-26）
- **背景**: W2 需要降低商品详情读请求对 MySQL 的压力，同时避免缓存穿透；product 服务对商品详情的读路径使用缓存，库存的真实值仍只以 MySQL 为准（ADR-0002）。项目是单机学习环境，缓存允许最终一致，不引入分布式事务。
- **决策**:
  - 只缓存单商品读取 `GetByID`；`List`/`Count` 以及活动创建时的商品校验直接走 DB，不做列表缓存。
  - Key 为 `product:detail:{id}`，值为 `model.Product` 的 JSON；负缓存写 JSON `null`，读取时用 `(product, hit, error)` 区分「未命中」和「命中但为空」。
  - 读路径采用 Cache Aside：先 `GET`，命中直接返回；未命中或 Redis 读失败回源 MySQL，命中后回填缓存，未找到则写负缓存。
  - 默认 TTL：正常值 5 分钟，负缓存 1 分钟；均可用 `PRODUCT_CACHE_TTL` / `PRODUCT_CACHE_NULL_TTL` 覆盖，并校验负缓存 TTL 不超过正常 TTL。
  - 创建商品后删除对应 key；当前没有商品更新/删除接口，因此暂无其他失效路径。
  - Redis 只作为缓存，不是商品事实源；Redis 启动 Ping 失败只告警不阻断启动，运行期读失败降级到 MySQL。单次 Redis IO 超时为 300ms。
- **理由**:
  - Cache Aside 与现有 repository 分层兼容，改动集中在 `CachedProductRepository` 装饰器，不污染 service/handler。
  - 负缓存用 JSON `null` 可以和正常值共用同一序列化路径，避免引入额外的哨兵值约定。
  - Redis 故障时读请求仍能回源 DB，优先保证可用性；缓存最终一致符合秒杀项目的取舍。
  - 只缓存详情、不缓存列表，避免分页/总数失效复杂度，符合单人项目投入产出比。
- **被否方案**:
  - Write Through / 双写：当前没有商品更新接口，且双写一致性复杂度高于收益。
  - 把 Redis 当商品事实源：与 ADR-0002「MySQL 唯一事实源」冲突。
  - 进程内本地缓存：多实例下失效困难，且与后续分布式叙事不一致。
  - 列表/总数缓存：分页参数组合多、失效复杂，先不做。
  - 逻辑过期 / singleflight：W2 基线只要求穿透防护，击穿与雪崩留到压测前补齐，不在本次强行引入。
- **影响与后续**:
  - 当前只解决穿透；击穿（singleflight/逻辑过期）与雪崩（TTL 随机抖动）待压测前实现。
  - 300ms IO 超时 + go-redis 默认重试可能在 Redis 故障时放大延迟，后续需显式设置 `MaxRetries`、快速失败/熔断，并对缓存告警日志采样。
  - 真实 Redis 序列化（尤其负缓存 `null`）尚无自动化测试，后续补 miniredis 或 testcontainers 集成测试。
  - Key 暂时没有版本前缀；若 `model.Product` JSON 结构发生不兼容变更，需要加 `v1` 前缀或缩短 TTL 过渡。
  - 新增商品更新/删除接口时，必须同步补缓存失效路径。

## ADR-0013: `seckill_stock` 的写入边界

- **日期**: 2026-10-30
- **状态**: 已接受
- **背景**: ADR-0002 规定「`seckill_activities.seckill_stock` 是唯一事实源，只有 order 服务写该列」；但计划 §2.1 又把 `seckill_activities` 的 owner 标为 product，且 product 创建活动时必须写入初始库存。两句话放在一起会被理解为 product 不得写 stock，和现有实现冲突。
- **决策**:
  - product 服务拥有活动主数据，允许在**创建活动时写入一次** `seckill_stock` 初始值。
  - 活动创建完成后，只有 order 服务可以修改 `seckill_stock`：扣减真实库存、超时/取消回补。
  - 其他服务、脚本或人工修数不得直接改该列；确需修数时必须走审计脚本并记录 `inventory_op_logs`（里程碑 A 后续表）。
- **理由**: 保留「库存单一事实源」和「一个服务写入」的核心约束，同时承认活动创建时的初始值是 product 的职责；避免为了字面上的“只有 order 写”把活动创建挪到 order，破坏领域边界。
- **被否方案**:
  - product 同时负责扣减/回补：出现两个运行时写者，违背 ADR-0002。
  - 活动创建也交给 order：活动主数据与订单域耦合，且与计划 §2.1 的 owner 约定冲突。
  - 独立 Inventory 服务：已列入刻意不做清单。
- **影响**: 后续代码/表注释应区分「创建时初始化」与「运行时扣减/回补」；对账与审计脚本按 order 服务为唯一运行时写者设计。

## ADR-0014: MySQL 配置与 DSN 格式统一

- **日期**: 2026-11-05
- **状态**: 已接受
- **背景**: Makefile 使用 `MYSQL_DB` 和 `mysql://` 形式的 `MYSQL_DSN`，应用使用 `MYSQL_DATABASE` 和 go-sql-driver DSN；迁移与应用可能连到不同库或使用不同格式。旧实现还直接拼接 DSN，缺少连接/读写超时，启动 Ping 使用无 deadline 的 `context.Background()`。
- **决策**:
  - 库名统一使用 `MYSQL_DATABASE`，不再使用 `MYSQL_DB`；密码使用 `MYSQL_PASSWORD`，缺省回退 `MYSQL_ROOT_PASSWORD`。
  - `MYSQL_DSN` 统一为 go-sql-driver 格式（不带 `mysql://`），应用和 Makefile 共用同一份连接信息；`make migrate` 调用 golang-migrate 时自动补 `mysql://` 前缀。
  - 应用使用 `mysql.NewConfig()` / `FormatDSN()` 生成默认 DSN，使用 `mysql.ParseDSN()` 校验覆盖值；兼容旧的 `mysql://` 覆盖写法（去除前缀后解析）。
  - 默认超时：连接 `3s`、读/写各 `5s`，可用 `MYSQL_DIAL_TIMEOUT` / `MYSQL_READ_TIMEOUT` / `MYSQL_WRITE_TIMEOUT` 覆盖；启动 `db.Open` 使用 5 秒 context 限制 Ping。
  - `MYSQL_DSN` 覆盖值缺少超时时补默认值，`parseTime` 强制为 `true`；覆盖值仍需自行保证时区参数符合预期。
- **理由**: 组件变量是单一事实源，迁移与应用不会因为变量名或 DSN 格式不同而连到不同库；由驱动生成/解析 DSN 比手拼更可控，超时和启动 deadline 避免故障时无限等待。
- **被否方案**:
  - 拆分 `MYSQL_DSN`（应用）与 `MYSQL_MIGRATE_DSN`（迁移）：格式隔离清楚，但需要维护两套变量，容易再次不一致。
  - 继续手拼 DSN：密码含特殊字符、参数遗漏和格式变化时容易出错。
  - 应用直接接受 `mysql://` 作为规范格式：与 go-sql-driver 原生格式不一致，且迁移与应用仍各自解析。
  - 不设置超时：故障时请求/启动可能长时间挂起，不利于快速失败。
- **影响**: `.env.example` / README 需要明确 `MYSQL_DSN` 格式与超时变量；`MYSQL_DSN` 覆盖是专家用法，必须与 `MYSQL_DATABASE` 指向同一库；后续 Redis 重试与连接池参数校验单独处理。
