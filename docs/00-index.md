# Monorepo + Pants 示範專案文件

## 文件索引

| 文件 | 說明 |
|------|------|
| [01-overview.md](./01-overview.md) | Monorepo 概念、整體架構 |
| [02-pants-concepts.md](./02-pants-concepts.md) | Pants 核心概念：pants.toml、BUILD、依賴推斷、快取 |
| [03-go-projects.md](./03-go-projects.md) | Go 專案詳解：userapi、orderapi、共享 library |
| [04-python-projects.md](./04-python-projects.md) | Python 專案詳解：FastAPI services、Pydantic models |
| [05-build-files-explained.md](./05-build-files-explained.md) | 每個 BUILD 檔案的完整解析 |
| [06-workflow.md](./06-workflow.md) | 日常開發流程、CI/CD 整合 |

---

## 快速開始

```bash
# 安裝依賴（只需一次）
pants generate-lockfiles

# 跑所有測試
pants test ::

# 打包所有 binary
pants package ::
```

## 專案結構一覽

```
monorepo_demo/
├── pants.toml              ← Pants 2.30 設定
├── BUILD                   ← go_mod + python_requirements
├── go.mod / go.sum         ← Go 模組
├── requirements.txt        ← Python 依賴
├── python-default.lock     ← Python 精確版本鎖定
│
├── go/
│   ├── cmd/
│   │   ├── userapi/        ← [Go 專案 1] User HTTP API（:8081）
│   │   └── orderapi/       ← [Go 專案 2] Order HTTP API（:8082）
│   └── pkg/
│       ├── httputil/       ← 共享：JSON response 工具
│       └── models/         ← 共享：User、Order 資料結構
│
└── python/
    ├── services/
    │   ├── user_service/   ← [Python 專案 1] FastAPI User Service
    │   └── product_service/← [Python 專案 2] FastAPI Product Service
    └── libs/
        └── common/         ← 共享：Pydantic 資料模型
```

## 技術棧

| 語言 | 框架/工具 | 用途 |
|------|-----------|------|
| Go 1.25 | 標準 `net/http` | User API、Order API |
| Python 3.11 | FastAPI + Pydantic v2 | User Service、Product Service |
| — | Pants 2.30 | 統一建置、測試、打包 |
| — | uv（透過 Pants） | Python 依賴解析 |
