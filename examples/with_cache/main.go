package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nuomiaa/iplocx" // 发布时替换为实际的 GitHub 路径
)

func main() {
	// 启用缓存的配置
	cfg := iplocx.Config{
		QQwryDBPath:   "../../data/qqwry.dat",
		GeoLiteDBPath: "../../data/GeoLite2-City.mmdb",
		EnableCache:   true,
		CacheSize:     1000,
	}

	locator, err := iplocx.NewLocator(cfg)
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer locator.Close()

	fmt.Println("========================================")
	fmt.Println("  缓存性能测试")
	fmt.Println("========================================")
	fmt.Println()

	testIP := "8.8.8.8"

	// 第一次查询（无缓存）
	start := time.Now()
	_, err = locator.Query(testIP)
	duration1 := time.Since(start)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("第一次查询 %s: %v (从数据库)\n", testIP, duration1)

	// 第二次查询（有缓存）
	start = time.Now()
	location, err := locator.Query(testIP)
	duration2 := time.Since(start)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("第二次查询 %s: %v (从缓存)\n\n", testIP, duration2)

	// 显示结果
	fmt.Printf("查询结果: %s, %s\n", location.Country, location.City)
	fmt.Printf("性能提升: %.2fx\n\n", float64(duration1)/float64(duration2))

	// 显示缓存统计
	if stats := locator.GetCacheStats(); stats != nil {
		fmt.Println("缓存统计:")
		fmt.Printf("  缓存大小: %d/%d\n", stats.Size, stats.Capacity)
		fmt.Printf("  命中次数: %d\n", stats.Hits)
		fmt.Printf("  未命中次数: %d\n", stats.Misses)
		fmt.Printf("  命中率: %.2f%%\n", stats.HitRate)
	}
}
