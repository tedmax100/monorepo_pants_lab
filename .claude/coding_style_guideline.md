# Coding Style Guideline

此檔案定義 Code Review 各等級的判斷標準。
由團隊維護，供 `/review-pr` skill 自動載入。

---

## MUST（嚴重問題）— 必須修正才能合併

以下情況一律標記為 MUST：

### 正確性
- 邏輯錯誤、會導致 runtime 崩潰的 bug
- 資料庫操作缺少 transaction 且有部分更新風險
- 非同步競態條件（race condition）

### 安全性
- SQL injection、XSS、SSRF 等 OWASP Top 10 漏洞
- 明文儲存密碼或 secret
- 未驗證外部輸入即使用（user input directly used in query/command）

### 資料完整性
- 生產環境 seed data 含測試用名稱或假資料（如 `Bob_test`、`test123`）
- 遷移腳本可能造成不可逆的資料遺失

---

## SHOULD（需要改進）— 強烈建議修正

以下情況標記為 SHOULD：

### 可維護性
- 函數超過 50 行且沒有拆分的必要性說明
- 魔術數字（magic number）未定義為常數
- 公開 API 缺少錯誤處理

### CI/CD
- GitHub Actions YAML 中變數未正確 export 到 `$GITHUB_OUTPUT`
- Shell 腳本使用未定義的變數（`$UNDEFINED_VAR`）
- `[skip ci]` 機制缺失導致可能的無限循環觸發

### 測試
- 新增功能沒有對應的測試案例
- 測試只測 happy path，缺少 error case

---

## MAY（建議優化）— 可選改進

以下情況標記為 MAY：

### 風格
- 命名不夠語意化（可理解但不夠清晰）
- 重複邏輯可提取為 helper（但不影響功能）
- 文件範例程式碼與實際 API 有細微差距

### 效能
- 可以用更簡潔的寫法達到同樣效果
- N+1 query（低流量服務可降為 MAY）

---

## 不予評論

以下情況**略過**，不發 review comment：
- 純格式調整（空白、縮排），若專案已有 linter 處理
- 版本號 bump
- 自動生成的檔案（lock files、generated code）
- `[skip ci]` commit message 本身
