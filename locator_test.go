package iplocx

import (
	"testing"
)

// TestNewLocator 测试创建查询器
func TestNewLocator(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "双数据源",
			config: Config{
				QQwryDBPath:   "./data/qqwry.dat",
				GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
			},
			wantErr: false,
		},
		{
			name: "只有QQwry",
			config: Config{
				QQwryDBPath: "./data/qqwry.dat",
			},
			wantErr: false,
		},
		{
			name: "只有GeoLite",
			config: Config{
				GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
			},
			wantErr: false,
		},
		{
			name:    "无数据源",
			config:  Config{},
			wantErr: true,
		},
		{
			name: "启用缓存",
			config: Config{
				QQwryDBPath: "./data/qqwry.dat",
				EnableCache: true,
				CacheSize:   100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locator, err := NewLocator(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLocator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if locator != nil {
				defer locator.Close()

				// 验证缓存配置
				if tt.config.EnableCache && locator.cache == nil {
					t.Error("EnableCache is true but cache is nil")
				}
			}
		})
	}
}

// TestQuery 测试查询功能
func TestQuery(t *testing.T) {
	cfg := Config{
		QQwryDBPath:   "./data/qqwry.dat",
		GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
	}

	locator, err := NewLocator(cfg)
	if err != nil {
		t.Fatalf("NewLocator() error = %v", err)
	}
	defer locator.Close()

	tests := []struct {
		name    string
		ip      string
		wantErr bool
		check   func(*Location) bool
	}{
		{
			name:    "IPv4中国IP",
			ip:      "124.224.56.76",
			wantErr: false,
			check: func(loc *Location) bool {
				return loc.Country == "中国"
			},
		},
		{
			name:    "IPv4美国IP",
			ip:      "8.8.8.8",
			wantErr: false,
			check: func(loc *Location) bool {
				return loc.Country == "美国"
			},
		},
		{
			name:    "IPv6中国IP",
			ip:      "240e:446:2b02:5ab:a8b4:8948:125e:5140",
			wantErr: false,
			check: func(loc *Location) bool {
				return loc.Country == "中国"
			},
		},
		{
			name:    "无效IP",
			ip:      "invalid-ip",
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := locator.Query(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && location != nil {
				if !tt.check(location) {
					t.Errorf("Query() location check failed: %+v", location)
				}
			}
		})
	}
}

// TestCache 测试缓存功能
func TestCache(t *testing.T) {
	cfg := Config{
		QQwryDBPath:   "./data/qqwry.dat",
		GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
		EnableCache:   true,
		CacheSize:     10,
	}

	locator, err := NewLocator(cfg)
	if err != nil {
		t.Fatalf("NewLocator() error = %v", err)
	}
	defer locator.Close()

	ip := "8.8.8.8"

	// 第一次查询
	_, err = locator.Query(ip)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	// 检查缓存统计
	stats := locator.GetCacheStats()
	if stats == nil {
		t.Fatal("GetCacheStats() returned nil")
	}

	if stats.Size != 1 {
		t.Errorf("Cache size = %d, want 1", stats.Size)
	}

	// 第二次查询（应该从缓存）
	_, err = locator.Query(ip)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	stats = locator.GetCacheStats()
	if stats.Hits != 1 {
		t.Errorf("Cache hits = %d, want 1", stats.Hits)
	}

	if stats.HitRate < 40 || stats.HitRate > 60 {
		t.Errorf("Cache hit rate = %.2f%%, want ~50%%", stats.HitRate)
	}

	// 清空缓存
	locator.ClearCache()
	stats = locator.GetCacheStats()
	if stats.Size != 0 {
		t.Errorf("Cache size after clear = %d, want 0", stats.Size)
	}
}

// TestProviderStatus 测试数据源状态
func TestProviderStatus(t *testing.T) {
	cfg := Config{
		QQwryDBPath: "./data/qqwry.dat",
		// 不配置 GeoLite
	}

	locator, err := NewLocator(cfg)
	if err != nil {
		t.Fatalf("NewLocator() error = %v", err)
	}
	defer locator.Close()

	status := locator.GetProviderStatus()
	if !status["qqwry"] {
		t.Error("QQwry should be available")
	}
	if status["geolite"] {
		t.Error("GeoLite should not be available")
	}
}

// TestProviderInfo 测试详细数据源信息
func TestProviderInfo(t *testing.T) {
	cfg := Config{
		QQwryDBPath:   "./data/qqwry.dat",
		GeoLiteDBPath: "invalid-path.mmdb", // 故意使用无效路径
	}

	locator, err := NewLocator(cfg)
	if err != nil {
		t.Fatalf("NewLocator() error = %v", err)
	}
	defer locator.Close()

	info := locator.GetProviderInfo()

	if !info["qqwry"].Available {
		t.Error("QQwry should be available")
	}

	if info["geolite"].Available {
		t.Error("GeoLite should not be available")
	}

	if len(info["geolite"].Errors) == 0 {
		t.Error("GeoLite should have error messages")
	}
}

// TestLocationMethods 测试 Location 的方法
func TestLocationMethods(t *testing.T) {
	tests := []struct {
		name            string
		location        Location
		wantDetailScore int
		wantIsEmpty     bool
		wantHasDetail   bool
	}{
		{
			name: "完整信息",
			location: Location{
				Country:  "中国",
				Province: "北京",
				City:     "北京",
				District: "朝阳区",
			},
			wantDetailScore: 15,
			wantIsEmpty:     false,
			wantHasDetail:   true,
		},
		{
			name: "只有国家",
			location: Location{
				Country: "美国",
			},
			wantDetailScore: 1,
			wantIsEmpty:     false,
			wantHasDetail:   false,
		},
		{
			name:            "空信息",
			location:        Location{},
			wantDetailScore: 0,
			wantIsEmpty:     true,
			wantHasDetail:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.location.GetDetailScore(); got != tt.wantDetailScore {
				t.Errorf("GetDetailScore() = %v, want %v", got, tt.wantDetailScore)
			}
			if got := tt.location.IsEmpty(); got != tt.wantIsEmpty {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.wantIsEmpty)
			}
			if got := tt.location.HasDetailedInfo(); got != tt.wantHasDetail {
				t.Errorf("HasDetailedInfo() = %v, want %v", got, tt.wantHasDetail)
			}
		})
	}
}

// TestConcurrency 测试并发安全性
func TestConcurrency(t *testing.T) {
	cfg := Config{
		QQwryDBPath:   "./data/qqwry.dat",
		GeoLiteDBPath: "./data/GeoLite2-City.mmdb",
		EnableCache:   true,
	}

	locator, err := NewLocator(cfg)
	if err != nil {
		t.Fatalf("NewLocator() error = %v", err)
	}
	defer locator.Close()

	ips := []string{
		"8.8.8.8",
		"1.1.1.1",
		"124.224.56.76",
		"114.114.114.114",
	}

	// 并发查询
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				ip := ips[j%len(ips)]
				_, _ = locator.Query(ip)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证缓存统计
	stats := locator.GetCacheStats()
	if stats == nil {
		t.Fatal("GetCacheStats() returned nil")
	}

	t.Logf("Concurrent test - Cache hits: %d, misses: %d, hit rate: %.2f%%",
		stats.Hits, stats.Misses, stats.HitRate)
}
