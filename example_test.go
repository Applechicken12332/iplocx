package iplocx_test

import (
	"fmt"
	"log"

	"github.com/nuomiaa/iplocx"
)

// Example_basic 基本使用示例
func Example_basic() {
	cfg := iplocx.Config{
		QQwryDBPath:   "./data/qqwry.dat",
		GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
	}

	locator, err := iplocx.NewLocator(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer locator.Close()

	location, err := locator.Query("8.8.8.8")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("国家: %s\n", location.Country)
	fmt.Printf("城市: %s\n", location.City)
	// Output:
	// 国家: 美国
	// 城市: 圣克拉拉
}

// Example_withCache 使用缓存示例
func Example_withCache() {
	cfg := iplocx.Config{
		QQwryDBPath:   "./data/qqwry.dat",
		GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
		EnableCache:   true,
		CacheSize:     500,
	}

	locator, err := iplocx.NewLocator(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer locator.Close()

	// 第一次查询（从数据库）
	_, _ = locator.Query("8.8.8.8")

	// 第二次查询（从缓存）
	location, _ := locator.Query("8.8.8.8")

	fmt.Printf("国家: %s\n", location.Country)

	// 查看缓存统计
	if stats := locator.GetCacheStats(); stats != nil {
		fmt.Printf("缓存命中率: %.1f%%\n", stats.HitRate)
	}
	// Output:
	// 国家: 美国
	// 缓存命中率: 50.0%
}

// Example_debugMode 调试模式示例
func Example_debugMode() {
	cfg := iplocx.Config{
		QQwryDBPath:   "./data/qqwry.dat",
		GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
		Debug:         true, // 启用调试输出
	}

	locator, err := iplocx.NewLocator(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer locator.Close()

	_, _ = locator.Query("124.224.56.76")
	// 会输出详细的数据合并过程
}

// Example_providerStatus 检查数据源状态
func Example_providerStatus() {
	cfg := iplocx.Config{
		QQwryDBPath: "./data/qqwry.dat",
		// 未配置 GeoLite
	}

	locator, err := iplocx.NewLocator(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer locator.Close()

	status := locator.GetProviderStatus()
	fmt.Printf("QQwry 可用: %v\n", status["qqwry"])
	fmt.Printf("GeoLite 可用: %v\n", status["geolite"])
	// Output:
	// QQwry 可用: true
	// GeoLite 可用: false
}
