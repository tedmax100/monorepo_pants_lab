# Monorepo 概述

## 什麼是 Monorepo？

**Monorepo**（Monolithic Repository）是將多個專案放在同一個版本控制倉庫中的做法，與之相對的是每個專案獨立一個 repo（Polyrepo）。

```
Polyrepo 方式：                     Monorepo 方式：
  repo-user-api/                      monorepo_demo/
  repo-order-api/         vs.         ├── go/cmd/userapi/
  repo-user-service/                  ├── go/cmd/orderapi/
  repo-product-service/               ├── python/services/user_service/
                                      └── python/services/product_service/
```

### Monorepo 的優點

| 優點 | 說明 |
|------|------|
| **統一的依賴管理** | 所有專案共用同一份依賴版本，避免版本不一致 |
| **跨專案重構** | 修改共享 library 時可以同時看到所有受影響的地方 |
| **原子性提交** | 一個 commit 可以同時修改前後端、多個服務 |
| **共享程式碼** | 公共 library 直接引用，不需要發布成獨立套件 |
| **統一 CI/CD** | 一套工具鏈管理所有語言和專案 |

### Monorepo 的挑戰 → Pants 的解決方案

| 挑戰 | Pants 如何解決 |
|------|---------------|
| 規模變大後建置變慢 | 細粒度快取 + 只重建有改動的部分 |
| 多語言工具鏈複雜 | 統一的 `pants test/build/lint` 指令 |
| 不同專案互相干擾 | 每個 target 在隔離的 sandbox 中執行 |
| 難以確定影響範圍 | `--changed-since` 只跑受影響的測試 |

---

## 本 Repo 的整體架構

```
monorepo_demo/
│
├── pants.toml              ← Pants 全局設定（版本、啟用後端）
├── BUILD                   ← 根目錄 BUILD：宣告 Go module、Python 依賴
├── go.mod / go.sum         ← Go 模組管理
├── requirements.txt        ← Python 依賴清單
├── python-default.lock     ← Pants 產生的 Python lockfile（精確版本）
│
├── go/                     ← 所有 Go 程式碼
│   ├── cmd/                ← 可執行的 binary（main package）
│   │   ├── userapi/        ← Go 專案 1：User HTTP API
│   │   └── orderapi/       ← Go 專案 2：Order HTTP API
│   └── pkg/                ← 可被複用的 library
│       ├── httputil/       ← 共享：JSON 回應工具
│       └── models/         ← 共享：資料結構定義
│
└── python/                 ← 所有 Python 程式碼
    ├── services/           ← 可部署的 FastAPI 應用
    │   ├── user_service/   ← Python 專案 1：User FastAPI Service
    │   └── product_service/← Python 專案 2：Product FastAPI Service
    └── libs/               ← 可被複用的 library
        └── common/         ← 共享：Pydantic 資料模型
```
