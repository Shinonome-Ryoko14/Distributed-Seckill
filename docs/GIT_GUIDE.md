# Git 实用速查（单人开发 / 本项目约定）

> 目标：把「只会 add/status/pull/commit/push」补到「能看懂参数、能安全撤销、能按分支流程协作」。
> 适用：本仓库（单仓库、单人开发、Conventional Commits、feature 分支）。

---

## 0. 心智模型

```
工作区(Working Tree) --add--> 暂存区(Index) --commit--> 本地仓库 --push--> 远程(Remote)
远程 --fetch--> 本地远程分支(origin/main) --merge/rebase--> 本地分支
```

- **工作区**：你正在编辑的文件。
- **暂存区**：下一次 commit 要包含的内容。
- **本地仓库**：`.git` 里的提交历史。
- **远程**：GitHub 上的仓库。

记住一句话：**commit 只提交暂存区**，所以 `add` 到哪，`commit` 就到哪。

---

## 1. 查看状态与差异

| 命令 | 作用 |
|---|---|
| `git status` | 完整状态 |
| `git status -sb` | 简洁状态 + 当前分支 + 领先/落后远程 |
| `git diff` | 工作区 vs 暂存区（还没 add 的改动） |
| `git diff --staged` | 暂存区 vs 上次提交（**commit 前必看**） |
| `git diff HEAD~1` | 当前工作区 vs 上一个提交 |
| `git diff --stat` | 只看改了哪些文件、多少行 |
| `git log --oneline --graph --decorate -n 10` | 最近 10 条提交的图形化历史 |
| `git log --stat -n 5` | 最近 5 条提交 + 文件改动统计 |
| `git show <commit>` | 查看某次提交的完整 diff |
| `git blame <file>` | 每一行最后由谁、哪次提交修改 |

常用组合：

```bash
git status -sb
git diff --staged
git log --oneline --graph --decorate -n 15
```

---

## 2. 暂存与提交

| 命令 | 说明 |
|---|---|
| `git add <path>` | 暂存指定文件/目录 |
| `git add .` | 暂存当前目录下的改动（不含上层） |
| `git add -A` | 暂存全部改动（新增 + 修改 + 删除） |
| `git add -u` | 只暂存**已跟踪**文件的修改/删除，不包含新文件 |
| `git add -p` | 交互式分块暂存（推荐，能避免把无关改动混进 commit） |
| `git commit -m "msg"` | 提交暂存区 |
| `git commit -am "msg"` | 跳过 `add`，提交**已跟踪**文件（新文件不生效） |
| `git commit --amend` | 修改**最后一次**提交（只在自己分支、未 push 时用） |
| `git commit --amend --no-edit` | 保留原提交信息，只更新内容 |
| `git restore --staged <file>` | 取消暂存（改动仍留在工作区） |
| `git restore <file>` | 丢弃工作区改动（**不可恢复**，先用 `git diff` 确认） |

### 提交信息规范（Conventional Commits）

```
<type>: <简短描述>
```

| type | 用途 |
|---|---|
| `feat` | 新功能 |
| `fix` | 修 bug |
| `refactor` | 重构（行为不变） |
| `test` | 测试相关 |
| `docs` | 文档 |
| `chore` | 构建/工具/杂项 |

示例：

```bash
git commit -m "feat: 实现 product 活动主数据接口（含活动域错误码）"
git commit -m "fix: 修复活动 SELECT 缺少 FROM"
git commit -m "docs: 更新 PROGRESS 与 Git 速查"
```

原则：**一个 commit 只做一件事**；标题写「做了什么」，正文再写「为什么」。

---

## 3. 分支

| 命令 | 说明 |
|---|---|
| `git branch` | 列出本地分支 |
| `git branch -a` | 列出本地 + 远程分支 |
| `git branch -vv` | 本地分支 + 追踪的远程分支 |
| `git switch -c feat/xxx` | 新建并切换分支（旧写法 `git checkout -b`） |
| `git switch main` | 切换到 main |
| `git switch -` | 切回上一个分支 |
| `git branch -m <old> <new>` | 重命名分支 |
| `git branch -d feat/xxx` | 删除已合并分支 |
| `git branch -D feat/xxx` | 强制删除未合并分支（确认不要了再用） |
| `git merge --no-ff feat/xxx` | 合并并保留合并点（历史更清晰） |
| `git rebase main` | 把当前分支的提交挪到 main 最新之上（历史更线性） |

### 冲突处理

```bash
# merge 或 rebase 时出现冲突
# 1. 编辑冲突文件，保留正确内容
git add <冲突文件>
git merge --continue      # merge 场景
git rebase --continue     # rebase 场景

# 想放弃这次 merge/rebase
git merge --abort
git rebase --abort
```

---

## 4. 远程协作

| 命令 | 说明 |
|---|---|
| `git remote -v` | 查看远程仓库地址 |
| `git remote add origin <url>` | 添加远程（通常克隆时已自动配置） |
| `git fetch origin` | 只下载远程更新，**不改工作区** |
| `git pull` | `fetch` + `merge` |
| `git pull --rebase` | `fetch` + `rebase`（**推荐**，历史干净） |
| `git push -u origin feat/xxx` | 首次推送并建立追踪（`-u` = `--set-upstream`） |
| `git push` | 之后直接推送当前分支 |
| `git push origin --delete feat/xxx` | 删除远程分支 |
| `git push --force-with-lease` | 安全强推（仅限自己的分支；**禁止对 main**） |

`--force-with-lease` 与 `--force` 的区别：前者会在远程被别人更新过时拒绝推送，避免覆盖他人提交。

---

## 5. 撤销与急救

| 命令 | 说明 |
|---|---|
| `git reset --soft HEAD~1` | 撤销最后一次 commit，改动**保留在暂存区** |
| `git reset --mixed HEAD~1` | 默认；改动**保留在工作区** |
| `git reset --hard HEAD~1` | 撤销 commit 并**丢弃改动**（危险） |
| `git reset --hard origin/main` | 强制与远程 main 一致（危险，先 stash） |
| `git stash push -m "wip"` | 暂存当前未提交改动 |
| `git stash list` | 查看 stash 列表 |
| `git stash pop` | 恢复最近一次 stash 并删除记录 |
| `git stash apply` | 恢复但保留 stash 记录 |
| `git reflog` | 查看 HEAD 移动历史，**找回误删/误 reset 的 commit** |
| `git cherry-pick <commit>` | 把某个 commit 摘到当前分支 |
| `git revert <commit>` | 生成一个反向提交，安全撤销已 push 的提交 |

卡住时的顺序：

```bash
git status          # 先看现在处于什么状态
git reflog          # 找误操作前的 commit
git stash list      # 看有没有暂存的改动
```

---

## 6. 单人日常流程

```bash
# 1. 同步 main
git switch main
git pull --rebase

# 2. 开 feature 分支
git switch -c feat/activity

# 3. 开发 + 自测
make lint && make test

# 4. 分块暂存并检查
git add -p
git diff --staged

# 5. 提交
git commit -m "feat: 实现 product 活动主数据接口"

# 6. 推送
git push -u origin feat/activity

# 7. 合并回 main（或开 PR）
git switch main
git pull --rebase
git merge --no-ff feat/activity
git push

# 8. 删除已合并分支
git branch -d feat/activity
git push origin --delete feat/activity
```

---

## 7. 本项目硬规则

- 提交信息使用 Conventional Commits：`feat: / fix: / refactor: / docs: / test: / chore:`。
- 分支命名：`feat/<模块>`、`fix/<问题>`。
- 每个改动提交前跑 `make lint && make test`（以 Makefile 为准）。
- 一个 commit 只做一件事；提交前用 `git diff --staged` 自查。
- `.env`、密钥、临时文件不入库（`.gitignore` 已覆盖）。
- 共享分支（`main`）禁止 `--force`；自己的 feature 分支必要时用 `--force-with-lease`。
- 完成定义（DoD）：代码 + 测试 + `make lint && make test` 通过 + 文档/决策更新，四者缺一不可。

---

## 8. 危险操作清单

执行前先确认，必要时先 `git stash` 或 `git branch backup-xxx`：

```bash
git reset --hard <commit>      # 丢弃工作区改动
git clean -fd                  # 删除未跟踪文件/目录（先 git clean -n 预览）
git push --force               # 覆盖远程历史
git branch -D <branch>         # 强删未合并分支
git checkout -- <file>         # 丢弃文件改动（等价 git restore <file>）
```

安全习惯：

```bash
git status              # 动手前先看
git diff                # 确认改了什么
git stash push -m "..." # 不确定就先存起来
git reflog              # 出事后第一反应
```
