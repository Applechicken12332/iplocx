package iplocx_test

import (
	"fmt"
	"log"

	"github.com/nuomiaa/iplocx"
)

// Example_stats 统计功能示例
func Example_stats() {
	cfg := iplocx.Config{
		QQwryDBPath:   "./data/qqwry.dat",
		GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
		EnableCache:   true,
	}

	locator, err := iplocx.NewLocator(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer locator.Close()

	// 执行多次查询
	testIPs := []string{
		"8.8.8.8",
		"124.224.56.76",
		"1.1.1.1",
		"8.8.8.8", // 重复查询，会命中缓存
	}

	for _, ip := range testIPs {
		_, _ = locator.Query(ip)
	}

	// 获取查询统计
	stats := locator.GetQueryStats()
	fmt.Printf("总查询: %d\n", stats.TotalQueries)
	fmt.Printf("合并数据使用: %d次\n", stats.CombinedHits)

	// 获取缓存统计
	if cacheStats := locator.GetCacheStats(); cacheStats != nil {
		fmt.Printf("缓存命中: %d次\n", cacheStats.Hits)
		fmt.Printf("缓存大小: %d条\n", cacheStats.Size)
	}

	// Output:
	// 总查询: 4
	// 合并数据使用: 3次
	// 缓存命中: 1次
	// 缓存大小: 3条
}

// Example_providerInfo 数据源详细信息示例
func Example_providerInfo() {
	cfg := iplocx.Config{
		QQwryDBPath:   "./data/qqwry.dat",
		GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
	}

	locator, err := iplocx.NewLocator(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer locator.Close()

	// 获取详细数据源信息（按固定顺序输出）
	info := locator.GetProviderInfo()

	// 按固定顺序输出
	for _, name := range []string{"qqwry", "geolite"} {
		provInfo := info[name]
		fmt.Printf("%s: 可用=%v\n", name, provInfo.Available)
		if len(provInfo.Errors) > 0 {
			fmt.Printf("  错误: %v\n", provInfo.Errors)
		}
	}
	// Output:
	// qqwry: 可用=true
	// geolite: 可用=true
}
