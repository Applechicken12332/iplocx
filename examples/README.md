# iplocx 示例程序

本目录包含了 iplocx 包的各种使用示例。

## 📁 目录结构

- **basic/** - 基础查询示例
- **with_cache/** - 缓存性能测试
- **stats_monitor/** - 性能监控和统计
- **complete_test/** - 完整功能测试（推荐）⭐

## 🚀 运行示例

### 前提条件

1. 下载数据库文件：
   - QQwry: http://update.cz88.net/soft/qqwry.rar
   - GeoLite2: https://dev.maxmind.com/geoip/geolite2-free-geolocation-data

2. 将数据库文件放置在正确位置：
   ```
   iptest/
   ├── data/
       └── qqwry.dat
       └── GeoLite2-City.mmdb
   ```

### 基础示例

```bash
cd basic
go run main.go
```

展示基本的 IP 查询功能，包括国内外 IP 的查询。

### 缓存示例

```bash
cd with_cache
go run main.go
```

演示缓存对性能的提升，对比有无缓存的查询速度。

### 统计监控示例

```bash
cd stats_monitor
go run main.go
```

展示完整的统计功能，包括：
- 查询成功率
- 数据源使用情况
- 缓存命中率
- 平均查询时间

### 完整功能测试 ⭐（推荐）

```bash
cd complete_test
go run main.go
```

**这是最全面的测试示例，涵盖所有功能！**

包含 12 个完整测试模块：
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

## 💡 更多示例

查看主包的 `example_test.go` 文件，里面有更多可以在 godoc 中展示的示例。

