package main

import (
	"fmt"
	"log"

	"github.com/nuomiaa/iplocx" // 发布时替换为实际的 GitHub 路径
)

func main() {
	cfg := iplocx.Config{
		QQwryDBPath:   "../../data/qqwry.dat",
		GeoLiteDBPath: "../../data/GeoLite2-City.mmdb",
		EnableCache:   true,
		CacheSize:     500,
	}

	locator, err := iplocx.NewLocator(cfg)
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer locator.Close()

	// 模拟一批查询
	testIPs := []string{
		"8.8.8.8", "1.1.1.1", "124.224.56.76", "114.114.114.114",
		"8.8.8.8", // 重复查询
		"180.89.94.90", "211.90.237.22",
		"8.8.8.8", // 再次重复
	}

	fmt.Println("========================================")
	fmt.Println("  性能监控示例")
	fmt.Println("========================================")
	fmt.Println()

	fmt.Println("执行查询...")
	for i, ip := range testIPs {
		_, err := locator.Query(ip)
		if err != nil {
			fmt.Printf("[%d] %s - 失败: %v\n", i+1, ip, err)
		} else {
			fmt.Printf("[%d] %s - 成功\n", i+1, ip)
		}
	}

	// 显示完整统计
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  统计报告")
	fmt.Println("========================================")
	fmt.Println()

	// 查询统计
	stats := locator.GetQueryStats()
	fmt.Println("📊 查询统计:")
	fmt.Printf("  总查询: %d 次\n", stats.TotalQueries)
	fmt.Printf("  成功: %d 次 (%.2f%%)\n", stats.SuccessQueries, stats.SuccessRate)
	fmt.Printf("  失败: %d 次\n", stats.FailedQueries)
	fmt.Printf("  平均耗时: %v\n\n", stats.AvgDuration)

	fmt.Println("📈 数据源使用:")
	fmt.Printf("  QQwry 独占: %d 次\n", stats.QQwryHits)
	fmt.Printf("  GeoLite 独占: %d 次\n", stats.GeoLiteHits)
	fmt.Printf("  智能合并: %d 次\n\n", stats.CombinedHits)

	// 缓存统计
	if cacheStats := locator.GetCacheStats(); cacheStats != nil {
		fmt.Println("💾 缓存统计:")
		fmt.Printf("  缓存大小: %d/%d\n", cacheStats.Size, cacheStats.Capacity)
		fmt.Printf("  命中: %d 次\n", cacheStats.Hits)
		fmt.Printf("  未命中: %d 次\n", cacheStats.Misses)
		fmt.Printf("  命中率: %.2f%%\n\n", cacheStats.HitRate)
	}

	// 数据源状态
	providerInfo := locator.GetProviderInfo()
	fmt.Println("🔌 数据源状态:")
	for _, name := range []string{"qqwry", "geolite"} {
		info := providerInfo[name]
		status := "❌ 不可用"
		if info.Available {
			status = "✅ 可用"
		}
		fmt.Printf("  %s: %s\n", name, status)
		if len(info.Errors) > 0 {
			for _, errMsg := range info.Errors {
				fmt.Printf("    错误: %s\n", errMsg)
			}
		}
	}
}
