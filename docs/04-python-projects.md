# Python 專案詳解

## 目錄結構

```
python/
├── services/                ← 可部署的 FastAPI 應用
│   ├── user_service/        ← Python 專案 1
│   │   ├── BUILD
│   │   ├── __init__.py
│   │   ├── app.py
│   │   └── app_test.py
│   └── product_service/     ← Python 專案 2
│       ├── BUILD
│       ├── __init__.py
│       ├── app.py
│       └── app_test.py
└── libs/                    ← 可複用的 library
    └── common/              ← 共享 Pydantic 模型
        ├── BUILD
        ├── __init__.py
        ├── models.py
        └── models_test.py
```

---

## Source Roots 設定（關鍵！）

Pants 需要知道 Python import 的根目錄，才能正確解析模組路徑。

```toml
# pants.toml
[source]
root_patterns = [
  "/python/libs",    # common.models → python/libs/common/models.py
  "/python/services" # user_service.app → python/services/user_service/app.py
]
```

**為什麼需要這個設定？**

```
python/libs/common/models.py 的 import 路徑：

  如果 root = /python/libs  → from common.models import User      ✓
  如果 root = /python       → from libs.common.models import User ✗（我們不想要）
  如果 root = /             → from python.libs.common.models ...  ✗（更不對）
```

---

## 共享 Library：common

### `python/libs/common/BUILD`
```python
python_sources(name="lib")    # 宣告所有 .py 檔（非測試）為 library
python_tests(name="tests")    # 宣告所有 _test.py 檔為測試
```

### `python/libs/common/models.py`
```python
from pydantic import BaseModel
from datetime import datetime

class UserBase(BaseModel):
    name: str
    email: str

class User(UserBase):
    id: str
    created_at: datetime

class ProductBase(BaseModel):
    name: str
    price: float
    description: str = ""

class Product(ProductBase):
    id: str
    created_at: datetime

class HealthResponse(BaseModel):
    status: str
    service: str
```

使用 **Pydantic v2** 的好處：
- 自動 JSON 序列化/反序列化
- 型別驗證（FastAPI 整合）
- 清晰的資料合約，兩個服務共用同一套定義

---

## Python 專案 1：User Service

### `python/services/user_service/BUILD`
```python
python_sources(name="lib")

python_tests(
    name="tests",
    dependencies=["//:reqs#httpx"],  # 明確聲明 httpx（FastAPI TestClient 的隱含依賴）
)

pex_binary(
    name="bin",
    entry_point="app.py",   # 打包後執行的入口檔案
)
```

> **為什麼 tests 需要手動加 httpx 依賴？**
>
> Pants 依賴推斷是靠分析 `import` 語句。`app_test.py` 寫的是：
> ```python
> from fastapi.testclient import TestClient
> ```
> Pants 推斷出 `fastapi` 依賴，但 `TestClient` 在執行時需要 `httpx`（透過
> `starlette.testclient`），這個隱含依賴 Pants 無法靜態推斷，所以要手動宣告。
>
> `//:reqs#httpx` 的解讀：
> - `//:reqs` → 根目錄 BUILD 中的 `python_requirements(name="reqs")`
> - `#httpx`  → 對應 requirements.txt 中 `httpx==0.28.1`

### `python/services/user_service/app.py`
```python
from fastapi import FastAPI, HTTPException
from common.models import HealthResponse, User, UserBase  # ← 引用共享 library

app = FastAPI(title="User Service", version="0.1.0")

_users: dict[str, User] = {}  # 簡單的記憶體儲存（示範用）

@app.get("/health", response_model=HealthResponse)
def health():
    return HealthResponse(status="ok", service="user-service")

@app.get("/users", response_model=list[User])
def list_users():
    return list(_users.values())

@app.post("/users", response_model=User, status_code=201)
def create_user(payload: UserBase):
    # FastAPI 自動：
    # 1. 解析 JSON body → UserBase model
    # 2. 驗證欄位型別
    # 3. 序列化回應 → JSON
    user = User(id=str(uuid4()), **payload.model_dump(), created_at=datetime.now())
    _users[user.id] = user
    return user
```

### `python/services/user_service/app_test.py`
```python
from fastapi.testclient import TestClient
from user_service.app import app  # ← 注意：用 user_service.app（source root 是 /python/services）

client = TestClient(app)

def test_create_and_list_users():
    resp = client.post("/users", json={"name": "Alice", "email": "alice@example.com"})
    assert resp.status_code == 201

    resp = client.get("/users")
    assert len(resp.json()) >= 1
```

---

## Python 專案 2：Product Service

結構與 User Service 相同，端點為 `/products`，監聽預設 port。

```bash
pants run python/services/product_service:bin
```

API 端點：
- `GET  /health`       → 健康檢查
- `GET  /products`     → 列出所有商品
- `GET  /products/{id}`→ 取得特定商品
- `POST /products`     → 建立新商品

---

## 依賴關係圖

```
python/services/user_service:tests
    ├── python/services/user_service:lib        ← 自動推斷（同目錄）
    │   ├── python/libs/common:lib              ← 自動推斷（from common.models import）
    │   │   └── //:reqs#pydantic               ← 自動推斷（from pydantic import）
    │   └── //:reqs#fastapi                    ← 自動推斷（from fastapi import）
    └── //:reqs#httpx                           ← 手動宣告（隱含依賴）

python/services/product_service:tests
    └── （結構相同）

python/libs/common:tests
    ├── python/libs/common:lib
    └── //:reqs#pydantic
```

---

## PEX — Python EXecutable

`pex_binary` 將 Python 應用打包成一個**自包含的可執行檔**（類似 Go 的 static binary）：

```bash
pants package python/services/user_service:bin
# → dist/python.services.user_service/bin.pex

# 直接執行（不需要 virtualenv）：
./dist/python.services.user_service/bin.pex
```

PEX 內含：
- Python 原始碼
- 所有依賴的 wheel 檔案
- 一個自解壓的啟動器

---

## Python 相關 Pants 設定重點

```toml
# pants.toml
[python]
interpreter_constraints = ["==3.11.*"]  # 指定支援的 Python 版本
enable_resolves = true                   # 啟用 lockfile 機制（強烈建議）
resolves = {                             # 定義 resolve 名稱 → lockfile 路徑
  python-default = "python-default.lock"
}
# 可定義多個 resolve（例如：不同服務用不同依賴集）

[python-bootstrap]
search_path = ["<PATH>"]   # Pants 搜尋 Python 解譯器的路徑
# <PYENV> 可加入 pyenv 管理的版本
```

### 更新依賴的流程

```bash
# 1. 修改 requirements.txt（新增或更新套件）
echo "httpx==0.29.0" >> requirements.txt

# 2. 重新產生 lockfile
pants generate-lockfiles

# 3. 提交 requirements.txt 和 python-default.lock
git add requirements.txt python-default.lock
git commit -m "chore: update httpx to 0.29.0"
```
