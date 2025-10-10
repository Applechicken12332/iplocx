# iplocx 项目结构说明

## 📁 目录结构

```
iplocx/                          # 包根目录
├── README.md                       # 项目说明文档（包含使用教程、API文档）
├── LICENSE                         # Apache 2.0 许可证
├── CONTRIBUTING.md                 # 贡献指南
├── PROJECT_STRUCTURE.md            # 本文件 - 项目结构说明
├── .gitignore                      # Git 忽略文件配置
├── go.mod                          # Go 模块定义
├── go.sum                          # Go 依赖锁定文件
│
├── 核心代码文件
│   ├── locator.go                  # 主查询器实现（Locator 结构体）
│   ├── cache.go                    # LRU 缓存实现
│   ├── types.go                    # 数据结构定义（Location 等）
│   ├── errors.go                   # 错误定义
│   ├── qqwry_provider.go           # QQwry 数据提供者
│   └── geolite_provider.go         # GeoLite2 数据提供者
│
├── 测试文件（重要！）
│   ├── locator_test.go             # 主查询器单元测试
│   ├── locator_bench_test.go       # 性能基准测试
│   ├── cache_test.go               # 缓存功能测试
│   ├── example_test.go             # 示例代码（会出现在 godoc）
│   └── example_stats_test.go       # 统计功能示例
│
├── internal/                       # 内部实现
│   └── qqwry/                      # QQwry 内部实现
│       └── qqwry.go
│
├── data/                           # 数据文件（不包含在 git 中）
│   ├── qqwry.dat                   # 纯真IP库（用户自行下载）
│   └── GeoLite2-City.mmdb          # GeoLite2数据库（用户自行下载）
│
└── examples/                       # 完整示例程序目录
    ├── README.md                   # 示例说明
    ├── basic/                      # 基础查询示例
    │   └── main.go
    ├── with_cache/                 # 缓存性能测试
    │   └── main.go
    ├── stats_monitor/              # 统计监控示例
    │   └── main.go
    └── complete_test/              # 完整功能测试 ⭐
        └── main.go
```

## 📝 文件说明

### 核心代码文件

| 文件 | 行数 | 说明 |
|------|------|------|
| `locator.go` | ~461 | 核心查询逻辑、数据合并算法、统计功能 |
| `cache.go` | ~115 | LRU 缓存实现，线程安全 |
| `types.go` | ~47 | 数据结构定义和辅助方法 |
| `errors.go` | ~18 | 错误常量定义 |
| `qqwry_provider.go` | ~79 | QQwry 数据库查询实现 |
| `geolite_provider.go` | ~99 | GeoLite2 数据库查询实现 |

### 测试文件

| 文件 | 行数 | 说明 |
|------|------|------|
| `locator_test.go` | ~345 | 完整的单元测试，覆盖所有功能 |
| `cache_test.go` | ~160 | 缓存功能测试和性能测试 |
| `locator_bench_test.go` | ~104 | 性能基准测试 |
| `example_test.go` | ~105 | 基础示例（godoc 可见）|
| `example_stats_test.go` | ~82 | 统计功能示例（godoc 可见）|

**测试文件非常重要！**
- ✅ 保证代码质量
- ✅ 提供使用示例
- ✅ 验证跨平台兼容性
- ✅ 不会编译到最终程序中

### 示例程序

| 示例 | 说明 |
|------|------|
| `examples/basic/` | 演示基本的 IP 查询功能 |
| `examples/with_cache/` | 演示缓存带来的性能提升 |
| `examples/stats_monitor/` | 演示统计和监控功能 |
| `examples/complete_test/` ⭐ | **完整功能测试（推荐）** - 包含12个测试模块 |

## 📦 作为独立包使用

### 安装

```bash
go get github.com/nuomiaa/iplocx
```

### 导入

```go
import "github.com/nuomiaa/iplocx"
```

### 文档

- GoDoc: https://godoc.org/github.com/nuomiaa/iplocx

## 🔧 开发工作流

### 添加新功能

1. 在相应的 `.go` 文件中实现
2. 在对应的 `*_test.go` 中添加测试
3. 更新 README.md 文档
4. 运行测试: `go test -v`

### 性能优化

1. 添加 benchmark: `func BenchmarkXXX(b *testing.B)`
2. 运行基准测试: `go test -bench=.`
3. 对比优化前后的结果

## 🏗️ 技术栈

- **Go 版本**: 1.21+
- **许可证**: Apache 2.0
- **依赖**:
  - `github.com/oschwald/geoip2-golang/v2` - GeoIP2 数据库读取（beta）
  - `golang.org/x/text` - 文本编码转换（GBK）
