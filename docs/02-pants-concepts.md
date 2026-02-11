# Pants Build System 核心概念

## Pants 是什麼？

Pants 是專為 Monorepo 設計的建置系統。它的核心理念是：

> **「告訴 Pants 你的程式碼在哪裡，它會自動推斷依賴關係，只建置你需要的部分。」**

---

## 核心組成

### 1. `pants.toml` — 全局設定檔

```toml
[GLOBAL]
pants_version = "2.30.0"          # 指定 Pants 版本（確保所有人用同一版本）
backend_packages = [
  "pants.backend.experimental.go", # 啟用 Go 支援
  "pants.backend.python",           # 啟用 Python 支援
  "pants.backend.python.lint.black",
  "pants.backend.python.typecheck.mypy",
]

[source]
root_patterns = [                  # 告訴 Pants Python source root 的位置
  "/python/libs",                  # → 讓 `from common.models import ...` 可以運作
  "/python/services",              # → 讓 `from user_service.app import ...` 可以運作
]

[python]
interpreter_constraints = ["==3.11.*"]  # 指定 Python 版本
enable_resolves = true                   # 啟用 lockfile 機制
resolves = { python-default = "python-default.lock" }

[golang]
minimum_expected_version = "1.25"  # Pants 需找到 >= 此版本的 Go
```

---

### 2. `BUILD` 檔案 — Target 宣告

`BUILD` 是 Pants 的建置描述檔，放在每個需要被 Pants 管理的目錄中。

**概念：Target（目標）**

Target 是 Pants 的最小建置單位，代表「一組檔案 + 描述它的 metadata」。

```python
# BUILD 檔案使用 Starlark 語法（Python 的子集）

# Go 相關 targets
go_mod()            # 宣告 Go module（對應 go.mod）
go_package()        # 宣告一個 Go package（對應一個目錄的 .go 檔）
go_binary()         # 宣告一個可執行的 Go binary（main package）

# Python 相關 targets
python_sources()    # 宣告 Python 原始碼（非測試）
python_tests()      # 宣告 Python 測試檔案（pytest）
python_requirements() # 從 requirements.txt 產生依賴 targets
pex_binary()        # 宣告可打包成 .pex 的 Python 應用
```

**Target 命名：`目錄路徑:target名稱`**

```
go/cmd/userapi:bin         → go/cmd/userapi/BUILD 中的 go_binary(name="bin")
go/pkg/httputil:httputil   → 自動推斷的 target 名稱
python/services/user_service:lib   → python_sources(name="lib")
python/services/user_service:tests → python_tests(name="tests")
```

---

### 3. 依賴推斷（Dependency Inference）

這是 Pants 最強大的功能之一：**自動分析 import 語句，推斷 target 之間的依賴**。

```go
// go/cmd/userapi/main.go
import (
    "github.com/example/monorepo-demo/go/pkg/httputil"  // ← Pants 自動發現依賴
    "github.com/example/monorepo-demo/go/pkg/models"
    "github.com/google/uuid"                              // ← 從 go.mod 的第三方依賴
)
```

```python
# python/services/user_service/app.py
from common.models import User, UserBase  # ← Pants 推斷依賴 python/libs/common:lib
from fastapi import FastAPI               # ← 推斷依賴 requirements.txt 中的 fastapi
```

不需要在 BUILD 中手動寫 `dependencies = [...]`，Pants 會自動處理！

---

### 4. 快取機制

Pants 對每個 target 建立 content-addressed cache：

```
第一次執行：
  pants test go/pkg/httputil:httputil   → 實際執行測試（~2秒）

第二次執行（程式碼沒有改變）：
  pants test go/pkg/httputil:httputil   → 直接讀取快取（<0.1秒）
  ✓ go/pkg/httputil:httputil succeeded in 0.02s (memoized)  ← 注意 "memoized"
```

---

### 5. Source Roots（Python 特有）

Python 的 import 路徑依賴於 `sys.path`。Pants 需要知道哪些目錄是「根目錄」。

```toml
# pants.toml
[source]
root_patterns = [
  "/python/libs",    # python/libs/common/ → import 為 common.models
  "/python/services" # python/services/user_service/ → import 為 user_service.app
]
```

```
python/libs/common/models.py
         ↑
    source root = python/libs
         ↓
from common.models import User  ✓
```

---

### 6. Lockfile 與 Resolve

Python 的依賴管理使用 **resolve + lockfile** 機制：

```
requirements.txt  →  pants generate-lockfiles  →  python-default.lock
（鬆散版本約束）         （解析精確版本）              （精確版本 + hash）
  fastapi==0.115.6                                  fastapi 0.115.6
  pydantic==2.10.4                                  pydantic 2.10.4
                                                    starlette 0.41.3
                                                    ... （所有遞迴依賴）
```

這確保所有開發者和 CI 都使用完全相同的套件版本。

---

## Pants 指令參考

```bash
# ── 測試 ──────────────────────────────────────────────
pants test ::                          # 跑所有測試（Go + Python）
pants test go/::                       # 只跑 Go 測試
pants test python/::                   # 只跑 Python 測試
pants test go/pkg/httputil:httputil    # 跑特定 target

# ── 建置 ──────────────────────────────────────────────
pants package ::                       # 打包所有 binary
pants package go/cmd/userapi:bin       # 打包 Go binary → dist/
pants package python/services/user_service:bin  # 打包 PEX → dist/

# ── 執行 ──────────────────────────────────────────────
pants run go/cmd/userapi:bin           # 執行 Go User API
pants run python/services/user_service:bin  # 執行 Python User Service

# ── 檢查 ──────────────────────────────────────────────
pants check ::                         # Go 編譯檢查 + Python mypy

# ── 分析 ──────────────────────────────────────────────
pants dependencies go/cmd/userapi:     # 查看 userapi 的依賴
pants dependees go/pkg/models:         # 查看哪些 target 依賴 models
pants --changed-since=HEAD test        # 只測試有改動的程式碼

# ── 依賴管理 ─────────────────────────────────────────
pants generate-lockfiles               # 更新 Python lockfile
pants tailor ::                        # 自動產生缺少的 BUILD 檔案
```
