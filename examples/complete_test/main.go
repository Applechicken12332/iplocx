package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nuomiaa/iplocx"
)

// formatDuration 格式化时间，自动选择合适的单位
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "< 1 ns"
	}
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2f μs", float64(d.Nanoseconds())/1000.0)
	} else if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2f s", d.Seconds())
}

// measureQuery 测量查询时间（多次测量取平均值以提高精度）
func measureQuery(locator *iplocx.Locator, ip string, iterations int) time.Duration {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = locator.Query(ip)
	}
	totalDuration := time.Since(start)
	return totalDuration / time.Duration(iterations)
}

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║        iplocx 完整功能测试示例程序                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ============================================================
	// 1. 基础初始化测试
	// ============================================================
	fmt.Println("【1. 基础初始化测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	cfg := iplocx.Config{
		QQwryDBPath:   "../../data/qqwry.dat",
		GeoLiteDBPath: "../../data/GeoLite2-City.mmdb",
		Debug:         false, // 先关闭调试，后面会测试
		EnableCache:   true,
		CacheSize:     100,
	}

	locator, err := iplocx.NewLocator(cfg)
	if err != nil {
		log.Fatalf("❌ 初始化失败: %v", err)
	}
	defer locator.Close()

	fmt.Println("✅ Locator 初始化成功")

	// 检查数据源状态
	status := locator.GetProviderStatus()
	fmt.Printf("✅ QQwry 数据源: %v\n", status["qqwry"])
	fmt.Printf("✅ GeoLite 数据源: %v\n", status["geolite"])

	// 获取详细数据源信息
	info := locator.GetProviderInfo()
	for name, provInfo := range info {
		if !provInfo.Available {
			fmt.Printf("⚠️  %s 不可用", name)
			if len(provInfo.Errors) > 0 {
				fmt.Printf(": %v", provInfo.Errors)
			}
			fmt.Println()
		}
	}
	fmt.Println()

	// ============================================================
	// 2. IPv4 国内 IP 测试（详细信息）
	// ============================================================
	fmt.Println("【2. IPv4 国内 IP 测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	chinaIPs := []string{
		"124.224.56.76",   // 中国 IP
		"114.114.114.114", // 中国 DNS
		"223.5.5.5",       // 阿里 DNS
		"119.75.217.109",  // 中国电信
	}

	for _, ip := range chinaIPs {
		fmt.Printf("\n查询 IP: %s\n", ip)
		location, err := locator.Query(ip)
		if err != nil {
			fmt.Printf("  ❌ 查询失败: %v\n", err)
			continue
		}
		printLocation(location)
	}
	fmt.Println()

	// ============================================================
	// 3. IPv4 国外 IP 测试
	// ============================================================
	fmt.Println("【3. IPv4 国外 IP 测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	foreignIPs := []string{
		"8.8.8.8",        // Google DNS (美国)
		"1.1.1.1",        // Cloudflare DNS
		"208.67.222.222", // OpenDNS
		"9.9.9.9",        // Quad9 DNS
	}

	for _, ip := range foreignIPs {
		fmt.Printf("\n查询 IP: %s\n", ip)
		location, err := locator.Query(ip)
		if err != nil {
			fmt.Printf("  ❌ 查询失败: %v\n", err)
			continue
		}
		printLocation(location)
	}
	fmt.Println()

	// ============================================================
	// 4. IPv6 地址测试
	// ============================================================
	fmt.Println("【4. IPv6 地址测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	ipv6IPs := []string{
		"2400:3200::1",                          // 阿里云 DNS (中国)
		"240e:446:2b02:5ab:a8b4:8948:125e:5140", // 中国电信 IPv6
		"2001:4860:4860::8888",                  // Google DNS IPv6
	}

	for _, ip := range ipv6IPs {
		fmt.Printf("\n查询 IP: %s\n", ip)
		location, err := locator.Query(ip)
		if err != nil {
			fmt.Printf("  ⚠️  查询失败: %v (QQwry 不支持 IPv6)\n", err)
			continue
		}
		printLocation(location)
	}
	fmt.Println()

	// ============================================================
	// 5. 缓存功能测试
	// ============================================================
	fmt.Println("【5. 缓存功能测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	testIP := "8.8.8.8"

	// 清空缓存，从头开始
	locator.ClearCache()
	fmt.Println("✅ 缓存已清空")

	// 第一次查询（不命中缓存）- 测量10次取平均
	firstDuration := measureQuery(locator, testIP, 10)
	fmt.Printf("首次查询 %s (10次平均): 耗时 %s (缓存未命中)\n", testIP, formatDuration(firstDuration))

	// 缓存命中测试（多次测量取平均值）
	fmt.Printf("\n缓存命中性能测试:\n")

	// 预热一下缓存
	for i := 0; i < 100; i++ {
		_, _ = locator.Query(testIP)
	}

	// 测量缓存命中性能
	cachedDuration := measureQuery(locator, testIP, 10000)
	fmt.Printf("  缓存命中 (10000次平均): 耗时 %s\n", formatDuration(cachedDuration))

	// 显示缓存统计
	cacheStats := locator.GetCacheStats()
	if cacheStats != nil {
		fmt.Printf("\n缓存统计信息:\n")
		fmt.Printf("  当前大小: %d/%d\n", cacheStats.Size, cacheStats.Capacity)
		fmt.Printf("  命中次数: %d\n", cacheStats.Hits)
		fmt.Printf("  未命中次数: %d\n", cacheStats.Misses)
		fmt.Printf("  命中率: %.2f%%\n", cacheStats.HitRate)
		fmt.Printf("\n性能对比:\n")
		fmt.Printf("  无缓存查询: %s\n", formatDuration(firstDuration))
		fmt.Printf("  缓存命中: %s\n", formatDuration(cachedDuration))
		if cachedDuration > 0 && firstDuration > 0 {
			speedup := float64(firstDuration) / float64(cachedDuration)
			fmt.Printf("  性能提升: %.0fx 倍\n", speedup)
		}
	}
	fmt.Println()

	// ============================================================
	// 6. 批量查询性能测试
	// ============================================================
	fmt.Println("【6. 批量查询性能测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 重置统计，开始新测试
	locator.ResetStats()
	locator.ClearCache()

	testIPs := []string{
		"8.8.8.8",
		"1.1.1.1",
		"124.224.56.76",
		"114.114.114.114",
		"223.5.5.5",
	}

	fmt.Printf("批量查询 %d 个不同 IP，每个查询 20 次 (含缓存)...\n", len(testIPs))
	start := time.Now()

	for i := 0; i < 20; i++ {
		for _, ip := range testIPs {
			_, _ = locator.Query(ip)
		}
	}

	totalDuration := time.Since(start)
	totalQueries := len(testIPs) * 20

	fmt.Printf("\n批量查询完成:\n")
	fmt.Printf("  总查询数: %d\n", totalQueries)
	fmt.Printf("  总耗时: %s\n", formatDuration(totalDuration))
	fmt.Printf("  平均每次查询: %s\n", formatDuration(totalDuration/time.Duration(totalQueries)))
	if totalDuration.Seconds() > 0 {
		fmt.Printf("  QPS: %.0f 次/秒\n", float64(totalQueries)/totalDuration.Seconds())
	} else {
		fmt.Printf("  QPS: > 1,000,000 次/秒 (查询太快无法精确测量)\n")
	}
	fmt.Println()

	// ============================================================
	// 7. 查询统计信息测试
	// ============================================================
	fmt.Println("【7. 查询统计信息测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	stats := locator.GetQueryStats()
	fmt.Printf("查询统计 (本轮批量测试):\n")
	fmt.Printf("  总查询次数: %d\n", stats.TotalQueries)
	fmt.Printf("  成功次数: %d (含缓存命中)\n", stats.SuccessQueries)
	fmt.Printf("  失败次数: %d\n", stats.FailedQueries)
	fmt.Printf("  成功率: %.2f%%\n", stats.SuccessRate)
	fmt.Printf("  平均查询时间: %s\n\n", formatDuration(stats.AvgDuration))

	fmt.Printf("数据源使用统计 (仅首次查询):\n")
	fmt.Printf("  QQwry 单独使用: %d 次\n", stats.QQwryHits)
	fmt.Printf("  GeoLite 单独使用: %d 次\n", stats.GeoLiteHits)
	fmt.Printf("  数据合并使用: %d 次\n", stats.CombinedHits)

	// 计算数据源使用比例
	totalHits := stats.QQwryHits + stats.GeoLiteHits + stats.CombinedHits
	if totalHits > 0 {
		fmt.Printf("\n数据源使用比例 (首次查询时):\n")
		fmt.Printf("  QQwry: %.1f%%\n", float64(stats.QQwryHits)/float64(totalHits)*100)
		fmt.Printf("  GeoLite: %.1f%%\n", float64(stats.GeoLiteHits)/float64(totalHits)*100)
		fmt.Printf("  合并: %.1f%%\n", float64(stats.CombinedHits)/float64(totalHits)*100)
	}

	fmt.Printf("\n说明: 本测试在批量查询前重置了统计，只统计了这100次查询\n")
	fmt.Printf("      (5个IP首次查询 + 95次缓存命中 = 100次总查询)\n")
	fmt.Println()

	// ============================================================
	// 8. 错误处理测试
	// ============================================================
	fmt.Println("【8. 错误处理测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	invalidIPs := []string{
		"invalid-ip",
		"999.999.999.999",
		"",
		"abc.def.ghi.jkl",
	}

	for _, ip := range invalidIPs {
		fmt.Printf("查询无效 IP: '%s'\n", ip)
		_, err := locator.Query(ip)
		if err != nil {
			fmt.Printf("  ✅ 正确处理错误: %v\n", err)
		} else {
			fmt.Printf("  ⚠️  应该返回错误但没有\n")
		}
	}
	fmt.Println()

	// ============================================================
	// 9. Location 方法测试
	// ============================================================
	fmt.Println("【9. Location 方法测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	loc, _ := locator.Query("124.224.56.76")
	if loc != nil {
		fmt.Printf("IP: %s\n", loc.IP)
		fmt.Printf("  IsEmpty(): %v\n", loc.IsEmpty())
		fmt.Printf("  HasDetailedInfo(): %v\n", loc.HasDetailedInfo())
		fmt.Printf("  GetDetailScore(): %d 分\n", loc.GetDetailScore())
		fmt.Printf("  评分说明: 国家(1分) + 省(2分) + 市(4分) + 区(8分)\n")
	}
	fmt.Println()

	// ============================================================
	// 10. 调试模式测试
	// ============================================================
	fmt.Println("【10. 调试模式测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 创建带调试模式的新实例
	debugCfg := iplocx.Config{
		QQwryDBPath:   "../../data/qqwry.dat",
		GeoLiteDBPath: "../../data/GeoLite2-City.mmdb",
		Debug:         true, // 开启调试
		EnableCache:   false,
	}

	debugLocator, err := iplocx.NewLocator(debugCfg)
	if err != nil {
		log.Printf("⚠️  调试模式初始化失败: %v", err)
	} else {
		defer debugLocator.Close()
		fmt.Println("启用调试模式，查询 8.8.8.8:")
		fmt.Println("────────────────────────────────────────────────────────────")
		_, _ = debugLocator.Query("8.8.8.8")
		fmt.Println("────────────────────────────────────────────────────────────")
	}
	fmt.Println()

	// ============================================================
	// 11. 并发安全测试
	// ============================================================
	fmt.Println("【11. 并发安全测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	concurrentIPs := []string{
		"8.8.8.8",
		"1.1.1.1",
		"124.224.56.76",
	}

	locator.ResetStats()
	start = time.Now()

	// 启动多个 goroutine 并发查询
	done := make(chan bool)
	goroutineCount := 10
	queriesPerGoroutine := 100

	for i := 0; i < goroutineCount; i++ {
		go func(id int) {
			for j := 0; j < queriesPerGoroutine; j++ {
				ip := concurrentIPs[j%len(concurrentIPs)]
				_, _ = locator.Query(ip)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < goroutineCount; i++ {
		<-done
	}

	concurrentDuration := time.Since(start)
	totalConcurrentQueries := goroutineCount * queriesPerGoroutine

	fmt.Printf("并发测试完成:\n")
	fmt.Printf("  Goroutine 数量: %d\n", goroutineCount)
	fmt.Printf("  每个 Goroutine 查询: %d 次\n", queriesPerGoroutine)
	fmt.Printf("  总查询数: %d\n", totalConcurrentQueries)
	fmt.Printf("  总耗时: %s\n", formatDuration(concurrentDuration))
	fmt.Printf("  并发 QPS: %.0f 次/秒\n", float64(totalConcurrentQueries)/concurrentDuration.Seconds())

	stats = locator.GetQueryStats()
	fmt.Printf("  成功率: %.2f%%\n", stats.SuccessRate)

	cacheStats = locator.GetCacheStats()
	if cacheStats != nil {
		fmt.Printf("  缓存命中率: %.2f%%\n", cacheStats.HitRate)
	}
	fmt.Println()

	// ============================================================
	// 12. 资源清理测试
	// ============================================================
	fmt.Println("【12. 资源清理测试】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 清空缓存
	locator.ClearCache()
	cacheStats = locator.GetCacheStats()
	if cacheStats != nil && cacheStats.Size == 0 {
		fmt.Println("✅ 缓存清理成功")
	}

	// 重置统计
	locator.ResetStats()
	stats = locator.GetQueryStats()
	if stats.TotalQueries == 0 {
		fmt.Println("✅ 统计重置成功")
	}

	fmt.Println("✅ 准备关闭 Locator...")
	err = locator.Close()
	if err != nil {
		fmt.Printf("⚠️  关闭时出现错误: %v\n", err)
	} else {
		fmt.Println("✅ Locator 关闭成功")
	}
	fmt.Println()

	// ============================================================
	// 测试总结
	// ============================================================
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   测试完成！                               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

// printLocation 格式化打印位置信息
func printLocation(loc *iplocx.Location) {
	if loc == nil {
		fmt.Println("  ⚠️  位置信息为空")
		return
	}

	// 基础地理信息
	if loc.Country != "" {
		fmt.Printf("  🌍 国家: %s\n", loc.Country)
	}
	if loc.Province != "" {
		fmt.Printf("  🏛️  省/州: %s\n", loc.Province)
	}
	if loc.City != "" {
		fmt.Printf("  🏙️  城市: %s\n", loc.City)
	}
	if loc.District != "" {
		fmt.Printf("  📍 区/县: %s\n", loc.District)
	}

	// 运营商信息
	if loc.ISP != "" {
		fmt.Printf("  🌐 运营商: %s\n", loc.ISP)
	}

	// 地理坐标
	if loc.Latitude != 0 || loc.Longitude != 0 {
		fmt.Printf("  📌 坐标: %.4f, %.4f\n", loc.Latitude, loc.Longitude)
	}

	// 时区
	if loc.TimeZone != "" {
		fmt.Printf("  🕐 时区: %s\n", loc.TimeZone)
	}

	// 数据来源
	fmt.Printf("  📊 数据源: %s", loc.Source)
	switch loc.Source {
	case "qqwry":
		fmt.Print(" (纯真IP库)")
	case "geolite2":
		fmt.Print(" (GeoLite2)")
	case "combined":
		fmt.Print(" (智能合并)")
	}
	fmt.Println()

	// 详细程度评分
	score := loc.GetDetailScore()
	fmt.Printf("  ⭐ 详细度: %d 分", score)
	if score >= 10 {
		fmt.Print(" (非常详细)")
	} else if score >= 7 {
		fmt.Print(" (详细)")
	} else if score >= 3 {
		fmt.Print(" (一般)")
	} else {
		fmt.Print(" (简单)")
	}
	fmt.Println()
}
