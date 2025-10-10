# 贡献指南

感谢您对 iplocx 的关注。本指南将帮助您快速开始贡献。

## 开发环境

### 前置条件

- Go 1.18+
- Git
- 数据库文件（用于测试）

### 环境设置

```bash
# 克隆仓库
git clone https://github.com/nuomiaa/iplocx.git
cd iplocx

# 安装依赖
go mod download

# 下载测试数据
# QQwry: https://raw.githubusercontent.com/FW27623/qqwry/refs/heads/main/qqwry.dat
# GeoLite2: https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
# 将文件放置在 data/ 目录下
```

## 开发流程

### 1. 创建分支

```bash
git checkout -b feature/your-feature-name
# 或
git checkout -b fix/bug-description
```

### 2. 代码开发

确保您的代码符合以下标准：

- 遵循 [Effective Go](https://golang.org/doc/effective_go.html) 指南
- 使用 `gofmt` 格式化代码
- 导出的函数和类型必须有文档注释
- 保持函数简短，职责单一

### 3. 编写测试

- 新功能必须包含单元测试
- 测试覆盖率应保持在 80% 以上
- 性能相关代码需要添加 benchmark

```bash
# 运行测试
go test -v ./...

# 查看覆盖率
go test -cover ./...

# 运行性能测试
go test -bench=. -benchmem
```

### 4. 代码检查

```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 运行 linter（如果安装了 golangci-lint）
golangci-lint run
```

### 5. 提交更改

使用语义化的提交信息：

```bash
git commit -m "feat: 添加新功能"
git commit -m "fix: 修复查询错误"
git commit -m "docs: 更新API文档"
git commit -m "test: 添加单元测试"
git commit -m "perf: 优化查询性能"
git commit -m "refactor: 重构数据合并逻辑"
```

提交类型：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `test`: 测试相关
- `perf`: 性能优化
- `refactor`: 代码重构
- `style`: 代码格式
- `chore`: 构建/工具相关

### 6. 提交 Pull Request

```bash
git push origin feature/your-feature-name
```

在 GitHub 上创建 Pull Request，并：
- 清楚描述更改内容
- 关联相关的 Issue
- 确保所有测试通过

## 代码规范

### 命名约定

- 导出的标识符使用大驼峰：`NewLocator`, `QueryStats`
- 私有标识符使用小驼峰：`mergeData`, `scoreProvider`
- 常量使用大驼峰或全大写：`ErrInvalidIP`, `DefaultCacheSize`

### 文档注释

```go
// NewLocator creates a new IP location query handler.
// It initializes the specified data providers and returns an error
// if no valid providers are available.
func NewLocator(cfg Config) (*Locator, error) {
    // ...
}
```

### 错误处理

```go
// 使用包级错误常量
var ErrInvalidIP = errors.New("invalid IP address")

// 返回明确的错误信息
if ip == "" {
    return nil, ErrInvalidIP
}
```

## 测试规范

### 单元测试

```go
func TestQuery(t *testing.T) {
    locator, err := NewLocator(Config{
        QQwryDBPath: "../data/qqwry.dat",
    })
    if err != nil {
        t.Fatal(err)
    }
    defer locator.Close()

    location, err := locator.Query("8.8.8.8")
    if err != nil {
        t.Errorf("Query failed: %v", err)
    }
    if location.Country == "" {
        t.Error("Expected country information")
    }
}
```

### 性能测试

```go
func BenchmarkQuery(b *testing.B) {
    locator, _ := NewLocator(Config{
        QQwryDBPath: "../data/qqwry.dat",
    })
    defer locator.Close()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        locator.Query("8.8.8.8")
    }
}
```

## 报告问题

### Bug 报告

创建 Issue 时请包含：

- Go 版本 (`go version`)
- 操作系统和架构
- 最小可复现代码
- 预期行为
- 实际行为
- 错误日志（如果有）

示例：

```markdown
**环境**
- Go 版本: 1.21.0
- OS: Ubuntu 22.04 (amd64)

**问题描述**
查询特定IP时返回错误...

**复现步骤**
1. ...
2. ...

**期望结果**
应该返回...

**实际结果**
返回错误: ...
```

### 功能请求

包含以下信息：

- 功能描述
- 使用场景
- 预期 API 设计（可选）
- 实现建议（可选）

## 文档更新

如果您的更改影响到用户接口，请同步更新：

- `README.md` - 主要文档
- `example_test.go` - 示例代码
- GoDoc 注释 - 代码文档
- `examples/` - 示例程序（如需要）

## 发布流程

由维护者负责：

1. 更新版本号
2. 更新 CHANGELOG
3. 创建 Git tag
4. 发布 GitHub Release

## 行为准则

- 保持专业和尊重
- 欢迎建设性反馈
- 尊重不同的观点
- 专注于对项目最有利的决策

## 获取帮助

如有任何问题：

- 创建 GitHub Issue
- 查看现有的 Pull Requests
- 阅读 [GoDoc 文档](https://godoc.org/github.com/nuomiaa/iplocx)

感谢您的贡献！
