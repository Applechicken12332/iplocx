package iplocx

import (
	"testing"
)

// 基准测试：并行查询性能
func BenchmarkQuery(b *testing.B) {
	cfg := Config{
		QQwryDBPath:   "./data/qqwry.dat",
		GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
	}

	locator, err := NewLocator(cfg)
	if err != nil {
		b.Fatalf("初始化失败: %v", err)
	}
	defer locator.Close()

	testIPs := []string{
		"124.224.56.76",   // 中国IP
		"8.8.8.8",         // 美国IP
		"114.114.114.114", // 中国DNS
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ip := testIPs[i%len(testIPs)]
		_, _ = locator.Query(ip)
	}
}

// 基准测试：单独测试QQwry性能
func BenchmarkQQwryOnly(b *testing.B) {
	provider, err := NewQQwryProvider("./data/qqwry.dat")
	if err != nil {
		b.Fatalf("初始化失败: %v", err)
	}
	defer provider.Close()

	testIPs := []string{
		"124.224.56.76",
		"8.8.8.8",
		"114.114.114.114",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ip := testIPs[i%len(testIPs)]
		_, _ = provider.Query(ip)
	}
}

// 基准测试：单独测试GeoLite性能
func BenchmarkGeoLiteOnly(b *testing.B) {
	provider, err := NewGeoLiteProvider("./data/GeoLite2-City.mmdb")
	if err != nil {
		b.Fatalf("初始化失败: %v", err)
	}
	defer provider.Close()

	testIPs := []string{
		"124.224.56.76",
		"8.8.8.8",
		"114.114.114.114",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ip := testIPs[i%len(testIPs)]
		_, _ = provider.Query(ip)
	}
}

// 并发测试：测试并发查询的性能
func BenchmarkQueryParallel(b *testing.B) {
	cfg := Config{
		QQwryDBPath:   "./data/qqwry.dat",
		GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
	}

	locator, err := NewLocator(cfg)
	if err != nil {
		b.Fatalf("初始化失败: %v", err)
	}
	defer locator.Close()

	testIPs := []string{
		"124.224.56.76",
		"8.8.8.8",
		"114.114.114.114",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ip := testIPs[i%len(testIPs)]
			_, _ = locator.Query(ip)
			i++
		}
	})
}
