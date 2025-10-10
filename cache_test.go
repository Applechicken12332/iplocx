package iplocx

import (
	"testing"
)

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(3)

	// 测试基本的 Put 和 Get
	t.Run("基本操作", func(t *testing.T) {
		loc1 := &Location{IP: "1.1.1.1", Country: "澳大利亚"}
		cache.Put("1.1.1.1", loc1)

		got, found := cache.Get("1.1.1.1")
		if !found {
			t.Error("应该找到缓存项")
		}
		if got.Country != "澳大利亚" {
			t.Errorf("Country = %s, want 澳大利亚", got.Country)
		}
	})

	// 测试未命中
	t.Run("缓存未命中", func(t *testing.T) {
		_, found := cache.Get("not-exists")
		if found {
			t.Error("不应该找到不存在的项")
		}
	})

	// 测试容量限制
	t.Run("容量限制", func(t *testing.T) {
		cache.Clear()

		cache.Put("1.1.1.1", &Location{IP: "1.1.1.1"})
		cache.Put("2.2.2.2", &Location{IP: "2.2.2.2"})
		cache.Put("3.3.3.3", &Location{IP: "3.3.3.3"})
		cache.Put("4.4.4.4", &Location{IP: "4.4.4.4"})

		stats := cache.Stats()
		if stats.Size != 3 {
			t.Errorf("缓存大小 = %d, want 3", stats.Size)
		}

		// 最旧的项（1.1.1.1）应该被移除
		_, found := cache.Get("1.1.1.1")
		if found {
			t.Error("最旧的项应该被移除")
		}

		// 其他项应该还在
		_, found = cache.Get("4.4.4.4")
		if !found {
			t.Error("最新的项应该存在")
		}
	})

	// 测试LRU策略
	t.Run("LRU策略", func(t *testing.T) {
		cache.Clear()

		cache.Put("1.1.1.1", &Location{IP: "1.1.1.1"})
		cache.Put("2.2.2.2", &Location{IP: "2.2.2.2"})
		cache.Put("3.3.3.3", &Location{IP: "3.3.3.3"})

		// 访问1.1.1.1，使其成为最新使用
		cache.Get("1.1.1.1")

		// 添加新项，应该移除2.2.2.2（最久未使用）
		cache.Put("4.4.4.4", &Location{IP: "4.4.4.4"})

		_, found := cache.Get("2.2.2.2")
		if found {
			t.Error("2.2.2.2 应该被移除")
		}

		_, found = cache.Get("1.1.1.1")
		if !found {
			t.Error("1.1.1.1 应该还在（刚访问过）")
		}
	})

	// 测试统计信息
	t.Run("统计信息", func(t *testing.T) {
		cache.Clear()

		cache.Put("1.1.1.1", &Location{IP: "1.1.1.1"})
		cache.Get("1.1.1.1") // 命中
		cache.Get("2.2.2.2") // 未命中

		stats := cache.Stats()
		if stats.Hits != 1 {
			t.Errorf("Hits = %d, want 1", stats.Hits)
		}
		if stats.Misses != 1 {
			t.Errorf("Misses = %d, want 1", stats.Misses)
		}
		if stats.HitRate != 50.0 {
			t.Errorf("HitRate = %.2f, want 50.00", stats.HitRate)
		}
	})

	// 测试更新现有项
	t.Run("更新现有项", func(t *testing.T) {
		cache.Clear()

		cache.Put("1.1.1.1", &Location{IP: "1.1.1.1", Country: "旧值"})
		cache.Put("1.1.1.1", &Location{IP: "1.1.1.1", Country: "新值"})

		got, _ := cache.Get("1.1.1.1")
		if got.Country != "新值" {
			t.Errorf("Country = %s, want 新值", got.Country)
		}

		stats := cache.Stats()
		if stats.Size != 1 {
			t.Errorf("更新后大小应该还是1，got %d", stats.Size)
		}
	})
}

// BenchmarkCachePut 性能测试：写入
func BenchmarkCachePut(b *testing.B) {
	cache := NewLRUCache(1000)
	loc := &Location{IP: "1.1.1.1", Country: "测试"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Put("1.1.1.1", loc)
	}
}

// BenchmarkCacheGet 性能测试：读取
func BenchmarkCacheGet(b *testing.B) {
	cache := NewLRUCache(1000)
	loc := &Location{IP: "1.1.1.1", Country: "测试"}
	cache.Put("1.1.1.1", loc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("1.1.1.1")
	}
}

// BenchmarkCacheMixed 性能测试：混合操作
func BenchmarkCacheMixed(b *testing.B) {
	cache := NewLRUCache(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ip := "1.1.1.1"
		if i%2 == 0 {
			cache.Put(ip, &Location{IP: ip})
		} else {
			cache.Get(ip)
		}
	}
}
