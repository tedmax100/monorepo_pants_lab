# BUILD 檔案完整解析

每個目錄的 `BUILD` 檔案是 Pants 理解這個 repo 結構的關鍵。

---

## 根目錄 `BUILD`

```python
# BUILD（根目錄）

# Go module target：讀取 go.mod，讓 Pants 理解整個 Go module
# 會自動為 go.mod 中每個 require 建立 go_third_party_package target
go_mod(name="mod")

# Python 依賴 target：讀取 requirements.txt
# 為每一行建立一個 python_requirement target
# 引用方式：//:reqs#fastapi、//:reqs#pydantic 等
python_requirements(name="reqs")
```

---

## Go BUILD 檔案

### `go/pkg/httputil/BUILD` — Library Package
```python
go_package()
# ↑ 最簡單的 BUILD，告訴 Pants：
# - 這個目錄的所有 .go 檔組成一個 package
# - package 名稱從 .go 檔的 package 宣告取得
# - 測試檔案（_test.go）自動歸類為測試
# - 依賴關係從 import 語句自動推斷
```

### `go/pkg/models/BUILD` — Library Package
```python
go_package()   # 同上
```

### `go/cmd/userapi/BUILD` — Executable Binary
```python
go_package()   # 宣告 package（包含 main.go）

go_binary(
    name="bin",
    # 可選欄位：
    # output_path = "userapi",    # 輸出 binary 的名稱（預設用目錄名）
    # dependencies = [...],        # 通常不需要，靠依賴推斷
)
# 有了 go_binary，才能執行：
#   pants run go/cmd/userapi:bin
#   pants package go/cmd/userapi:bin
```

### `go/cmd/orderapi/BUILD` — Executable Binary
```python
go_package()
go_binary(name="bin")
```

---

## Python BUILD 檔案

### `python/libs/common/BUILD` — Shared Library
```python
python_sources(name="lib")
# ↑ 宣告這個目錄的非測試 .py 檔為 library
# 其他 target 可以 import 這裡的程式碼

python_tests(name="tests")
# ↑ 宣告測試檔案（*_test.py）
# 自動設定 pytest runner
```

### `python/services/user_service/BUILD` — FastAPI Service
```python
python_sources(name="lib")
# 宣告 app.py、__init__.py 等原始碼

python_tests(
    name="tests",
    dependencies=["//:reqs#httpx"],
    # ↑ 手動宣告 httpx 依賴
    # 路徑解析：
    #   //       = repo 根目錄
    #   :reqs    = 根目錄 BUILD 中 python_requirements(name="reqs")
    #   #httpx   = requirements.txt 中的 httpx 套件
)

pex_binary(
    name="bin",
    entry_point="app.py",
    # ↑ 指定打包後的入口點
    # 完整路徑：user_service/app.py 中的 __main__
    # pants package 後可直接執行 ./dist/...bin.pex
)
```

### `python/services/product_service/BUILD` — FastAPI Service
```python
python_sources(name="lib")

python_tests(
    name="tests",
    dependencies=["//:reqs#httpx"],  # 同樣需要手動宣告
)

pex_binary(
    name="bin",
    entry_point="app.py",
)
```

---

## Target 地址語法總結

```
[//][路徑][:target名稱][#生成的target]

//                    → repo 根目錄（絕對路徑）
go/cmd/userapi        → 相對路徑
:bin                  → target 名稱
:                     → 目錄中的所有 target（省略名稱）
::                    → 這個目錄及所有子目錄的所有 target

範例：
  go/cmd/userapi:bin          → userapi 目錄中 name="bin" 的 target
  go/cmd/userapi:             → userapi 目錄的所有 target
  go/::                       → go/ 目錄下的所有 target（遞迴）
  ::                          → 整個 repo 的所有 target
  //:reqs#fastapi             → 根目錄 reqs 中的 fastapi 套件 target
  //:mod#github.com/google/uuid → Go module 中的 uuid 套件 target
```

---

## `pants tailor` — 自動產生 BUILD 檔案

如果你新增了 Go 或 Python 檔案，可以用 `pants tailor ::` 自動產生對應的 BUILD：

```bash
# 範例：新增一個 Go 檔案後
touch go/pkg/newutil/util.go
pants tailor ::
# → 自動建立 go/pkg/newutil/BUILD 內含 go_package()
```

這讓你不需要手動寫 BUILD 就能快速起步。

---

## 常見問題：BUILD 檔案的位置

```
規則：每個需要被 Pants 獨立追蹤的「目錄」都需要一個 BUILD 檔案

✓ 正確：
  go/pkg/httputil/BUILD  （一個 package，一個 BUILD）
  go/pkg/models/BUILD    （獨立的 package）

✗ 不需要：
  go/pkg/BUILD           （pkg/ 目錄本身不是一個 package）
  go/BUILD               （go/ 目錄本身不是一個 package）
```
