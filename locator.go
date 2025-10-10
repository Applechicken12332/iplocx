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
	if loc := l.queryCacheIfEnabled(ip, startTime); loc != nil {
		return loc, nil
	}

	// 并行查询数据源
	qqwryLoc, geoliteLoc, qqwryErr, geoliteErr := l.queryProviders(ip)

	// 处理查询结果
	result, err := l.processQueryResults(ip, qqwryLoc, geoliteLoc, qqwryErr, geoliteErr)

	// 更新统计
	l.updateQueryStats(startTime)
	return result, err
}

// queryCacheIfEnabled 查询缓存（如果启用）
func (l *Locator) queryCacheIfEnabled(ip string, startTime time.Time) *Location {
	if l.cache == nil {
		return nil
	}

	if loc, found := l.cache.Get(ip); found {
		atomic.AddInt64(&l.stats.SuccessQueries, 1)
		atomic.AddInt64(&l.stats.TotalDuration, int64(time.Since(startTime)))
		return loc
	}
	return nil
}

// queryProviders 并行查询所有数据源
func (l *Locator) queryProviders(ip string) (*Location, *Location, error, error) {
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

	// 收集查询结果
	var qqwryLoc, geoliteLoc *Location
	var qqwryErr, geoliteErr error

	for i := 0; i < queryCount; i++ {
		res := <-results
		if res.source == "qqwry" {
			qqwryLoc, qqwryErr = res.location, res.err
		} else {
			geoliteLoc, geoliteErr = res.location, res.err
		}
	}

	return qqwryLoc, geoliteLoc, qqwryErr, geoliteErr
}

// processQueryResults 处理查询结果并返回最终位置
func (l *Locator) processQueryResults(ip string, qqwryLoc, geoliteLoc *Location, qqwryErr, geoliteErr error) (*Location, error) {
	// 如果两者都失败，返回错误
	if l.isBothEmpty(qqwryLoc, geoliteLoc) {
		return l.handleQueryFailure(qqwryErr, geoliteErr)
	}

	// 只有一个数据源有结果
	if singleResult := l.handleSingleSource(ip, qqwryLoc, geoliteLoc); singleResult != nil {
		return singleResult, nil
	}

	// 两者都有结果，智能合并
	return l.handleMergedResult(ip, qqwryLoc, geoliteLoc), nil
}

// isBothEmpty 检查两个位置是否都为空
func (l *Locator) isBothEmpty(qqwryLoc, geoliteLoc *Location) bool {
	return (qqwryLoc == nil || qqwryLoc.IsEmpty()) && (geoliteLoc == nil || geoliteLoc.IsEmpty())
}

// handleQueryFailure 处理查询失败的情况
func (l *Locator) handleQueryFailure(qqwryErr, geoliteErr error) (*Location, error) {
	atomic.AddInt64(&l.stats.FailedQueries, 1)
	if qqwryErr != nil {
		return nil, qqwryErr
	}
	if geoliteErr != nil {
		return nil, geoliteErr
	}
	return nil, ErrNoData
}

// handleSingleSource 处理只有单个数据源有结果的情况
func (l *Locator) handleSingleSource(ip string, qqwryLoc, geoliteLoc *Location) *Location {
	if qqwryLoc == nil || qqwryLoc.IsEmpty() {
		atomic.AddInt64(&l.stats.GeoLiteHits, 1)
		atomic.AddInt64(&l.stats.SuccessQueries, 1)
		if l.cache != nil {
			l.cache.Put(ip, geoliteLoc)
		}
		return geoliteLoc
	}

	if geoliteLoc == nil || geoliteLoc.IsEmpty() {
		atomic.AddInt64(&l.stats.QQwryHits, 1)
		atomic.AddInt64(&l.stats.SuccessQueries, 1)
		if l.cache != nil {
			l.cache.Put(ip, qqwryLoc)
		}
		return qqwryLoc
	}

	return nil
}

// handleMergedResult 处理合并结果
func (l *Locator) handleMergedResult(ip string, qqwryLoc, geoliteLoc *Location) *Location {
	atomic.AddInt64(&l.stats.CombinedHits, 1)
	atomic.AddInt64(&l.stats.SuccessQueries, 1)
	merged := l.mergeLocations(qqwryLoc, geoliteLoc)
	if l.cache != nil && merged != nil {
		l.cache.Put(ip, merged)
	}
	return merged
}

// updateQueryStats 更新查询统计信息
func (l *Locator) updateQueryStats(startTime time.Time) {
	atomic.AddInt64(&l.stats.TotalDuration, int64(time.Since(startTime)))
}

// mergeLocations 智能合并两个位置信息
// 策略：
// 1. 比较两个数据源的详细程度分数
// 2. 以分数高的为基础，用另一个补充缺失字段
// 3. 保留各自独有的信息（如ISP、经纬度）
func (l *Locator) mergeLocations(qqwry, geolite *Location) *Location {
	// 打印合并调试信息
	l.logMergeDebugInfo(qqwry, geolite)

	// 选择主次数据源
	primary, secondary, _, secondarySource := l.selectPrimarySources(qqwry, geolite)

	// 创建基础合并结果
	merged := l.createBaseMergedLocation(primary)

	// 补充缺失字段
	l.fillMissingFields(merged, secondary, secondarySource)

	// 合并特殊字段（ISP、经纬度、时区）
	l.mergeSpecialFields(merged, qqwry, geolite)

	l.debugLog("\n")
	return merged
}

// logMergeDebugInfo 打印合并调试信息
func (l *Locator) logMergeDebugInfo(qqwry, geolite *Location) {
	qqwryScore := qqwry.GetDetailScore()
	geoliteScore := geolite.GetDetailScore()

	l.debugLog("\n=== 数据合并调试 ===\n")
	l.debugLog("QQwry  评分: %d | Country:%s | Province:%s | City:%s | District:%s | ISP:%s\n",
		qqwryScore, qqwry.Country, qqwry.Province, qqwry.City, qqwry.District, qqwry.ISP)
	l.debugLog("GeoLite 评分: %d | Country:%s | Province:%s | City:%s | District:%s | 经纬度:(%.4f,%.4f) | 时区:%s\n",
		geoliteScore, geolite.Country, geolite.Province, geolite.City, geolite.District,
		geolite.Latitude, geolite.Longitude, geolite.TimeZone)
	l.debugLog("==================\n\n")
}

// selectPrimarySources 选择主次数据源
func (l *Locator) selectPrimarySources(qqwry, geolite *Location) (*Location, *Location, string, string) {
	qqwryScore := qqwry.GetDetailScore()
	geoliteScore := geolite.GetDetailScore()

	var primary, secondary *Location
	var primarySource, secondarySource string

	if qqwryScore >= geoliteScore {
		primary, secondary = qqwry, geolite
		primarySource, secondarySource = "qqwry", "geolite"
	} else {
		primary, secondary = geolite, qqwry
		primarySource, secondarySource = "geolite", "qqwry"
	}

	l.debugLog("选择 %s 作为主数据源（分数: %d vs %d）\n", primarySource,
		primary.GetDetailScore(), secondary.GetDetailScore())

	return primary, secondary, primarySource, secondarySource
}

// createBaseMergedLocation 创建基础合并结果
func (l *Locator) createBaseMergedLocation(primary *Location) *Location {
	return &Location{
		IP:       primary.IP,
		Country:  primary.Country,
		Province: primary.Province,
		City:     primary.City,
		District: primary.District,
		Source:   "combined",
	}
}

// fillMissingFields 用次要数据源补充缺失的地理信息字段
func (l *Locator) fillMissingFields(merged, secondary *Location, secondarySource string) {
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
}

// mergeSpecialFields 合并特殊字段（ISP、经纬度、时区）
func (l *Locator) mergeSpecialFields(merged, qqwry, geolite *Location) {
	// ISP信息：优先使用qqwry（只有qqwry有ISP信息）
	l.mergeISP(merged, qqwry, geolite)

	// 经纬度：优先使用geolite（只有geolite有这些信息）
	l.mergeCoordinates(merged, qqwry, geolite)

	// 时区：优先使用geolite
	l.mergeTimeZone(merged, qqwry, geolite)
}

// mergeISP 合并ISP信息
func (l *Locator) mergeISP(merged, qqwry, geolite *Location) {
	if qqwry.ISP != "" {
		l.debugLog("  → ISP 来自 qqwry: %s\n", qqwry.ISP)
		merged.ISP = qqwry.ISP
	} else if geolite.ISP != "" {
		l.debugLog("  → ISP 来自 geolite: %s\n", geolite.ISP)
		merged.ISP = geolite.ISP
	}
}

// mergeCoordinates 合并经纬度信息
func (l *Locator) mergeCoordinates(merged, qqwry, geolite *Location) {
	if geolite.Latitude != 0 || geolite.Longitude != 0 {
		l.debugLog("  → 经纬度 来自 geolite: (%.4f, %.4f)\n", geolite.Latitude, geolite.Longitude)
		merged.Latitude = geolite.Latitude
		merged.Longitude = geolite.Longitude
	} else {
		merged.Latitude = qqwry.Latitude
		merged.Longitude = qqwry.Longitude
	}
}

// mergeTimeZone 合并时区信息
func (l *Locator) mergeTimeZone(merged, qqwry, geolite *Location) {
	if geolite.TimeZone != "" {
		l.debugLog("  → 时区 来自 geolite: %s\n", geolite.TimeZone)
		merged.TimeZone = geolite.TimeZone
	} else if qqwry.TimeZone != "" {
		l.debugLog("  → 时区 来自 qqwry: %s\n", qqwry.TimeZone)
		merged.TimeZone = qqwry.TimeZone
	}
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
