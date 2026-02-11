# 開發工作流程

## 初次設定

```bash
# 1. 確認 Pants 已安裝
pants --version   # 輸出：2.30.0

# 2. 確認 Go 版本
go version        # 需要 >= 1.25

# 3. 確認 Python 版本
python3 --version # 需要 3.11.x

# 4. 產生 Python lockfile（首次需要）
pants generate-lockfiles
# → 產生 python-default.lock

# 5. 確認所有 target 都能被找到
pants list ::     # 列出所有 targets
```

---

## 日常開發流程

### 修改 Go 程式碼後

```bash
# 只測試修改的部分（利用 Git diff）
pants --changed-since=HEAD test

# 或指定 target 測試
pants test go/pkg/models:models

# 驗證能編譯
pants check go/::

# 打包確認
pants package go/cmd/userapi:bin
ls dist/go.cmd.userapi/bin
```

### 修改 Python 程式碼後

```bash
# 只測試修改的部分
pants --changed-since=HEAD test

# 型別檢查
pants check python/::

# 執行特定服務測試
pants test python/services/user_service:tests

# 打包 PEX
pants package python/services/user_service:bin
./dist/python.services.user_service/bin.pex
```

### 修改共享 Library 後

```bash
# 查看哪些 target 依賴這個 library
pants dependees python/libs/common:lib
# → python/services/user_service:lib
# → python/services/user_service:tests
# → python/services/product_service:lib
# → python/services/product_service:tests

pants dependees go/pkg/models:models
# → go/cmd/userapi:userapi
# → go/cmd/orderapi:orderapi

# 跑所有受影響的測試
pants --changed-since=HEAD test
# Pants 會自動找出所有受影響的 target 並只跑那些
```

---

## 測試結果說明

```
pants test ::

✓ go/pkg/httputil:httputil succeeded in 0.02s (memoized)
          ↑                              ↑         ↑
    target 地址               執行時間    從快取取得（沒有重新執行）

✓ go/pkg/models:models succeeded in 0.02s (memoized)
✓ python/libs/common/models_test.py:tests succeeded in 0.72s (memoized)
✓ python/services/product_service/app_test.py:tests succeeded in 1.22s
✓ python/services/user_service/app_test.py:tests succeeded in 1.20s
```

**memoized** 表示 Pants 判斷自上次執行以來程式碼沒有變動，直接使用快取結果，不重新執行測試。

---

## 新增功能的步驟

### 情境：新增一個 Go notification service

```bash
# 1. 建立目錄和檔案
mkdir -p go/cmd/notificationapi
cat > go/cmd/notificationapi/main.go << 'EOF'
package main
// ...
EOF

# 2. 讓 Pants 自動產生 BUILD
pants tailor go/cmd/notificationapi/

# 3. 驗證
pants check go/cmd/notificationapi:
pants test go/cmd/notificationapi:
```

### 情境：新增一個 Python analytics service

```bash
# 1. 建立目錄和檔案
mkdir -p python/services/analytics_service

# 2. 建立 BUILD
cat > python/services/analytics_service/BUILD << 'EOF'
python_sources(name="lib")
python_tests(name="tests")
pex_binary(name="bin", entry_point="app.py")
EOF

# 3. 建立 app.py（可以 import common.models）
```

---

## 新增 Python 依賴

```bash
# 1. 在 requirements.txt 新增套件
echo "redis==5.2.1" >> requirements.txt

# 2. 重新產生 lockfile
pants generate-lockfiles

# 3. 提交兩個檔案
git add requirements.txt python-default.lock
```

---

## CI/CD 整合範例

```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'

      - name: Bootstrap Pants
        run: pants --version

      - name: Run all tests
        run: pants test ::

      - name: Type check
        run: pants check ::

      - name: Package all binaries
        run: pants package ::

  # 只測試有改動的部分（Pull Request 時更快）
  changed-test:
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # 需要完整歷史來計算 diff

      - name: Test changed targets only
        run: pants --changed-since=origin/main --changed-dependees=transitive test
```

---

## 效能對比

在大型 monorepo 中，`--changed-since` 的效果非常顯著：

```
修改一個檔案：go/pkg/models/models.go

pants test ::                            → 跑 5 個 test targets（全部）
pants --changed-since=HEAD test          → 只跑 2 個 test targets（受影響的）
  ↑ go/pkg/models:models                （直接改動）
  ↑ go/cmd/userapi:userapi              （依賴 models 的 binary）
  ↑ go/cmd/orderapi:orderapi            （依賴 models 的 binary）

省下約 40% 的時間（在有 100+ services 的 repo 中省下 95%+）
```
