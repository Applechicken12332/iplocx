# 项目结构

本文档描述 iplocx 项目的代码组织和架构设计。

## 目录结构

```
iplocx/
├── README.md                   # 项目文档
├── LICENSE                     # Apache 2.0 许可证
├── CONTRIBUTING.md             # 贡献指南
├── PROJECT_STRUCTURE.md        # 项目结构说明
├── go.mod                      # Go 模块定义
├── go.sum                      # 依赖锁定文件
│
├── 核心代码
│   ├── locator.go              # 主查询器实现
│   ├── cache.go                # LRU 缓存
│   ├── types.go                # 数据结构定义
│   ├── errors.go               # 错误定义
│   ├── qqwry_provider.go       # QQwry 数据提供者
│   └── geolite_provider.go     # GeoLite2 数据提供者
│
├── 测试代码
│   ├── locator_test.go         # 核心功能测试
│   ├── locator_bench_test.go   # 性能基准测试
│   ├── cache_test.go           # 缓存测试
│   ├── example_test.go         # 示例代码
│   └── example_stats_test.go   # 统计功能示例
│
├── internal/                   # 内部实现
│   └── qqwry/
│       └── qqwry.go           # QQwry 解析器
│
├── data/                       # 数据文件（不在版本控制中）
│   ├── qqwry.dat              # 纯真IP库
│   └── GeoLite2-City.mmdb     # GeoLite2 数据库
│
└── examples/                   # 示例程序
    ├── README.md
    ├── basic/
    ├── with_cache/
    ├── stats_monitor/
    └── complete_test/
```

## 核心模块

### locator.go (461 行)

主查询器实现，包含：

- `Locator` 结构体：查询器主体
- `NewLocator()`: 初始化方法
- `Query()`: IP 查询接口
- `queryProviders()`: 并行查询数据源
- `mergeLocations()`: 智能数据合并
- 统计和监控接口

关键特性：
- 并行查询双数据源
- 基于评分的数据合并算法
- 实时性能统计
- 线程安全设计

### cache.go (115 行)

LRU 缓存实现：

- `Cache` 结构体
- `Get()`, `Set()`: 缓存操作
- `GetStats()`: 缓存统计
- 线程安全的读写锁

性能特点：
- 11.45 ns 查询延迟
- 1亿+ QPS 吞吐量
- O(1) 时间复杂度

### types.go (47 行)

数据结构定义：

- `Location`: IP 地理位置信息
- `Config`: 配置选项
- `CacheStats`: 缓存统计
- `QueryStatsSnapshot`: 查询统计

辅助方法：
- `IsEmpty()`: 判断位置信息是否为空
- `HasDetailedInfo()`: 是否有详细信息
- `GetDetailScore()`: 计算详细程度评分

### qqwry_provider.go (79 行)

QQwry 数据源实现：

- 解析纯真IP库格式
- 提取国内IP详细信息
- 支持运营商和区县级信息
- 仅支持 IPv4

### geolite_provider.go (99 行)

GeoLite2 数据源实现：

- 使用 MaxMind 官方 SDK
- 支持 IPv4 和 IPv6
- 提供经纬度和时区信息
- 国际IP准确度高

### errors.go (18 行)

错误常量定义：

```go
var (
    ErrDatabaseNotFound
    ErrInvalidIP
    ErrNoData
    ErrNoProvider
)
```

## 测试架构

### 单元测试 (locator_test.go, 345 行)

覆盖范围：
- 初始化和配置验证
- 查询功能（IPv4/IPv6）
- 数据合并逻辑
- 错误处理
- 统计功能
- 并发安全性

### 性能测试 (locator_bench_test.go, 104 行)

测试场景：
- `BenchmarkQQwryOnly`: QQwry 性能
- `BenchmarkGeoLiteOnly`: GeoLite2 性能
- `BenchmarkQuery`: 双数据源合并性能
- `BenchmarkQueryParallel`: 多核并发性能

### 缓存测试 (cache_test.go, 160 行)

测试项：
- LRU 淘汰策略
- 并发读写安全
- 统计准确性
- `BenchmarkCacheGet`: 缓存性能

### 示例代码 (example_test.go, 105 行)

GoDoc 可见的示例：
- 基本查询
- 缓存使用
- 统计监控

## 数据流

```
用户请求
    ↓
Query(ip)
    ↓
缓存检查 ────命中───→ 返回结果
    ↓ 未命中
并行查询
    ├─→ QQwry
    └─→ GeoLite2
    ↓
数据合并（评分算法）
    ↓
写入缓存
    ↓
返回结果
```

## 性能指标

基于 AMD Ryzen 9 7945HX (32核):

| 模块 | 延迟 | 吞吐量 (QPS) |
|------|------|-------------|
| 缓存 | 11.45 ns | 104,000,000 |
| 双源合并 | 9.57 μs | 117,000 |
| QQwry | 2.17 μs | 541,000 |
| GeoLite2 | 1.94 μs | 614,000 |

测试覆盖率：**80.8%**

## 依赖关系

```go
// go.mod
module github.com/nuomiaa/iplocx

require (
    github.com/oschwald/geoip2-golang v1.x.x
    golang.org/x/text v0.x.x
)
```

外部依赖：
- `oschwald/geoip2-golang`: GeoIP2 数据库读取
- `golang.org/x/text`: 文本编码转换（GBK）

## 设计原则

### 1. 模块化设计

- 独立的数据提供者接口
- 可插拔的缓存层
- 清晰的职责分离

### 2. 并发安全

- 无锁读取设计
- 读写锁保护缓存
- Goroutine 安全的统计

### 3. 性能优先

- 并行查询减少延迟
- LRU 缓存提升重复查询
- 零拷贝数据传递

### 4. 可测试性

- 依赖注入设计
- 完整的单元测试
- 性能基准测试

## 扩展指南

### 添加新数据源

1. 创建 `xxx_provider.go`
2. 实现 `Query(ip string) (*Location, error)`
3. 在 `NewLocator()` 中集成
4. 添加对应测试

### 优化查询性能

1. 添加 benchmark 测试
2. 使用 pprof 分析瓶颈
3. 优化热点代码
4. 验证性能提升

### 添加新功能

1. 在 `types.go` 定义新结构
2. 在 `locator.go` 实现逻辑
3. 添加单元测试
4. 更新文档和示例

## 版本控制

- 主分支：`main`
- 开发分支：`feature/*`, `fix/*`
- 标签格式：`v1.2.3` (语义化版本)

## 文档维护

核心文档：
- `README.md`: 用户文档和 API 说明
- `CONTRIBUTING.md`: 开发者指南
- `PROJECT_STRUCTURE.md`: 架构设计
- GoDoc 注释: 代码级文档

更新原则：
- 代码变更同步更新文档
- 保持示例代码可运行
- 及时更新性能数据

## 相关资源

- [GoDoc](https://godoc.org/github.com/nuomiaa/iplocx)
- [GitHub Repository](https://github.com/nuomiaa/iplocx)
- [Go Report Card](https://goreportcard.com/report/github.com/nuomiaa/iplocx)
