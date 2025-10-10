package iplocx

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Provider IP查询提供者接口
type Provider interface {
	Query(ip string) (*Location, error)
	Close() error
}

// Locator IP地理位置查询器
type Locator struct {
	qqwryProvider   Provider
	geoliteProvider Provider
	config          Config
	cache           *LRUCache
	initErrors      []string // 初始化时的错误信息
	stats           *QueryStats
	mu              sync.RWMutex
}

// QueryStats 查询统计信息
type QueryStats struct {
	TotalQueries   int64 // 总查询次数
	SuccessQueries int64 // 成功查询次数
	FailedQueries  int64 // 失败查询次数
	QQwryHits      int64 // QQwry数据源使用次数
	GeoLiteHits    int64 // GeoLite数据源使用次数
	CombinedHits   int64 // 合并数据使用次数
	TotalDuration  int64 // 总查询时间（纳秒）
}

// Config 配置选项
type Config struct {
	QQwryDBPath   string // 纯真IP库数据库路径
	GeoLiteDBPath string // GeoLite2数据库路径
	Debug         bool   // 是否启用调试输出
	EnableCache   bool   // 是否启用缓存
	CacheSize     int    // 缓存大小（默认1000）
}

// NewLocator 创建IP查询器
func NewLocator(cfg Config) (*Locator, error) {
	locator := &Locator{
		config: cfg,
		stats:  &QueryStats{},
	}

	// 初始化缓存（如果启用）
	if cfg.EnableCache {
		cacheSize := cfg.CacheSize
		if cacheSize <= 0 {
			cacheSize = 1000 // 默认缓存1000条
		}
		locator.cache = NewLRUCache(cacheSize)
	}

	// 初始化QQwry提供者（可选）
	if cfg.QQwryDBPath != "" {
		qqwryProvider, err := NewQQwryProvider(cfg.QQwryDBPath)
		if err == nil {
			locator.qqwryProvider = qqwryProvider
		} else {
			locator.initErrors = append(locator.initErrors,
				fmt.Sprintf("QQwry初始化失败: %v", err))
		}
	}

	// 初始化GeoLite提供者（可选）
	if cfg.GeoLiteDBPath != "" {
		geoliteProvider, err := NewGeoLiteProvider(cfg.GeoLiteDBPath)
		if err == nil {
			locator.geoliteProvider = geoliteProvider
		} else {
			locator.initErrors = append(locator.initErrors,
				fmt.Sprintf("GeoLite初始化失败: %v", err))
		}
	}

	// 至少需要一个提供者
	if locator.qqwryProvider == nil && locator.geoliteProvider == nil {
		if len(locator.initErrors) > 0 {
			return nil, fmt.Errorf("%w: %v", ErrNoProvider, locator.initErrors)
		}
		return nil, ErrNoProvider
	}

	return locator, nil
}

// Query 查询IP地址
// 策略：
// 1. 先查询缓存
// 2. 并行查询QQwry和GeoLite2两个数据源
// 3. 根据数据完整度评分，选择更详细的数据源作为基础
// 4. 使用另一个数据源补充缺失的字段
// 5. 将结果放入缓存
func (l *Locator) Query(ip string) (*Location, error) {
	startTime := time.Now()
	atomic.AddInt64(&l.stats.TotalQueries, 1)

	// 先查询缓存
	if l.cache != nil {
		if loc, found := l.cache.Get(ip); found {
			atomic.AddInt64(&l.stats.SuccessQueries, 1) // 缓存命中也算成功查询
			atomic.AddInt64(&l.stats.TotalDuration, int64(time.Since(startTime)))
			return loc, nil
		}
	}

	// 获取provider引用（使用短暂的读锁）
	l.mu.RLock()
	qqwryProvider := l.qqwryProvider
	geoliteProvider := l.geoliteProvider
	l.mu.RUnlock()

	// 使用channel接收并行查询结果
	type result struct {
		location *Location
		err      error
		source   string
	}

	results := make(chan result, 2)
	queryCount := 0

	// 并行查询QQwry
	if qqwryProvider != nil {
		queryCount++
		go func() {
			loc, err := qqwryProvider.Query(ip)
			results <- result{location: loc, err: err, source: "qqwry"}
		}()
	}

	// 并行查询GeoLite2
	if geoliteProvider != nil {
		queryCount++
		go func() {
			loc, err := geoliteProvider.Query(ip)
			results <- result{location: loc, err: err, source: "geolite2"}
		}()
	}

	// 如果没有可用的提供者
	if queryCount == 0 {
		return nil, ErrNoProvider
	}

	// 收集查询结果
	var qqwryLoc, geoliteLoc *Location
	var qqwryErr, geoliteErr error

	for i := 0; i < queryCount; i++ {
		res := <-results
		if res.source == "qqwry" {
			qqwryLoc = res.location
			qqwryErr = res.err
		} else {
			geoliteLoc = res.location
			geoliteErr = res.err
		}
	}

	// 如果两者都失败，返回错误
	if (qqwryLoc == nil || qqwryLoc.IsEmpty()) && (geoliteLoc == nil || geoliteLoc.IsEmpty()) {
		atomic.AddInt64(&l.stats.FailedQueries, 1)
		atomic.AddInt64(&l.stats.TotalDuration, int64(time.Since(startTime)))
		if qqwryErr != nil {
			return nil, qqwryErr
		}
		if geoliteErr != nil {
			return nil, geoliteErr
		}
		return nil, ErrNoData
	}

	// 只有一个数据源有结果
	if qqwryLoc == nil || qqwryLoc.IsEmpty() {
		atomic.AddInt64(&l.stats.GeoLiteHits, 1)
		atomic.AddInt64(&l.stats.SuccessQueries, 1)
		atomic.AddInt64(&l.stats.TotalDuration, int64(time.Since(startTime)))
		if l.cache != nil {
			l.cache.Put(ip, geoliteLoc)
		}
		return geoliteLoc, nil
	}
	if geoliteLoc == nil || geoliteLoc.IsEmpty() {
		atomic.AddInt64(&l.stats.QQwryHits, 1)
		atomic.AddInt64(&l.stats.SuccessQueries, 1)
		atomic.AddInt64(&l.stats.TotalDuration, int64(time.Since(startTime)))
		if l.cache != nil {
			l.cache.Put(ip, qqwryLoc)
		}
		return qqwryLoc, nil
	}

	// 两者都有结果，智能合并
	atomic.AddInt64(&l.stats.CombinedHits, 1)
	atomic.AddInt64(&l.stats.SuccessQueries, 1)
	merged := l.mergeLocations(qqwryLoc, geoliteLoc)
	if l.cache != nil && merged != nil {
		l.cache.Put(ip, merged)
	}
	atomic.AddInt64(&l.stats.TotalDuration, int64(time.Since(startTime)))
	return merged, nil
}

// mergeLocations 智能合并两个位置信息
// 策略：
// 1. 比较两个数据源的详细程度分数
// 2. 以分数高的为基础，用另一个补充缺失字段
// 3. 保留各自独有的信息（如ISP、经纬度）
func (l *Locator) mergeLocations(qqwry, geolite *Location) *Location {
	// 计算详细程度分数
	qqwryScore := qqwry.GetDetailScore()
	geoliteScore := geolite.GetDetailScore()

	// 调试输出：打印数据来源
	l.debugLog("\n=== 数据合并调试 ===\n")
	l.debugLog("QQwry  评分: %d | Country:%s | Province:%s | City:%s | District:%s | ISP:%s\n",
		qqwryScore, qqwry.Country, qqwry.Province, qqwry.City, qqwry.District, qqwry.ISP)
	l.debugLog("GeoLite 评分: %d | Country:%s | Province:%s | City:%s | District:%s | 经纬度:(%.4f,%.4f) | 时区:%s\n",
		geoliteScore, geolite.Country, geolite.Province, geolite.City, geolite.District,
		geolite.Latitude, geolite.Longitude, geolite.TimeZone)
	l.debugLog("==================\n\n")

	// 选择分数高的作为基础数据源
	var primary, secondary *Location
	var primarySource, secondarySource string
	if qqwryScore >= geoliteScore {
		primary = qqwry
		secondary = geolite
		primarySource = "qqwry"
		secondarySource = "geolite"
	} else {
		primary = geolite
		secondary = qqwry
		primarySource = "geolite"
		secondarySource = "qqwry"
	}

	l.debugLog("选择 %s 作为主数据源（分数: %d vs %d）\n", primarySource,
		primary.GetDetailScore(), secondary.GetDetailScore())

	// 以主数据源为基础创建合并结果
	merged := &Location{
		IP:       primary.IP,
		Country:  primary.Country,
		Province: primary.Province,
		City:     primary.City,
		District: primary.District,
		Source:   "combined",
	}

	// 用次要数据源补充缺失的地理信息字段
	if merged.Country == "" && secondary.Country != "" {
		l.debugLog("  → Country 补充自 %s: %s\n", secondarySource, secondary.Country)
		merged.Country = secondary.Country
	}
	if merged.Province == "" && secondary.Province != "" {
		l.debugLog("  → Province 补充自 %s: %s\n", secondarySource, secondary.Province)
		merged.Province = secondary.Province
	}
	if merged.City == "" && secondary.City != "" {
		l.debugLog("  → City 补充自 %s: %s\n", secondarySource, secondary.City)
		merged.City = secondary.City
	}
	if merged.District == "" && secondary.District != "" {
		l.debugLog("  → District 补充自 %s: %s\n", secondarySource, secondary.District)
		merged.District = secondary.District
	}

	// ISP信息：优先使用qqwry（只有qqwry有ISP信息）
	if qqwry.ISP != "" {
		l.debugLog("  → ISP 来自 qqwry: %s\n", qqwry.ISP)
		merged.ISP = qqwry.ISP
	} else if geolite.ISP != "" {
		l.debugLog("  → ISP 来自 geolite: %s\n", geolite.ISP)
		merged.ISP = geolite.ISP
	}

	// 经纬度和时区：优先使用geolite（只有geolite有这些信息）
	if geolite.Latitude != 0 || geolite.Longitude != 0 {
		l.debugLog("  → 经纬度 来自 geolite: (%.4f, %.4f)\n", geolite.Latitude, geolite.Longitude)
		merged.Latitude = geolite.Latitude
		merged.Longitude = geolite.Longitude
	} else {
		merged.Latitude = qqwry.Latitude
		merged.Longitude = qqwry.Longitude
	}

	if geolite.TimeZone != "" {
		l.debugLog("  → 时区 来自 geolite: %s\n", geolite.TimeZone)
		merged.TimeZone = geolite.TimeZone
	} else if qqwry.TimeZone != "" {
		l.debugLog("  → 时区 来自 qqwry: %s\n", qqwry.TimeZone)
		merged.TimeZone = qqwry.TimeZone
	}

	l.debugLog("\n")
	return merged
}

// debugLog 调试日志输出（仅在Debug模式下）
func (l *Locator) debugLog(format string, args ...interface{}) {
	if l.config.Debug {
		fmt.Printf(format, args...)
	}
}

// ProviderInfo 数据源信息
type ProviderInfo struct {
	Available bool     // 是否可用
	Errors    []string // 初始化错误（如果有）
}

// GetProviderStatus 获取数据源加载状态
func (l *Locator) GetProviderStatus() map[string]bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return map[string]bool{
		"qqwry":   l.qqwryProvider != nil,
		"geolite": l.geoliteProvider != nil,
	}
}

// GetProviderInfo 获取详细的数据源信息
func (l *Locator) GetProviderInfo() map[string]ProviderInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()

	info := map[string]ProviderInfo{
		"qqwry": {
			Available: l.qqwryProvider != nil,
		},
		"geolite": {
			Available: l.geoliteProvider != nil,
		},
	}

	// 添加错误信息
	for _, errMsg := range l.initErrors {
		if len(errMsg) > 6 && errMsg[:6] == "QQwry" {
			qi := info["qqwry"]
			qi.Errors = append(qi.Errors, errMsg)
			info["qqwry"] = qi
		} else if len(errMsg) > 7 && errMsg[:7] == "GeoLite" {
			gi := info["geolite"]
			gi.Errors = append(gi.Errors, errMsg)
			info["geolite"] = gi
		}
	}

	return info
}

// GetCacheStats 获取缓存统计信息
func (l *Locator) GetCacheStats() *CacheStats {
	if l.cache == nil {
		return nil
	}
	stats := l.cache.Stats()
	return &stats
}

// ClearCache 清空缓存
func (l *Locator) ClearCache() {
	if l.cache != nil {
		l.cache.Clear()
	}
}

// GetQueryStats 获取查询统计信息
func (l *Locator) GetQueryStats() QueryStatsSnapshot {
	return QueryStatsSnapshot{
		TotalQueries:   atomic.LoadInt64(&l.stats.TotalQueries),
		SuccessQueries: atomic.LoadInt64(&l.stats.SuccessQueries),
		FailedQueries:  atomic.LoadInt64(&l.stats.FailedQueries),
		QQwryHits:      atomic.LoadInt64(&l.stats.QQwryHits),
		GeoLiteHits:    atomic.LoadInt64(&l.stats.GeoLiteHits),
		CombinedHits:   atomic.LoadInt64(&l.stats.CombinedHits),
		AvgDuration:    l.getAvgDuration(),
		SuccessRate:    l.getSuccessRate(),
	}
}

// QueryStatsSnapshot 查询统计快照（用于读取）
type QueryStatsSnapshot struct {
	TotalQueries   int64         // 总查询次数
	SuccessQueries int64         // 成功查询次数
	FailedQueries  int64         // 失败查询次数
	QQwryHits      int64         // QQwry数据源使用次数
	GeoLiteHits    int64         // GeoLite数据源使用次数
	CombinedHits   int64         // 合并数据使用次数
	AvgDuration    time.Duration // 平均查询时间
	SuccessRate    float64       // 成功率（百分比）
}

// getAvgDuration 计算平均查询时间
func (l *Locator) getAvgDuration() time.Duration {
	total := atomic.LoadInt64(&l.stats.TotalQueries)
	if total == 0 {
		return 0
	}
	duration := atomic.LoadInt64(&l.stats.TotalDuration)
	return time.Duration(duration / total)
}

// getSuccessRate 计算成功率
func (l *Locator) getSuccessRate() float64 {
	total := atomic.LoadInt64(&l.stats.TotalQueries)
	if total == 0 {
		return 0
	}
	success := atomic.LoadInt64(&l.stats.SuccessQueries)
	return float64(success) / float64(total) * 100
}

// ResetStats 重置统计信息
func (l *Locator) ResetStats() {
	atomic.StoreInt64(&l.stats.TotalQueries, 0)
	atomic.StoreInt64(&l.stats.SuccessQueries, 0)
	atomic.StoreInt64(&l.stats.FailedQueries, 0)
	atomic.StoreInt64(&l.stats.QQwryHits, 0)
	atomic.StoreInt64(&l.stats.GeoLiteHits, 0)
	atomic.StoreInt64(&l.stats.CombinedHits, 0)
	atomic.StoreInt64(&l.stats.TotalDuration, 0)
}

// Close 关闭所有数据库连接
func (l *Locator) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var errs []error

	if l.qqwryProvider != nil {
		if err := l.qqwryProvider.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close qqwry: %w", err))
		}
	}

	if l.geoliteProvider != nil {
		if err := l.geoliteProvider.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close geolite: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}
