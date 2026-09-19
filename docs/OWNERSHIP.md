# 基础设施所有权补课包（internal/*）

> 背景：`internal/{config,log,errs,ctxkey,db,jwt,response}` 由 AI 代写（用户逐次确认）。本文件的目标是把这些代码转换成"你自己的"。
> 用法：先读「概念速览」理解 → 合上文件做「自测题」→ 展开 `<details>` 对答案 → 做「破坏性实验」→ 完成「默写目标」。
> 门槛：完成 `config` / `errs` / `jwt` 三件默写 + 全部破坏性实验，方可进入任务 2（user 服务）。**未通过不得在其上继续堆功能。**

---

## 0. 全局：一次请求的链路

```
HTTP 请求
  → handler      校验入参、调 service、用 response 写回
  → service      业务逻辑（注册 bcrypt、登录签发 JWT）
  → repository   手写 SQL 访问 MySQL
    ↑ 依赖：config 提供配置、db 提供连接池、log 打日志、jwt 鉴权、errs 表达错误
```

自测：合上文件，凭记忆说出这三个箭头各自的职责，以及 `internal/` 每个包被谁用到。

---

## 1. internal/config

### 概念速览
- **12-factor 配置**：配置放环境变量，不放代码；密钥不入库。
- **优先级**：进程环境变量 > `deploy/.env` > 代码内默认值。越"临时/机器相关"的越靠后。
- **可测试性**：`load(lookup func(string) string)` 把"读环境"这一步做成参数注入，测试时可以传假函数，不去污染真实进程环境。
- **DSN**：`user:pass@tcp(host:port)/dbname?参数`。`parseTime=true` 让 MySQL 的 `TIMESTAMP` 扫进 Go 变成 `time.Time`，否则是 `[]byte`。

### 自测题
<details><summary>Q1 为什么 `Load()` 里不直接到处 `os.Getenv`，而要多套一层 `load(lookup)`？</summary>

为了可测试。真实 `os.Getenv` 是全局状态，测试里改它会影响其他测试且难清理；把"查环境"抽象成函数参数后，测试可以传一个 map 查找函数，完全隔离。这叫"依赖注入"的最简形式。
</details>

<details><summary>Q2 `MYSQL_DSN` 里不带 `parseTime=true` 会发生什么？</summary>

查 `TIMESTAMP/DATETIME` 列得到 `[]byte`（或 string），无法直接 `Scan` 进 `time.Time`，代码会报类型错误或拿到错误数据。
</details>

<details><summary>Q3 什么情况下服务一启动就退出？</summary>

`JWT_SECRET` 为空、`MYSQL_MAX_OPEN_CONNS` 不是数字、`JWT_TTL` 不是合法时长、`MaxIdle > MaxOpen`。启动期失败好过运行期出诡异 bug。
</details>

### 破坏性实验
1. `JWT_SECRET=` 置空 → `go test ./internal/config -run RequiresJWTSecret -v` 看报错
2. `JWT_TTL=abc` → 运行 `TestLoadRejectsInvalidValues`
3. 同名环境变量已存在时加载 `.env`，确认不被覆盖（`TestLoadDotEnvDoesNotOverride`）

### 默写目标
不看代码写出 `mysqlDSN(lookup)` 的签名与拼接格式。

---

## 2. internal/config/dotenv.go

### 概念速览
- **best-effort**：文件不存在不算错误（开发机可能没有 `.env`），静默跳过。
- **不覆盖已有环境变量**：这是原则——环境变量是更高优先级来源，不能被文件冲掉。

### 自测题
<details><summary>Q 为什么"已存在的环境变量不被 .env 覆盖"很重要？</summary>

CI/容器里往往用环境变量注入配置，如果 `.env` 把线上值覆盖了，会造成"本地能跑、线上跑错"，且极难排查。优先级必须单向：更"正式"的来源优先。
</details>

---

## 3. internal/log

### 概念速览
- **结构化日志**：`key=value` / JSON，机器可解析，便于日志平台检索聚合；`fmt.Println` 是纯文本，搜不了字段。
- **`log/slog`**：Go 1.21+ 标准库，零依赖（ADR-0005）。
- **为什么 `New()` 不 `SetDefault`**：`SetDefault` 是全局副作用，应由 `main.go` 在装配期显式调用；包本身保持无状态、可测试。

### 自测题
<details><summary>Q JSON handler 相比文本日志的优势？</summary>

每条日志是合法 JSON，日志系统（Loki/ELK）能直接按字段 `level`、`trace_id`、`user_id` 做索引和过滤，不用正则解析。
</details>

<details><summary>Q `LOG_LEVEL=warn` 时 `slog.Info` 会输出吗？</summary>

不会。HandlerOptions.Level 会过滤低于该级别的日志。生产常用 info，排障临时开 debug。
</details>

---

## 4. internal/errs

### 概念速览
- **Go 的错误是值**：`error` 是接口；用 `fmt.Errorf("...: %w", err)` 包装能保留错误链。
- **`errors.Is`**：沿链比较"是不是这个哨兵错误"；**`errors.As`**：沿链把错误断言成某具体类型。
- **`AppError`**：把「业务码 + 对外消息 + HTTP 状态 + 内部 cause」绑在一起。`cause` 只进日志，不进响应。

### 自测题
<details><summary>Q 为什么要 `From(err)` 把未知错误统一成 500，而不直接把 err.Error() 返回？</summary>

直接返回会把 SQL 语句、表名、驱动细节暴露给攻击者（信息泄漏）。统一成"internal server error"，真正的 cause 记到服务端日志。
</details>

<details><summary>Q `errors.Is(err, ErrInvalidToken)` 是怎么穿过 `%w` 包装找到目标的？</summary>

每个用 `%w` 包装的错误实现了 `Unwrap()`，`errors.Is` 会逐层调用 `Unwrap` 形成错误链，直到匹配或链尾。
</details>

### 破坏性实验
构造一个未知 error 传给 `response.Fail`，确认返回 `{"code":50000,"message":"internal server error"}` 且没有原始信息。

### 默写目标
不看代码写出 `From(err)`。

---

## 5. internal/ctxkey

### 概念速览
- **context**：跨函数边界传递取消信号、超时、请求级元数据；约定作为函数**首参**。
- **`WithValue` 只放请求级元数据**（trace_id、user_id），不放业务可选参数。
- **自定义未导出 `key int`**：防止不同包用同名字符串互相覆盖。

### 自测题
<details><summary>Q 为什么用自定义类型 `type key int` 而不是字符串 "userID" 当键？</summary>

字符串键容易撞车（别的包也用 "userID"）；未导出的自定义类型只有本包能构造，天然隔离。这是官方推荐做法。
</details>

---

## 6. internal/db

### 概念速览
- **`sql.Open` 不建立连接**，只校验 DSN 并返回连接池句柄；真正连库是懒加载的，所以必须 `PingContext` 探活。
- **连接池三参数**：
  - `MaxOpenConns`：最大并发连接数，保护 MySQL 不被压垮。
  - `MaxIdleConns`：保留的空闲连接，避免频繁拨号开销。
  - `ConnMaxLifetime`：连接最长存活时间，回收陈旧连接（防 MySQL `wait_timeout`、负载均衡器断连）。
- `*sql.DB` 是并发安全的，全局一个即可，不要每次请求 Open。

### 自测题
<details><summary>Q Ping 失败时为什么要 Close 再返回错误？</summary>

Open 已经创建了池对象，若不关闭会泄漏资源（后台可能仍在尝试连接）。Close 失败无法覆盖主错误，所以 best-effort 忽略并保留原始错误。
</details>

### 破坏性实验
把 `MYSQL_PORT` 改成 3999，用 `go test` 或临时 main 触发 `db.Open`，看报错文案是否清晰。

---

## 7. internal/jwt（面试最高频）

### 概念速览
- **JWT = `base64url(header).base64url(payload).base64url(signature)`**，**只签名不加密**，payload 任何人可解码查看，所以不能放敏感信息。
- **HS256**：用共享密钥做 HMAC-SHA256，签发和校验用同一个 secret（对称）。RS256 是非对称（私钥签、公钥验），本项目单网关签发用 HS256 足够（ADR-0010）。
- **签名防篡改**：改了 payload 但没密钥就无法生成匹配签名。
- **标准声明**：`sub`（主体）、`iss`（签发者）、`iat`（签发时间）、`exp`（过期时间）；库会校验 `exp`。
- **必须校验算法**：不校验 `alg` 会遇到两类攻击：
  1. `alg=none`：攻击者把算法改成 none、去掉签名直接伪造。
  2. **算法混淆**：服务端本用 RS256，攻击者改成 HS256 并用公开的公钥当 HMAC 密钥签名。
  解决：key func 检查 `SigningMethodHMAC` + `WithValidMethods(["HS256"])` 白名单。
- **无状态**：服务端不存 session，token 自带身份；代价是不易主动失效，需靠短 TTL + 刷新或黑名单。

### 自测题
<details><summary>Q payload 里放了 user_id，用户能自己改成别人吗？</summary>

不能。改 payload 会导致签名不匹配，校验失败。但 payload 是**明文**，所以不能放密码、身份证等敏感数据。
</details>

<details><summary>Q 为什么把 `now` 做成 Manager 的字段？</summary>

为了可测试过期逻辑。`Parse` 用 `jwt.WithTimeFunc(m.now)`，测试里把 now 调到未来就能验证过期，无需 `time.Sleep`。
</details>

<details><summary>Q 为什么同时放 `uid` 和 `sub`？</summary>

`sub` 是 JWT 标准声明（字符串），放标准值便于通用工具识别；`uid` 是自定义的、强类型便利字段，业务直接 `claims.UserID` 用。
</details>

### 破坏性实验
1. token 最后一个字符改掉 → 解析失败
2. 换一个 secret 解析 → 失败
3. `NewManager(..., -1*time.Minute, ...)` 签发再解析 → 过期失败

### 默写目标
不看代码写出 `Generate` 与 `Parse`。

---

## 8. internal/response

### 概念速览
- **统一响应体** `{code, message, data}`：成功 `code=0`，`data` 空则省略；失败由 `AppError` 映射。
- **框架无关**：只依赖 `net/http`，这样还没引 gin 时基础设施就能编译测试；需要时再加 gin 薄适配。
- **业务码 ≠ HTTP 码**：HTTP 表达传输层结果，业务码表达业务语义；两者可组合（如 400 + 10001）。

### 自测题
<details><summary>Q `Fail(w, err)` 为什么只接受 error，不接收 code/message？</summary>

强制所有错误先归一化为 `AppError`，保证"错误→响应"只有一条路径，不会各 handler 各写一套；未知错误自动兜底 500 且不泄漏。
</details>

---

## 9. 横向：面试选型追问对照

| 追问 | 对照 |
|---|---|
| 日志为什么 slog 不用 zap | ADR-0005：零依赖，功能够用 |
| 数据访问为什么 database/sql 不用 GORM | ADR-0003：SQL 可控，能讲 EXPLAIN/索引/事务 |
| JWT 为什么 HS256、放哪、怎么校验 | ADR-0010 |
| 库存为什么 DB 是唯一事实源 | ADR-0002 |

## 10. 通过判据（自查）
- [ ] `config` / `errs` / `jwt` 三件默写通过
- [ ] 全部破坏性实验做完并有记录
- [ ] 能不看文档复述第 0 节的请求链路
- [ ] 能回答每个包的"为什么这么设计"

---

## 11. 调整版门槛与实践记录（2026-09-19 起）

> 背景：完整逐字默写对当前阶段过陡，改为**分层重建 + 破坏性实验 + 整体装配**。
> 本节的记录由用户实验/口述、AI 整理，属非业务产物（harness §19.2 第 5 条）。

### 11.1 调整说明
- §10 的"完整默写"调整为**原地骨架填空**：保留真实文件的类型/字段/签名，把考点函数体留成 `// TODO(user): 目标 | 验收`，用户填 body，`go test` 红改到绿——与写 SQL / YAML 同一套方法（harness §19）。
- 顺序：`errs` → `jwt` → `config`（由小到大）。
- 重建前先提交基线 commit，供 `git diff` 对照与 `git checkout` 还原。
- 新增"整体装配"验收：临时 `cmd/wirecheck` 串联 `config → log → db → response`，跑通后删除。

### 11.2 破坏性实验记录

| # | 包 | 操作 | 期望 | 实际 | 结论 |
|---|---|---|---|---|---|
| E1 | config | `JWT_SECRET` 置空调用 `load` | `JWT_SECRET is required` | `JWT_SECRET is required` | 密钥无默认值，缺失即启动期失败 |
| E2 | config | `JWT_TTL=abc` / `MYSQL_MAX_OPEN_CONNS=abc` | `must be a duration/integer` | `JWT_TTL must be a duration: time: invalid duration "abc"`；`MYSQL_MAX_OPEN_CONNS must be an integer: strconv.Atoi: parsing "abc": invalid syntax` | 配置类型错误在启动期暴露，附带原始解析错误（`%w`） |
| E3 | config | 已存在环境变量 + `.env` 同名 | 环境变量不被覆盖 | `TestLoadDotEnvDoesNotOverride` PASS，值仍为 `original` | 优先级单向：环境变量 > `.env` |
| E4 | errs | 未知 error 传给 `response.Fail` | `{"code":50000,"message":"internal server error"}`，无原始信息 | `status=500 body={"code":50000,"message":"internal server error"}`（原始 SQL 未出现） | 未知错误兜底 500，cause 只进日志不对外 |
| E5 | db | `MYSQL_PORT=3999` 调 `db.Open` | `ping mysql: ...` 且连接池被 Close | `ping mysql: dial tcp 127.0.0.1:3999: i/o timeout`，`conn=<nil>` | `sql.Open` 懒连接，Ping 探活失败即报错并释放池 |
| E6 | jwt | token 末字符改掉 | `Parse` 返回 `ErrInvalidToken` | `TestParseRejectsTamperedToken` PASS | 签名防篡改 |
| E7 | jwt | 换 secret 解析 | 同上 | `TestParseRejectsWrongSecret` PASS | 密钥不符则校验失败 |
| E8 | jwt | `TTL=-1m` 签发再解析 | 过期失败 | `TestParseRejectsExpiredToken` PASS | `exp` 过期被拒 |

> 附：`TestParseRejectsWrongIssuer` 也 PASS，确认 `iss` 被校验。

### 11.3 骨架填空清单

| 包 | 函数体填完 | `go test` 全绿 | 复述 |
|---|---|---|---|
| errs | [ ] | [ ] | [ ] |
| jwt | [ ] | [ ] | [ ] |
| config | [ ] | [ ] | [ ] |

### 11.4 整体装配验收
- [ ] `cmd/wirecheck` 能启动并返回统一 JSON 响应
- [ ] 改错 db 端口可观察到启动期失败
- [ ] 验收后删除 `cmd/wirecheck` 与全部 `zz_scratch_test.go`

### 11.5 通过判据（调整版）
- [ ] 三件套 L1 签名 100%、L2 骨架 100%、L3 现有单测全绿
- [ ] E1–E8 实验完成并记录期望/实际
- [ ] 能不看文档复述第 0 节请求链路 + 每包"为什么"
- [ ] 整体装配跑通
