# git subtree 备忘

把 monorepo 里的一个子目录拆成独立仓库、或把独立仓库嵌进来，**同时保留提交历史**。这里只记我实际会用到的部分。

## 一句话

`git subtree split` = 把某个子目录的**历史**投影成一个独立分支（路径重写为根，提交保留）。
不是文件拷贝 —— 直接 `cp -r` 出去再 `git init` 会丢掉所有 commit。

## 拆出去（monorepo → 独立仓库）

```bash
cd ~/workspaces/demos

# 1. 生成只含 go/wordfreq 历史的分支，路径已重写成仓库根
git subtree split -P go/wordfreq -b wordfreq

# 2. 检查：main.go 现在在根目录，历史完整
git log --oneline wordfreq
git ls-tree --name-only wordfreq

# 3. 建空仓库并推上去
gh repo create wordfreq --public --source=. --remote=wordfreq-origin
git push wordfreq-origin wordfreq:main

# 4. 清理
git branch -D wordfreq
git remote remove wordfreq-origin
```

`-P` 是 `--prefix`，必填，指向子目录。`-b` 是目标分支名。

注意：

- 只有**改动过该目录**的提交会进来，改 `web/` 的提交不会出现。
- 提交信息不变，但 SHA 会变（树变了）。
- 可以重复执行，同样的输入得到同样的结果（所以能当缓存用）。

## 反向：嵌进来 / 双向同步

```bash
# 把外部仓库嵌成子目录（当 vendored 依赖），--squash 把上游多次提交压成一个
git subtree add --prefix=third_party/foo https://github.com/x/foo.git main --squash

# 本地子目录的改动推回上游
git subtree push --prefix=third_party/foo https://github.com/x/foo.git main

# 拉上游更新
git subtree pull --prefix=third_party/foo https://github.com/x/foo.git main --squash
```

`add` 和 `pull` 基本都该加 `--squash`，否则上游历史会灌进本仓库。

## 坑

| 问题 | 说明 |
|---|---|
| 慢 | 要遍历整个历史，大仓库几秒到几分钟 |
| 改名断链 | 目录搬过位置（如 `demos/go/wordfreq` → `go/wordfreq`），改名**之前**的提交追溯不到。`--annotate` 只标注，不解决 |
| 一次性提取 | 更推荐 `git filter-repo --path go/wordfreq/`（更快），但它是破坏性重写且需额外安装 |
| 历史本来就短 | 只有 1–2 个 commit 的 demo，直接复制 + `git init` 更省事 |
| 忘了 `-P` | 会报错或者拆出错的东西，一定带上前缀 |

## 什么时候值得拆

- 超过 ~1k 行，或有外部使用者（需要独立 module 路径才能 `go get` / npm 安装）
- 需要自己的 CI、release、tag、issue 追踪
- 会持续演进（见 [README](../README.md#when-to-split-a-demo-into-its-own-repo)）
- 有独立部署

否则就留在 `demos` 里 —— 拆出去容易，事后合并难。
