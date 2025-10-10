# iplocx

<div align="center">

[![GitHub release](https://img.shields.io/github/v/release/nuomiaa/iplocx?include_prereleases&sort=semver&logo=github)](https://github.com/nuomiaa/iplocx/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/nuomiaa/iplocx?logo=go&logoColor=white)](https://github.com/nuomiaa/iplocx/blob/main/go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/nuomiaa/iplocx)](https://goreportcard.com/report/github.com/nuomiaa/iplocx)
[![codecov](https://img.shields.io/badge/coverage-80.8%25-brightgreen?logo=codecov)](https://github.com/nuomiaa/iplocx)
[![CI Status](https://img.shields.io/badge/build-passing-brightgreen?logo=github-actions)](https://github.com/nuomiaa/iplocx/actions)

[![GitHub stars](https://img.shields.io/github/stars/nuomiaa/iplocx?style=social)](https://github.com/nuomiaa/iplocx/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/nuomiaa/iplocx?style=social)](https://github.com/nuomiaa/iplocx/network/members)
[![GitHub watchers](https://img.shields.io/github/watchers/nuomiaa/iplocx?style=social)](https://github.com/nuomiaa/iplocx/watchers)
[![GitHub contributors](https://img.shields.io/github/contributors/nuomiaa/iplocx)](https://github.com/nuomiaa/iplocx/graphs/contributors)

[![GoDoc](https://pkg.go.dev/badge/github.com/nuomiaa/iplocx.svg)](https://pkg.go.dev/github.com/nuomiaa/iplocx)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](https://opensource.org/licenses/Apache-2.0)
[![GitHub issues](https://img.shields.io/github/issues/nuomiaa/iplocx?logo=github)](https://github.com/nuomiaa/iplocx/issues)
[![GitHub pull requests](https://img.shields.io/github/issues-pr/nuomiaa/iplocx?logo=github)](https://github.com/nuomiaa/iplocx/pulls)
[![Last Commit](https://img.shields.io/github/last-commit/nuomiaa/iplocx?logo=git&logoColor=white)](https://github.com/nuomiaa/iplocx/commits/main)

[![Language](https://img.shields.io/badge/language-Go-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](https://github.com/nuomiaa/iplocx)
[![Architecture](https://img.shields.io/badge/arch-amd64%20%7C%20arm64-blue)](https://github.com/nuomiaa/iplocx)
[![Contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](CONTRIBUTING.md)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/nuomiaa/iplocx/graphs/commit-activity)

</div>

---

<p align="center">
高性能 IP 地理位置查询库，采用并行查询和智能数据合并技术，支持 IPv4/IPv6
</p>

<p align="center">
<a href="#性能指标">性能指标</a> •
<a href="#快速开始">快速开始</a> •
<a href="#核心技术">核心技术</a> •
<a href="#api-文档">API 文档</a> •
<a href="#性能基准测试">性能测试</a> •
<a href="#贡献">贡献</a>
</p>

---

## 目录

- [性能指标](#性能指标)
- [核心技术](#核心技术)
- [快速开始](#快速开始)
- [API 文档](#api-文档)
- [高级功能](#高级功能)
- [性能基准测试](#性能基准测试)
- [内存占用](#内存占用)
- [数据库文件](#数据库文件)
- [应用场景](#应用场景)
- [测试](#测试)
- [示例程序](#示例程序)
- [错误处理](#错误处理)
- [贡献](#贡献)
- [许可证](#许可证)
- [致谢](#致谢)

---

## 性能指标

基于 AMD Ryzen 9 7945HX (32核) 的基准测试结果：

| 查询模式 | 延迟 | 吞吐量 (QPS) |
|---------|------|-------------|
| 缓存命中 | 11.45 ns | 104,000,000 |
| 双数据源合并 | 9.57 μs | 117,000 |
| 8核并发 | 1.31 μs | 908,000 |
| QQwry 单独 | 2.17 μs | 541,000 |
| GeoLite2 单独 | 1.94 μs | 614,000 |

测试覆盖率：80.8%

## 核心技术

### 并行查询架构
同时查询 QQwry 和 GeoLite2 两个数据源，通过 Goroutine 并发执行，减少查询延迟。

### 智能数据合并
基于评分系统自动选择最优数据源：
- 国内 IP：优先使用 QQwry（包含运营商、区县级信息）
- 国际 IP：优先使用 GeoLite2（包含经纬度、时区信息）
- 自动补充缺失字段，提供完整的地理位置信息

### LRU 缓存机制
可选的 LRU 缓存层，缓存命中后性能提升 890 倍。

### 线程安全设计
无锁并发架构，支持高并发查询场景。

## 快速开始

### 安装

```bash
go get github.com/nuomiaa/iplocx
```

### 基本用法

```go
package main

import (
    "fmt"
    "log"
    "github.com/nuomiaa/iplocx"
)

func main() {
    locator, err := iplocx.NewLocator(iplocx.Config{
        QQwryDBPath:   "./data/qqwry.dat",
        GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
        EnableCache:   true,
        CacheSize:     10000,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer locator.Close()

    location, err := locator.Query("8.8.8.8")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%s | %s, %s, %s | %s | %.4f, %.4f | %s\n",
        location.IP,
        location.Country,
        location.Province,
        location.City,
        location.ISP,
        location.Latitude,
        location.Longitude,
        location.TimeZone,
    )
}
```

## API 文档

### 配置选项

```go
type Config struct {
    QQwryDBPath   string  // 纯真IP库路径（可选）
    GeoLiteDBPath string  // GeoLite2路径（可选）
    EnableCache   bool    // 启用LRU缓存
    CacheSize     int     // 缓存容量（默认1000）
    Debug         bool    // 调试模式
}
```

至少需要配置一个数据源。

### 核心方法

| 方法 | 说明 |
|------|------|
| `NewLocator(cfg Config) (*Locator, error)` | 创建查询器实例 |
| `Query(ip string) (*Location, error)` | 查询IP地址 |
| `GetCacheStats() *CacheStats` | 获取缓存统计 |
| `GetQueryStats() QueryStatsSnapshot` | 获取查询统计 |
| `GetProviderStatus() map[string]bool` | 检查数据源状态 |
| `GetProviderInfo() map[string]ProviderInfo` | 获取数据源详情 |
| `ClearCache()` | 清空缓存 |
| `ResetStats()` | 重置统计计数器 |
| `Close() error` | 释放资源 |

### 返回结构

```go
type Location struct {
    IP        string  // IP地址
    Country   string  // 国家
    Province  string  // 省/州
    City      string  // 城市
    District  string  // 区/县
    ISP       string  // 运营商
    Latitude  float64 // 纬度
    Longitude float64 // 经度
    TimeZone  string  // 时区
    Source    string  // 数据来源 (qqwry/geolite2/combined)
}
```

### 统计数据

```go
type QueryStatsSnapshot struct {
    TotalQueries   int64         // 总查询次数
    SuccessQueries int64         // 成功次数
    FailedQueries  int64         // 失败次数
    QQwryHits      int64         // QQwry使用次数
    GeoLiteHits    int64         // GeoLite使用次数
    CombinedHits   int64         // 数据合并次数
    AvgDuration    time.Duration // 平均查询时间
    SuccessRate    float64       // 成功率
}
```

## 高级功能

### 性能监控

```go
stats := locator.GetQueryStats()
fmt.Printf("查询: %d次 | 成功率: %.2f%% | 平均延迟: %v\n",
    stats.TotalQueries,
    stats.SuccessRate,
    stats.AvgDuration,
)
```

### 缓存统计

```go
cacheStats := locator.GetCacheStats()
fmt.Printf("缓存: %d/%d | 命中率: %.2f%%\n",
    cacheStats.Size,
    cacheStats.Capacity,
    cacheStats.HitRate,
)
```

### 数据源状态检查

```go
status := locator.GetProviderStatus()
fmt.Printf("QQwry: %v | GeoLite: %v\n",
    status["qqwry"],
    status["geolite"],
)

// 获取详细错误信息
info := locator.GetProviderInfo()
for name, provInfo := range info {
    if !provInfo.Available && len(provInfo.Errors) > 0 {
        fmt.Printf("%s 加载失败: %v\n", name, provInfo.Errors)
    }
}
```

### 调试模式

```go
locator, _ := iplocx.NewLocator(iplocx.Config{
    QQwryDBPath:   "./data/qqwry.dat",
    GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
    Debug:         true,
})
```

调试模式输出数据合并过程的详细信息，包括各数据源的评分和字段选择逻辑。

## 性能基准测试

### 单核性能

```
BenchmarkQQwryOnly-32      2,707,208 ops    2,173 ns/op
BenchmarkGeoLiteOnly-32    3,070,021 ops    1,942 ns/op
BenchmarkQuery-32            585,474 ops    9,573 ns/op
```

### 多核扩展性

```
核心数    QPS          延迟        扩展倍数
1核      174,895      6.32 μs     1.00x
2核      357,901      3.32 μs     2.05x
4核      634,393      1.96 μs     3.63x
8核      908,860      1.31 μs     5.20x
16核     726,627      1.67 μs     4.15x
32核     553,378      2.15 μs     3.16x
```

8核配置达到最佳性能平衡点。

### 缓存性能

```
BenchmarkCacheGet-32    521,265,956 ops    11.45 ns/op
```

缓存命中后延迟降低至 11.45 纳秒，吞吐量超过 1 亿 QPS。

### 运行基准测试

```bash
# 完整性能测试
go test -bench=Benchmark -benchmem -benchtime=5s

# 多核扩展性测试
go test -bench=BenchmarkQueryParallel -benchmem -benchtime=5s -cpu="1,2,4,8,16,32"

# 单项测试
go test -bench=BenchmarkQuery$ -benchmem -benchtime=5s
```

## 内存占用

- QQwry 数据库：~30 MB
- GeoLite2 数据库：~80 MB
- LRU 缓存：~200 KB / 1000条记录

## 数据库文件

### QQwry（纯真IP库）

- 下载：https://raw.githubusercontent.com/FW27623/qqwry/refs/heads/main/qqwry.dat
- 协议：IPv4
- 优势：国内IP信息准确，包含运营商和区县级信息

### GeoLite2

- 下载：https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
- 协议：IPv4 / IPv6
- 优势：国际IP信息准确，包含经纬度和时区信息

注：GeoLite2 需要注册 MaxMind 账号下载。

## 应用场景

- Web 应用地理位置识别
- 访问日志分析
- CDN 内容分发调度
- 安全审计与反欺诈
- 地域访问控制
- 广告定向投放

## 测试

```bash
# 运行所有测试
go test -v

# 查看覆盖率
go test -cover

# 运行基准测试
go test -bench=. -benchmem

# 运行特定测试
go test -v -run TestCache
```

## 示例程序

项目包含多个示例程序，位于 `examples/` 目录：

- `basic/` - 基础查询功能
- `with_cache/` - 缓存性能对比
- `stats_monitor/` - 统计监控功能
- `complete_test/` - 完整功能测试（12个测试模块）

运行示例：

```bash
cd examples/complete_test
go run main.go
```

## 错误处理

```go
var (
    ErrDatabaseNotFound  // 数据库文件未找到
    ErrInvalidIP         // 无效的IP地址
    ErrNoData            // 未找到数据
    ErrNoProvider        // 没有可用的查询提供者
)
```

## 贡献

我们欢迎并感谢所有形式的贡献！

### 如何贡献

- 🐛 [报告 Bug](https://github.com/nuomiaa/iplocx/issues/new?template=bug_report.md)
- 💡 [提出新功能](https://github.com/nuomiaa/iplocx/issues/new?template=feature_request.md)
- 📖 改进文档
- 🔧 提交代码

详见 [贡献指南](CONTRIBUTING.md)。

### 贡献者

感谢所有为这个项目做出贡献的人！

<a href="https://github.com/nuomiaa/iplocx/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=nuomiaa/iplocx" />
</a>

## 社区

### 讨论与支持

- 💬 [GitHub Discussions](https://github.com/nuomiaa/iplocx/discussions) - 提问和讨论
- 🐛 [Issue Tracker](https://github.com/nuomiaa/iplocx/issues) - Bug 报告和功能请求
- 📧 [Email](mailto:nuomiaa@gmail.com) - 直接联系维护者

### Star History

[![Star History Chart](https://api.star-history.com/svg?repos=nuomiaa/iplocx&type=Date)](https://star-history.com/#nuomiaa/iplocx&Date)

## 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可证。

```
Copyright 2024 nuomiaa

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
```

Apache 2.0 许可证特性：
- ✅ 商业使用
- ✅ 修改和分发
- ✅ 专利授权
- ✅ 私有使用
- ℹ️ 需保留版权和许可声明
- ℹ️ 需说明修改内容

## 致谢

### 数据源

- [纯真IP库](http://www.cz88.net/) - 提供国内IP地理位置数据
- [MaxMind GeoLite2](https://www.maxmind.com/) - 提供全球IP地理位置数据

### 依赖项目

- [oschwald/geoip2-golang](https://github.com/oschwald/geoip2-golang) - GeoIP2 数据库读取
- [golang.org/x/text](https://pkg.go.dev/golang.org/x/text) - 文本编码支持

### 相关项目

- [qqwry](https://github.com/FW27623/qqwry) - 纯真IP数据库维护
- [GeoIP2-CN](https://github.com/Hackl0us/GeoIP2-CN) - 中国IP数据库

---

<div align="center">

### ⭐ 如果这个项目对你有帮助，请给一个 Star！⭐

[![GitHub stars](https://img.shields.io/github/stars/nuomiaa/iplocx?style=social)](https://github.com/nuomiaa/iplocx/stargazers)

**Made with ❤️ by [nuomiaa](https://github.com/nuomiaa)**

[⬆ 回到顶部](#iplocx)

</div>
