# 示例程序

本目录包含 iplocx 的使用示例，展示库的主要功能和典型用法。

## 目录结构

```
examples/
├── README.md           # 本文件
├── basic/              # 基础查询示例
├── with_cache/         # 缓存性能演示
├── stats_monitor/      # 统计监控功能
└── complete_test/      # 完整功能测试（推荐）
```

## 运行前准备

### 1. 下载数据库文件

**QQwry（纯真IP库）**
```bash
curl -o ../data/qqwry.dat https://raw.githubusercontent.com/FW27623/qqwry/refs/heads/main/qqwry.dat
```

**GeoLite2**

访问 https://dev.maxmind.com/geoip/geolite2-free-geolocation-data

注册账号后下载 `GeoLite2-City.mmdb`，放置到 `../data/` 目录。

### 2. 目录结构

确保文件结构如下：

```
iplocx/
├── data/
│   ├── qqwry.dat
│   └── GeoLite2-City.mmdb
└── examples/
    └── ...
```

## 示例说明

### basic - 基础查询

```bash
cd basic && go run main.go
```

演示内容：
- 基本的 IP 查询功能
- 查询国内和国际 IP
- 显示完整的地理位置信息

适合初次使用。

### with_cache - 缓存性能

```bash
cd with_cache && go run main.go
```

演示内容：
- 无缓存 vs 有缓存性能对比
- 缓存命中率统计
- 性能提升数据

输出示例：
```
无缓存查询: 9.57 μs
缓存命中查询: 11.45 ns
性能提升: 836 倍
```

### stats_monitor - 统计监控

```bash
cd stats_monitor && go run main.go
```

演示内容：
- 查询统计（总数、成功率、平均延迟）
- 数据源使用情况
- 缓存统计信息
- 实时性能监控

适合生产环境监控参考。

### complete_test - 完整功能测试（推荐）

```bash
cd complete_test && go run main.go
```

这是最全面的测试示例，包含 12 个测试模块：

1. **基础初始化** - 数据源状态检查
2. **IPv4 国内 IP** - 查询国内 IP 详细信息
3. **IPv4 国外 IP** - 查询国际 IP
4. **IPv6 查询** - IPv6 地址支持
5. **缓存功能** - LRU 缓存测试和性能对比
6. **批量查询** - 批量查询性能测试
7. **统计功能** - 查询统计展示
8. **错误处理** - 异常情况处理
9. **辅助方法** - Location 辅助方法测试
10. **调试模式** - 数据合并过程可视化
11. **并发安全** - 多 Goroutine 并发测试
12. **资源清理** - 正确释放资源

适用场景：
- 首次使用，全面了解功能
- 验证环境配置
- 性能基准参考
- 开发调试

## 代码示例

### 最小示例

```go
package main

import (
    "fmt"
    "github.com/nuomiaa/iplocx"
)

func main() {
    locator, _ := iplocx.NewLocator(iplocx.Config{
        QQwryDBPath: "../data/qqwry.dat",
    })
    defer locator.Close()

    location, _ := locator.Query("8.8.8.8")
    fmt.Printf("%s: %s, %s\n", 
        location.IP, 
        location.Country, 
        location.City,
    )
}
```

### 启用缓存

```go
locator, _ := iplocx.NewLocator(iplocx.Config{
    QQwryDBPath:   "../data/qqwry.dat",
    GeoLiteDBPath: "../data/GeoLite2-City.mmdb",
    EnableCache:   true,
    CacheSize:     10000,
})
```

### 性能监控

```go
stats := locator.GetQueryStats()
fmt.Printf("查询: %d | 成功率: %.2f%% | 延迟: %v\n",
    stats.TotalQueries,
    stats.SuccessRate,
    stats.AvgDuration,
)
```

## 性能数据

基于示例程序的实际测试结果（AMD Ryzen 9 7945HX）：

| 场景 | 延迟 | 吞吐量 |
|------|------|--------|
| 首次查询 | 9.57 μs | 117,000 QPS |
| 缓存命中 | 11.45 ns | 104,000,000 QPS |
| 批量查询（1000次） | 平均 9.6 μs | - |

## 常见问题

### 数据库文件未找到

确保数据库文件路径正确：
- 相对路径：`../data/qqwry.dat`
- 绝对路径：`/path/to/data/qqwry.dat`

### GeoLite2 下载需要账号

GeoLite2 需要在 MaxMind 官网注册免费账号才能下载。

### 性能结果与文档不符

性能数据会因硬件配置而异，文档中的数据基于特定测试环境。

## 更多示例

查看主包的 `example_test.go` 文件，包含更多 GoDoc 可见的示例代码。

## 反馈

如有问题或建议，欢迎提交 [Issue](https://github.com/nuomiaa/iplocx/issues)。
