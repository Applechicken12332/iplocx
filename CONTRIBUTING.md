# 贡献指南

感谢你考虑为 iplocx 做出贡献！

## 开发流程

### 1. Fork 和克隆

```bash
# Fork 仓库后克隆到本地
git clone https://github.com/nuomiaa/iplocx.git
cd iplocx
```

### 2. 创建分支

```bash
git checkout -b feature/your-feature-name
# 或
git checkout -b fix/your-bug-fix
```

### 3. 进行更改

- 遵循 Go 代码规范
- 添加必要的测试
- 更新文档（如果需要）

### 4. 运行测试

```bash
# 运行所有测试
go test -v ./...

# 检查代码格式
go fmt ./...

# 运行 linter
go vet ./...

# 运行性能测试
go test -bench=. -benchmem
```

### 5. 提交更改

```bash
git add .
git commit -m "feat: 添加新功能"
# 或
git commit -m "fix: 修复某个bug"
```

提交信息格式：
- `feat:` 新功能
- `fix:` Bug 修复
- `docs:` 文档更新
- `test:` 添加测试
- `perf:` 性能优化
- `refactor:` 代码重构

### 6. 推送和创建 PR

```bash
git push origin feature/your-feature-name
```

然后在 GitHub 上创建 Pull Request。

## 代码规范

### Go 代码风格

- 使用 `gofmt` 格式化代码
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 导出的函数和类型必须有文档注释
- 保持函数简短，职责单一

### 测试要求

- 新功能必须有单元测试
- 测试覆盖率应保持在 80% 以上
- 性能敏感的代码需要添加 benchmark

### 文档要求

- 所有公开的 API 都要有清晰的文档注释
- 示例代码应该可以直接运行
- 更新 README.md（如果功能有变化）

## 报告问题

### Bug 报告

请包含：
- Go 版本
- 操作系统
- 重现步骤
- 预期行为
- 实际行为
- 错误日志

### 功能请求

请包含：
- 功能描述
- 使用场景
- 预期API（如果有想法）

## 开发环境设置

### 前置条件

- Go 1.18+
- Git

### 安装依赖

```bash
go mod download
```

### 下载测试数据

1. QQwry 数据库: http://update.cz88.net/soft/qqwry.rar
2. GeoLite2 数据库: https://dev.maxmind.com/geoip/geolite2-free-geolocation-data

### 运行测试

```bash
go test -v
```

## 行为准则

- 尊重所有贡献者
- 保持专业和友好
- 接受建设性的批评

## 问题？

如有任何问题，请：
- 提 Issue
- 发送邮件到 [nuomiaa@gmail.com]

感谢你的贡献！🎉

