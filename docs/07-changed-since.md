# 智能 CI：如何判斷哪個專案要編譯

## 核心問題

Monorepo 裡有 100 個 service，我只改了其中 1 個，為什麼要跑 100 個 service 的測試？

Pants 的答案是：**不需要，讓 Git diff + 依賴圖來決定。**

---

## 兩個核心旗標

```
--changed-since=<ref>
    │
    └─ 用 git diff <ref>...HEAD 找出有改動的檔案
       再對應到 Pants 的 targets

--changed-dependents=transitive
    │
    └─ 不只跑直接改動的 target，
       還向上追蹤「哪些 target 依賴這些改動」
       直到依賴圖的頂端
```

---

## 依賴圖 × Git Diff 的運作原理

```
假設 repo 的依賴圖長這樣：

  go/cmd/userapi:bin  ←──┐
                         ├── go/pkg/models:models  ← 改了這裡！
  go/cmd/orderapi:bin ←──┘

  python/services/user_service:bin    ←──┐
  python/services/user_service:tests  ←──┤
                                         ├── python/libs/common:lib  ← 或者改這裡
  python/services/product_service:bin ←──┤
  python/services/product_service:tests──┘

步驟 1：git diff HEAD 找到改動的檔案
  → go/pkg/models/models.go

步驟 2：對應到 Pants target
  → go/pkg/models:models

步驟 3：--changed-dependents=transitive 向上追蹤
  → go/pkg/models:models
  → go/cmd/userapi:userapi （依賴 models）
  → go/cmd/userapi:bin     （依賴 userapi package）
  → go/cmd/orderapi:orderapi
  → go/cmd/orderapi:bin

結論：只有 Go 的 targets 受影響，Python 完全不受影響，不需要跑
```

---

## 三種改動情境的實際輸出

### 情境 A：改 Go 共享 library（`go/pkg/models/models.go`）

```bash
$ pants --changed-since=HEAD list
go/pkg/models:models                    ← 直接改動的 target

$ pants --changed-since=HEAD --changed-dependents=transitive list
go/pkg/models:models
go/cmd/userapi:userapi                  ← 依賴 models
go/cmd/userapi:bin                      ← 依賴 userapi package
go/cmd/orderapi:orderapi
go/cmd/orderapi:bin
```

**Python 服務完全不受影響，不會被觸發。**

---

### 情境 B：改單一 Python 服務（`product_service/app.py`）

```bash
$ pants --changed-since=HEAD --changed-dependents=transitive list
python/services/product_service/app.py:lib
python/services/product_service:bin     ← 同一個 service 的 binary
python/services/product_service:tests   ← 同一個 service 的測試
python/services/product_service:lib
python/services/product_service/app_test.py:tests
```

**user_service 完全不受影響。Go 服務完全不受影響。**

---

### 情境 C：改共享 Python library（`common/models.py`）

```bash
$ pants --changed-since=HEAD --changed-dependents=transitive list
python/libs/common/models.py:lib
python/libs/common:lib
python/libs/common:tests
python/libs/common/models_test.py:tests
python/services/user_service:lib        ← 兩個服務都依賴 common
python/services/user_service:tests
python/services/user_service:bin
python/services/user_service/app.py:lib
python/services/user_service/app_test.py:tests
python/services/product_service:lib
python/services/product_service:tests
python/services/product_service:bin
...（共 14 個 targets）
```

**因為兩個服務都 import `from common.models import ...`，所以都要重跑。**

---

## CI 策略：PR vs Push to Main

```
                          ┌─── PR ────────────────────────────────────┐
                          │                                             │
                          │  --changed-since=origin/main              │
                          │  --changed-dependents=transitive           │
                          │                                             │
                          │  目的：快速回饋，只跑受影響的部分            │
                          │  fetch-depth: 0（需要完整歷史）             │
                          └─────────────────────────────────────────── ┘

                          ┌─── Push to main ────────────────────────── ┐
                          │                                             │
                          │  pants test ::（全部）                      │
                          │  pants check ::                             │
                          │  pants package ::                           │
                          │                                             │
                          │  目的：確認 main 分支永遠是綠的              │
                          │  上傳 artifacts（binaries、PEX）            │
                          └─────────────────────────────────────────── ┘
```

---

## GitHub Actions 關鍵設定解析

```yaml
# 1. fetch-depth: 0 ── 必要！
- uses: actions/checkout@v4
  with:
    fetch-depth: 0   # 預設是 fetch-depth: 1（只取最新一個 commit）
                     # Pants 需要完整歷史才能計算 git diff origin/main...HEAD
                     # 如果只有 1 個 commit，diff 就看不到正確的改動範圍

# 2. --changed-since 的 ref 格式
pants --changed-since=origin/${{ github.base_ref }}
#                              ↑
#                    在 PR 中，base_ref 是 "main"
#                    等同於 git diff origin/main...HEAD
#                    計算從分支 fork 出來後所有的改動

# 3. GitHub Step Summary ── 讓 PR 的 Reviewer 一眼看到影響範圍
pants --changed-since=... list >> $GITHUB_STEP_SUMMARY
# 會在 GitHub Actions 頁面顯示受影響的 targets 清單
```

---

## 為什麼 PR 需要 `fetch-depth: 0`？

```
錯誤情況（fetch-depth: 1，只有最新 commit）：

  git log --oneline
  a1b2c3d  (HEAD) fix: update product service    ← 只有這一個

  git diff origin/main...HEAD
  → 無法比較（缺少 origin/main 的歷史）
  → Pants 報錯或把所有東西都視為 "changed"

正確情況（fetch-depth: 0，完整歷史）：

  git log --oneline
  a1b2c3d  (HEAD) fix: update product service
  f4e5d6c  feat: add product model
  9g8h7i6  (origin/main) chore: initial setup   ← 有共同祖先

  git diff origin/main...HEAD
  → 只看到這個 PR 的改動：product_service/app.py
  → Pants 正確識別受影響範圍
```

---

## 進階：matrix strategy（並行跑不同語言）

```yaml
# 進一步優化：按語言分工並行

jobs:
  detect-changes:
    outputs:
      go_changed: ${{ steps.check.outputs.go }}
      python_changed: ${{ steps.check.outputs.python }}
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }

      - id: check
        run: |
          GO=$(pants --changed-since=origin/main --changed-dependents=transitive list 2>/dev/null | grep '^go/' | wc -l)
          PY=$(pants --changed-since=origin/main --changed-dependents=transitive list 2>/dev/null | grep '^python/' | wc -l)
          echo "go=$([ $GO -gt 0 ] && echo true || echo false)" >> $GITHUB_OUTPUT
          echo "python=$([ $PY -gt 0 ] && echo true || echo false)" >> $GITHUB_OUTPUT

  test-go:
    needs: detect-changes
    if: needs.detect-changes.outputs.go_changed == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - run: pants --changed-since=origin/main --changed-dependents=transitive test
        env:
          PANTS_FILTER: "go/"   # 只跑 Go 相關

  test-python:
    needs: detect-changes
    if: needs.detect-changes.outputs.python_changed == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - run: pants --changed-since=origin/main --changed-dependents=transitive test
        env:
          PANTS_FILTER: "python/"
```

---

## 指令速查

```bash
# 查看有哪些檔案改動
git diff --name-only origin/main

# 查看改動影響的 Pants targets（直接）
pants --changed-since=origin/main list

# 查看改動影響的 Pants targets（含依賴者）
pants --changed-since=origin/main --changed-dependents=transitive list

# 只跑受影響的測試
pants --changed-since=origin/main --changed-dependents=transitive test

# 只 check 受影響的
pants --changed-since=origin/main --changed-dependents=transitive check

# 只打包受影響的 binary
pants --changed-since=origin/main --changed-dependents=transitive package

# HEAD = 未 commit 的改動
pants --changed-since=HEAD --changed-dependents=transitive test
```
