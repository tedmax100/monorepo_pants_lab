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

### `go/cmd/userapi/BUILD` — Executable Binary + Docker Image
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

docker_image(
    name="docker",
    repository="ghcr.io/{build_args.GITHUB_REPOSITORY_OWNER}/userapi",
    image_tags=["{build_args.IMAGE_TAG}"],
    registries=["@ghcr"],
    dependencies=[":bin"],
    # {build_args.XXX}：執行期才展開的佔位符
    # pants.toml [docker] build_args 有預設值（local / latest），
    # CI 用 --docker-build-args 覆寫為 github.repository_owner / github.sha
    # pants publish go/cmd/userapi:docker → build + push to ghcr.io
)
```

### `go/cmd/orderapi/BUILD` — Executable Binary + Docker Image
```python
go_package()
go_binary(name="bin")
docker_image(
    name="docker",
    repository="ghcr.io/{build_args.GITHUB_REPOSITORY_OWNER}/orderapi",
    image_tags=["{build_args.IMAGE_TAG}"],
    registries=["@ghcr"],
    dependencies=[":bin"],
)
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

### `python/services/user_service/BUILD` — FastAPI Service + Docker Image
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
    # ↑ 用於本地直接執行：./dist/...bin.pex
    # app.py 本身沒有呼叫 uvicorn.run()，不適合 Docker 入口點
)

pex_binary(
    name="server",
    entry_point="uvicorn",
    dependencies=[":lib", "//:reqs#uvicorn"],
    # ↑ 用於 Docker image
    # entry_point="uvicorn" → PEX 執行時呼叫 uvicorn CLI
    # Dockerfile 的 CMD 傳入 "user_service.app:app --host 0.0.0.0 --port 8080"
    # 打包後放在 dist/python.services.user_service/server.pex
)

docker_image(
    name="docker",
    repository="ghcr.io/{build_args.GITHUB_REPOSITORY_OWNER}/user-service",
    image_tags=["{build_args.IMAGE_TAG}"],
    registries=["@ghcr"],
    dependencies=[":server"],
    # Dockerfile 裡 COPY python.services.user_service/server.pex /app/server.pex
    # 注意路徑格式：dist/ 下的 dotted path（斜線 → 點）
)
```

### `python/services/product_service/BUILD` — FastAPI Service + Docker Image
```python
python_sources(name="lib")

python_tests(
    name="tests",
    dependencies=["//:reqs#httpx"],
)

pex_binary(name="bin", entry_point="app.py")

pex_binary(
    name="server",
    entry_point="uvicorn",
    dependencies=[":lib", "//:reqs#uvicorn"],
)

docker_image(
    name="docker",
    repository="ghcr.io/{build_args.GITHUB_REPOSITORY_OWNER}/product-service",
    image_tags=["{build_args.IMAGE_TAG}"],
    registries=["@ghcr"],
    dependencies=[":server"],
)
```

---

## `docker_image` 關鍵知識

### `{build_args.XXX}` 插值

`docker_image` 的 `repository` 和 `image_tags` 欄位支援 `{build_args.XXX}` 佔位符，**不支援** `{env.XXX}`。

```toml
# pants.toml — 提供預設值，讓本地 build 不會出錯
[docker]
registries = { ghcr = { address = "ghcr.io" } }
build_args = [
    "GITHUB_REPOSITORY_OWNER=local",   # 預設值
    "IMAGE_TAG=latest",
]
env_vars = ["DOCKER_HOST"]   # 讓 Colima / 遠端 Docker 的 socket 能被找到
```

CI 用 `--docker-build-args` 覆寫（注意是複數）：
```bash
pants \
  --docker-build-args="GITHUB_REPOSITORY_OWNER=${{ github.repository_owner }}" \
  --docker-build-args="IMAGE_TAG=${{ github.sha }}" \
  publish go/cmd/userapi:docker
```

### Dockerfile 的 COPY 路徑格式

Pants 把 artifact 放在 `dist/` 下，路徑以 **點** 分隔（對應目錄層級）：

```
dist/
  go.cmd.userapi/bin            ← go/cmd/userapi:bin 的輸出
  go.cmd.orderapi/bin
  python.services.user_service/server.pex    ← pex_binary(name="server")
  python.services.product_service/server.pex
```

Dockerfile 的 `COPY` 路徑等於去掉 `dist/` 前綴：
```dockerfile
# Go 服務
COPY go.cmd.userapi/bin /usr/local/bin/userapi

# Python 服務
COPY python.services.user_service/server.pex /app/server.pex
```

### `pants publish` vs `pants package`

| 指令 | 效果 |
|------|------|
| `pants package go/cmd/userapi:docker` | 只在本地 build image（不 push） |
| `pants publish go/cmd/userapi:docker` | build + push 到 registries |

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
