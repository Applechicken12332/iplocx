package main

import (
	"fmt"
	"log"

	"github.com/nuomiaa/iplocx" // 发布时替换为实际的 GitHub 路径
)

func main() {
	// 基本配置
	cfg := iplocx.Config{
		QQwryDBPath:   "../../data/qqwry.dat", // 相对于示例目录
		GeoLiteDBPath: "../../data/GeoLite2-City.mmdb",
	}

	// 创建查询器
	locator, err := iplocx.NewLocator(cfg)
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer locator.Close()

	// 查询 IP
	testIPs := []string{
		"8.8.8.8",
		"1.1.1.1",
		"124.224.56.76",
	}

	fmt.Println("========================================")
	fmt.Println("  基础 IP 查询示例")
	fmt.Println("========================================")
	fmt.Println()

	for _, ip := range testIPs {
		fmt.Printf("查询 IP: %s\n", ip)
		fmt.Println("----------------------------------------")

		location, err := locator.Query(ip)
		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n\n", err)
			continue
		}

		fmt.Printf("🌍 国家: %s\n", location.Country)
		if location.Province != "" {
			fmt.Printf("🏛️  省/州: %s\n", location.Province)
		}
		if location.City != "" {
			fmt.Printf("🏙️  城市: %s\n", location.City)
		}
		if location.ISP != "" {
			fmt.Printf("🌐 运营商: %s\n", location.ISP)
		}
		if location.Latitude != 0 || location.Longitude != 0 {
			fmt.Printf("📌 经纬度: %.4f, %.4f\n", location.Latitude, location.Longitude)
		}
		fmt.Printf("📊 数据源: %s\n\n", location.Source)
	}
}
