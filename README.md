# iplocx - IP地理位置查询库

[![Go Report Card](https://goreportcard.com/badge/github.com/nuomiaa/iplocx)](https://goreportcard.com/report/github.com/nuomiaa/iplocx)
[![GoDoc](https://godoc.org/github.com/nuomiaa/iplocx?status.svg)](https://godoc.org/github.com/nuomiaa/iplocx)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Go语言IP地理位置查询库，智能结合纯真IP库(QQwry)和GeoLite2数据库，提供准确、全面的IP地理位置信息。

## ✨ 特性

- 🔍 **智能查询策略**：并行查询双数据源，智能合并结果
- 🌍 **双数据源支持**：结合国内和国际IP数据库的优势
- 📊 **智能数据合并**：基于评分系统的智能合并算法
- 🔒 **线程安全**：支持高并发查询，无锁竞争
- 🎯 **统一接口**：简洁易用的API设计
- 📦 **独立模块**：标准Go包结构，易于集成
- ⚡ **LRU缓存**：可选的高性能缓存机制，性能提升数百倍
- 📈 **性能监控**：完整的查询统计和性能指标
- 🐛 **调试模式**：可控的详细日志输出，便于排查问题
- ✅ **完整测试**：80.8% 测试覆盖率，全面的单元测试和示例

## 🚀 快速开始

### 安装

```bash
go get github.com/nuomiaa/iplocx
```

### 基本使用

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/nuomiaa/iplocx"
)

func main() {
    // 配置数据源
    cfg := iplocx.Config{
        QQwryDBPath:   "./data/qqwry.dat",
        GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
    }

    // 创建查询器
    locator, err := iplocx.NewLocator(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer locator.Close()

    // 查询IP
    location, err := locator.Query("8.8.8.8")
    if err != nil {
        log.Fatal(err)
    }

    // 使用结果
    fmt.Printf("国家: %s\n", location.Country)
    fmt.Printf("省份: %s\n", location.Province)
    fmt.Printf("城市: %s\n", location.City)
    fmt.Printf("运营商: %s\n", location.ISP)
    fmt.Printf("经纬度: %.4f, %.4f\n", location.Latitude, location.Longitude)
    fmt.Printf("时区: %s\n", location.TimeZone)
}
```

## 📖 使用指南

### 启用缓存

启用缓存可以大幅提升重复查询性能（提升100倍以上）：

```go
cfg := iplocx.Config{
    QQwryDBPath:   "./data/qqwry.dat",
    GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
    EnableCache:   true,
    CacheSize:     1000, // 缓存1000条记录
}

locator, _ := iplocx.NewLocator(cfg)
defer locator.Close()

// 查询缓存统计
stats := locator.GetCacheStats()
fmt.Printf("缓存命中率: %.2f%%\n", stats.HitRate)
fmt.Printf("当前缓存: %d/%d\n", stats.Size, stats.Capacity)
```

### 性能监控

获取查询统计信息，监控系统性能：

```go
// 获取查询统计
stats := locator.GetQueryStats()
fmt.Printf("总查询次数: %d\n", stats.TotalQueries)
fmt.Printf("成功次数: %d\n", stats.SuccessQueries)
fmt.Printf("成功率: %.2f%%\n", stats.SuccessRate)
fmt.Printf("平均耗时: %v\n", stats.AvgDuration)

// 数据源使用情况
fmt.Printf("QQwry使用: %d次\n", stats.QQwryHits)
fmt.Printf("GeoLite使用: %d次\n", stats.GeoLiteHits)
fmt.Printf("数据合并: %d次\n", stats.CombinedHits)

// 重置统计
locator.ResetStats()
```

### 检查数据源状态

在初始化后检查数据源是否正常加载：

```go
// 简单状态检查
status := locator.GetProviderStatus()
fmt.Printf("QQwry可用: %v\n", status["qqwry"])
fmt.Printf("GeoLite可用: %v\n", status["geolite"])

// 详细信息（包含错误）
info := locator.GetProviderInfo()
for name, provInfo := range info {
    fmt.Printf("%s: 可用=%v\n", name, provInfo.Available)
    if len(provInfo.Errors) > 0 {
        fmt.Printf("  错误: %v\n", provInfo.Errors)
    }
}
```

### 调试模式

启用调试模式查看详细的数据合并过程：

```go
cfg := iplocx.Config{
    QQwryDBPath:   "./data/qqwry.dat",
    GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
    Debug:         true, // 启用调试输出
}

locator, _ := iplocx.NewLocator(cfg)
defer locator.Close()

// 查询时会输出详细的数据合并过程
location, _ := locator.Query("8.8.8.8")
// 输出示例:
// === 数据合并调试 ===
// QQwry  评分: 15 | Country:美国 | Province:加利福尼亚州 ...
// GeoLite 评分: 1 | Country:美国 | Province: ...
// 选择 qqwry 作为主数据源（分数: 15 vs 1）
```

## 📊 数据结构

### Location

```go
type Location struct {
    IP        string  // IP地址
    Country   string  // 国家
    Province  string  // 省/州
    City      string  // 市
    District  string  // 区/县
    ISP       string  // 运营商
    Latitude  float64 // 纬度
    Longitude float64 // 经度
    TimeZone  string  // 时区
    Source    string  // 数据来源 (qqwry/geolite2/combined)
}
```

**辅助方法：**

- `IsEmpty() bool` - 判断位置信息是否为空
- `HasDetailedInfo() bool` - 判断是否有详细的省市信息
- `GetDetailScore() int` - 获取详细程度分数（国家1分+省2分+市4分+区8分）

### Config

```go
type Config struct {
    QQwryDBPath   string // 纯真IP库路径（可选）
    GeoLiteDBPath string // GeoLite2路径（可选）
    Debug         bool   // 是否启用调试输出
    EnableCache   bool   // 是否启用缓存
    CacheSize     int    // 缓存大小（默认1000）
}
```

**注意：** 至少需要配置一个数据源（QQwry 或 GeoLite2）

### CacheStats

```go
type CacheStats struct {
    Size     int     // 当前缓存数量
    Capacity int     // 缓存容量
    Hits     int64   // 命中次数
    Misses   int64   // 未命中次数
    HitRate  float64 // 命中率（百分比）
}
```

### QueryStatsSnapshot

```go
type QueryStatsSnapshot struct {
    TotalQueries   int64         // 总查询次数
    SuccessQueries int64         // 成功查询次数（含缓存命中）
    FailedQueries  int64         // 失败查询次数
    QQwryHits      int64         // QQwry数据源使用次数
    GeoLiteHits    int64         // GeoLite数据源使用次数
    CombinedHits   int64         // 合并数据使用次数
    AvgDuration    time.Duration // 平均查询时间
    SuccessRate    float64       // 成功率（百分比）
}
```

## 🔧 API 参考

### 主要方法

| 方法 | 说明 |
|------|------|
| `NewLocator(cfg Config) (*Locator, error)` | 创建IP查询器 |
| `Query(ip string) (*Location, error)` | 查询IP地址 |
| `GetCacheStats() *CacheStats` | 获取缓存统计信息 |
| `GetQueryStats() QueryStatsSnapshot` | 获取查询统计信息 |
| `GetProviderStatus() map[string]bool` | 获取数据源状态 |
| `GetProviderInfo() map[string]ProviderInfo` | 获取详细数据源信息 |
| `ClearCache()` | 清空缓存 |
| `ResetStats()` | 重置统计信息 |
| `Close() error` | 关闭查询器，释放资源 |

### 错误类型

```go
var (
    ErrDatabaseNotFound // 数据库文件未找到
    ErrInvalidIP        // 无效的IP地址
    ErrNoData           // 未找到数据
    ErrNoProvider       // 没有可用的查询提供者
)
```

## 🎯 查询策略

### 1. 缓存优先

如果启用缓存，首先从缓存中查找，缓存命中直接返回。

### 2. 并行查询

同时查询 QQwry 和 GeoLite2 两个数据源，减少查询延迟。

### 3. 智能合并

根据数据完整度评分选择主数据源，用另一个数据源补充缺失字段：

- **QQwry 优势**：国内IP信息准确，包含运营商、区县信息
- **GeoLite2 优势**：支持IPv6，国际IP准确，包含经纬度、时区信息

**合并规则：**

- 选择评分高的数据作为基础
- ISP信息优先使用QQwry
- 经纬度和时区优先使用GeoLite2
- 地理信息字段按需补充

## ⚡ 性能

### 性能指标

| 场景 | 性能 | 说明 |
|------|------|------|
| QQwry 查询 | ~50,000 ops/s | 单独使用QQwry |
| GeoLite2 查询 | ~30,000 ops/s | 单独使用GeoLite2 |
| 并行合并查询 | ~25,000 ops/s | 双数据源智能合并 |
| 缓存命中 | ~1,000,000 ops/s | LRU缓存命中 |

### 内存占用

- **QQwry**: ~30MB
- **GeoLite2**: ~80MB
- **缓存**: 每1000条约 ~200KB

### 性能优化建议

1. **高频查询场景**：启用缓存，设置合适的 CacheSize
2. **内存受限**：关闭缓存，只使用一个数据源
3. **国内业务为主**：优先使用 QQwry
4. **国际业务为主**：使用 GeoLite2
5. **混合场景**：双数据源 + 缓存（推荐）

## 📦 数据库文件

### QQwry（纯真IP库）

- **下载地址**: https://raw.githubusercontent.com/FW27623/qqwry/refs/heads/main/qqwry.dat
- **文件名**: `qqwry.dat`
- **特点**: 
  - 国内IP信息准确
  - 包含运营商信息
  - 包含区县级信息
  - 仅支持 IPv4

### GeoLite2

- **下载地址**: https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
- **文件名**: `GeoLite2-City.mmdb`
- **特点**: 
  - 支持 IPv4 和 IPv6
  - 国际IP信息准确
  - 包含经纬度坐标
  - 包含时区信息

**注意**: 需要注册 MaxMind 账号才能下载 GeoLite2 数据库。

## 📝 示例程序

查看 [examples](examples/) 目录获取完整的示例程序：

### 基础示例

```bash
cd examples/basic
go run main.go
```

演示基本的 IP 查询功能，包括国内外 IP 的查询。

### 缓存性能测试

```bash
cd examples/with_cache
go run main.go
```

演示缓存对性能的提升，对比有无缓存的查询速度。

### 统计监控示例

```bash
cd examples/stats_monitor
go run main.go
```

展示完整的统计功能，包括查询成功率、数据源使用情况、缓存命中率、平均查询时间。

### 完整功能测试 ⭐（推荐）

```bash
cd examples/complete_test
go run main.go
```

**这是最全面的测试示例**，涵盖所有功能，包含 12 个完整测试模块：

1. ✅ 基础初始化和数据源状态检查
2. ✅ IPv4 国内 IP 查询（详细信息）
3. ✅ IPv4 国外 IP 查询
4. ✅ IPv6 地址查询
5. ✅ LRU 缓存功能和性能对比
6. ✅ 批量查询性能测试
7. ✅ 查询统计信息展示
8. ✅ 错误处理机制验证
9. ✅ Location 辅助方法测试
10. ✅ 调试模式演示
11. ✅ 并发安全性测试
12. ✅ 资源清理功能测试

**适用场景：**
- 首次使用，想全面了解所有功能
- 验证安装配置是否正确
- 性能基准测试
- 开发调试参考

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test -v

# 查看测试覆盖率
go test -cover

# 运行性能基准测试
go test -bench=. -benchmem

# 运行特定测试
go test -v -run TestCache
```

### 测试覆盖率

当前测试覆盖率: **80.8%**

包含：
- 单元测试（locator_test.go）
- 缓存测试（cache_test.go）
- 性能基准测试（locator_bench_test.go）
- 示例测试（example_test.go, example_stats_test.go）

## 🔍 使用场景

### 适用场景

- ✅ Web 应用用户地理位置识别
- ✅ 访问日志分析和统计
- ✅ 内容分发网络（CDN）调度
- ✅ 防欺诈和安全审计
- ✅ 广告定向投放
- ✅ 访问控制和地域限制

### 不适用场景

- ❌ 需要实时更新的 IP 数据库（需自行实现数据库更新机制）
- ❌ 超高精度定位需求（数据库精度有限）
- ❌ 分布式缓存共享（当前为本地缓存）

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

详见 [CONTRIBUTING.md](CONTRIBUTING.md)

## 📄 许可证

本项目采用 Apache License 2.0 许可证 - 详见 [LICENSE](LICENSE) 文件

Apache 2.0 许可证特点：
- ✅ 允许商业使用
- ✅ 允许修改和分发
- ✅ 明确授予专利权
- ✅ 要求保留版权和许可声明
- ✅ 要求说明修改内容

## 🙏 致谢

- [纯真IP库](http://www.cz88.net/) - 提供国内IP数据
- [MaxMind GeoLite2](https://www.maxmind.com/) - 提供全球IP地理位置数据

---

**Made with ❤️ by [nuomiaa](https://github.com/nuomiaa)**
