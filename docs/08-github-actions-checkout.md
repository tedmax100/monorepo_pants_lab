# GitHub Actions：fetch-depth 策略

## 問題：`fetch-depth: 0` 會不會很慢？

**會慢，但問題不在 `fetch-depth: 0`，而在你有沒有加 `filter`。**

Git 倉庫的大小分兩層：

```
Git 倉庫的組成：
  ┌──────────────────────────────────────────────────────┐
  │  commit objects  提交記錄、作者、時間戳              │  ← 通常幾 MB
  │  tree objects    每個 commit 的目錄結構              │  ← 中等
  │  blob objects    所有版本的實際檔案內容              │  ← 這才是大的，可能 GB+
  └──────────────────────────────────────────────────────┘

fetch-depth: 0                  → 全部都抓（commits + trees + blobs）
fetch-depth: 0 + filter: blob:none → 只抓 commit + tree，不抓檔案內容
fetch-depth: 1                  → 只抓最新一個 commit（但算不了 diff）
```

Pants 的 `--changed-since` 只需要 **commit history** 來計算 merge-base，
完全不需要每個檔案的歷史內容（blob）。

---

## 三種策略比較

| 策略 | 設定 | 速度 | 安全性 | 適用情境 |
|------|------|------|--------|---------|
| **A. Blobless clone** | `fetch-depth: 0` + `filter: blob:none` | 快 | 完全可靠 | **推薦，大部分情況** |
| B. 完整 clone | `fetch-depth: 0` | 慢 | 完全可靠 | 小 repo、或需要檔案內容 |
| C. 固定 depth | `fetch-depth: 50` | 最快 | 有邊緣情況 | PR 分支短、追求極速 |

---

## 策略 A：Blobless Clone（推薦）

```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0        # 抓所有 commit 歷史（算 merge-base 用）
    filter: blob:none     # 不抓檔案內容，需要時按需取得
```

`filter: blob:none` 告訴 GitHub 的 git server：
「只傳 commit 和 tree 物件，blob 先不傳」。

實際效果：
- 一個有 3 年歷史、2000 個 commit 的 repo
- 完整 clone：可能 500 MB，需要 30 秒
- Blobless clone：通常 10~30 MB，需要 3~5 秒

Pants 的 `git diff --name-only` 只需要 commit 和 tree，
所以 blobless clone 對它來說完全夠用。

---

## 策略 C：固定 Depth（最快，有風險）

```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 50         # 只抓最近 50 個 commit

- name: Fetch base branch
  run: |
    git fetch --depth=50 origin ${{ github.base_ref }}
    # 萬一 merge-base 不在 shallow history 裡，繼續加深
    git merge-base origin/${{ github.base_ref }} HEAD \
      || git fetch --deepen=100 origin ${{ github.base_ref }}
```

**邊緣情況**：如果 PR branch 從 main fork 出來已經超過 50 個 commit，
merge-base 就不在 shallow history 裡，`git diff origin/main...HEAD` 會失敗。

```
正常情況（PR 只有幾個 commit）：

  ...──M──A──B──C  (main, depth=50 夠用)
           └──X──Y  (PR branch)

  merge-base = A，在 history 裡 ✓

邊緣情況（PR branch 很長，或從很舊的 commit fork 出來）：

  ...──M  (depth=50 只有這裡)
  [更早的 commit 不存在]──A──B──...──X──Y  (PR branch)

  merge-base 不在 shallow history 裡 → git 報錯 ✗
```

---

## Push to main 不需要任何歷史

注意：**push to main 的 job 根本不需要 `fetch-depth` 設定**。

```yaml
# push to main：直接跑全部 targets
- uses: actions/checkout@v4   # 預設 depth=1，夠了

- run: pants test ::           # 全部跑，不做 diff 計算
```

只有 PR 的 `--changed-since` 才需要 git history 來計算 diff。

---

## 兩個 Job 的 checkout 策略對比

```yaml
# PR job：需要 diff，用 blobless clone
- uses: actions/checkout@v4
  with:
    fetch-depth: 0
    filter: blob:none

# main push job：不需要 diff，預設即可
- uses: actions/checkout@v4
  # 不設任何參數（預設 fetch-depth: 1）
```

---

## 關於 `filter: tree:0`（更激進的選項）

```yaml
filter: tree:0    # Treeless clone：只抓 commit，不抓 tree 和 blob
```

比 `blob:none` 更快，但 `git diff --name-only` 需要 tree 物件（目錄結構）
才能知道哪些路徑有改動。用 `tree:0` 會讓 Pants 的 diff 計算失敗。

```
Pants --changed-since 需要的物件：

  commit objects  ✓  知道有哪些 commit
  tree objects    ✓  知道每個 commit 改了哪些路徑（git diff --name-only）
  blob objects    ✗  不需要檔案的實際內容

→ filter: blob:none 是最適合的選擇
```

---

## 總結

```
你的情境                         → 建議策略
─────────────────────────────────────────────────────────
大部分情況                       → A（fetch-depth: 0 + filter: blob:none）
小 repo（< 100 MB）               → B（fetch-depth: 0，簡單省事）
PR branch 很短（< 20 commits）    → C（fetch-depth: 50，最快）
Push to main                      → 預設 depth=1（不需要 diff）
```
