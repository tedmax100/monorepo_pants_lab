# Go 專案詳解

## 目錄結構

```
go/
├── cmd/                  ← 約定：可執行的應用放在 cmd/
│   ├── userapi/          ← Go 專案 1
│   │   ├── BUILD
│   │   └── main.go
│   └── orderapi/         ← Go 專案 2
│       ├── BUILD
│       └── main.go
└── pkg/                  ← 約定：可複用的 library 放在 pkg/
    ├── httputil/         ← 共享工具
    │   ├── BUILD
    │   ├── response.go
    │   └── response_test.go
    └── models/           ← 共享資料模型
        ├── BUILD
        ├── models.go
        └── models_test.go
```

> `cmd/` 和 `pkg/` 是 Go 社群的慣例，不是 Pants 的要求。

---

## 根目錄 BUILD（Go Module 宣告）

```python
# BUILD（根目錄）
go_mod(name="mod")          # 告訴 Pants：這裡有 go.mod
python_requirements(name="reqs")
```

`go_mod` 讓 Pants 能：
1. 讀取 `go.mod` 知道 module 路徑（`github.com/example/monorepo-demo`）
2. 從 `go.sum` 下載第三方依賴
3. 為每個第三方套件自動建立 `go_third_party_package` target

---

## 共享 Library：httputil

### `go/pkg/httputil/BUILD`
```python
go_package()   # 宣告這個目錄是一個 Go package，就這樣！
```

### `go/pkg/httputil/response.go`
```go
package httputil

import (
    "encoding/json"
    "net/http"
)

// JSONResponse 寫入 JSON 格式的 HTTP 回應
func JSONResponse(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

// ErrorResponse 寫入 JSON 格式的錯誤回應
func ErrorResponse(w http.ResponseWriter, status int, message string) {
    JSONResponse(w, status, map[string]string{"error": message})
}
```

**這個 package 被兩個 binary 共享**，Pants 透過 import 路徑自動推斷依賴：

```go
// go/cmd/userapi/main.go 和 go/cmd/orderapi/main.go 都有：
import "github.com/example/monorepo-demo/go/pkg/httputil"
//                                          ↑
//                             Pants 看到這個 import，自動連結到
//                             go/pkg/httputil:httputil target
```

---

## 共享 Library：models

### `go/pkg/models/models.go`
```go
package models

import "time"

type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type Order struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Product   string    `json:"product"`
    Quantity  int       `json:"quantity"`
    Total     float64   `json:"total"`
    CreatedAt time.Time `json:"created_at"`
}
```

---

## Go 專案 1：User API

### `go/cmd/userapi/BUILD`
```python
go_package()        # 宣告 package（含 main.go）

go_binary(
    name="bin",     # 宣告可執行 binary
    # output_path="userapi",  # 可選：指定輸出的 binary 名稱
)
```

### `go/cmd/userapi/main.go` 重點
```go
package main

import (
    "github.com/example/monorepo-demo/go/pkg/httputil"  // 共享工具
    "github.com/example/monorepo-demo/go/pkg/models"    // 共享模型
    "github.com/google/uuid"                              // 第三方依賴
)

// 路由
// GET  /users  → 列出所有用戶
// POST /users  → 建立新用戶

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /users", handleListUsers)
    mux.HandleFunc("POST /users", handleCreateUser)
    http.ListenAndServe(":8081", mux)
}
```

### 執行與打包
```bash
pants run go/cmd/userapi:bin          # 直接執行（開發用）
pants package go/cmd/userapi:bin      # 打包成 binary → dist/go.cmd.userapi/bin
```

---

## Go 專案 2：Order API

與 User API 結構相同，監聽 `:8082`，管理訂單資料。

```bash
pants run go/cmd/orderapi:bin
pants package go/cmd/orderapi:bin
```

---

## 依賴關係圖

```
go/cmd/userapi:bin
    ├── go/cmd/userapi:userapi (go_package)
    │   ├── go/pkg/httputil:httputil  ← 共享
    │   ├── go/pkg/models:models      ← 共享
    │   └── [//:mod]#github.com/google/uuid  ← 第三方
    └── （Pants 自動推斷，不需手動寫 dependencies）

go/cmd/orderapi:bin
    ├── go/cmd/orderapi:orderapi (go_package)
    │   ├── go/pkg/httputil:httputil  ← 同一份共享 library
    │   ├── go/pkg/models:models      ← 同一份共享 library
    │   └── [//:mod]#github.com/google/uuid
    └── ...
```

---

## Go 相關 Pants 設定重點

```toml
# pants.toml
[GLOBAL]
backend_packages = [
  "pants.backend.experimental.go",  # ← "experimental" 代表 Go 支援仍在積極開發中
]

[golang]
minimum_expected_version = "1.25"   # Pants 會在 PATH 中搜尋 >= 1.25 的 Go
# go_search_paths = ["/usr/local/go/bin"]  # 可選：指定 Go 的搜尋路徑
```

> **注意**：`pants.backend.experimental.go` 中的 "experimental" 表示 Go 支援功能
> 仍在持續完善，但已可在生產環境使用，pantsbuild 官方也有完整的範例 repo。

---

## Go + Pants 的特殊行為

### 第三方套件的引用
Go 的第三方套件不需要在 BUILD 中手動宣告。`go_mod` target 會：
1. 讀取 `go.mod` 中的 `require` 清單
2. 自動為每個套件產生 `go_third_party_package` target
3. 透過 import 路徑自動匹配

### 不需要 `go mod download`（但需要 `go mod tidy`）
Pants 會自己下載依賴，但 `go.mod` 和 `go.sum` 需要手動保持更新：
```bash
go mod tidy   # 更新 go.mod 和 go.sum
```
