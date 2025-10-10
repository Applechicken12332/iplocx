package iplocx

import (
	"net/netip"
	"strings"

	"github.com/nuomiaa/iplocx/internal/qqwry"
)

// QQwryProvider 纯真IP库查询提供者
type QQwryProvider struct {
	db *qqwry.QQwry
	// 内置实现已经线程安全，不需要互斥锁
}

// NewQQwryProvider 创建纯真IP库查询提供者
func NewQQwryProvider(dbPath string) (*QQwryProvider, error) {
	db, err := qqwry.NewQQwry(dbPath)
	if err != nil {
		return nil, ErrDatabaseNotFound
	}
	return &QQwryProvider{db: db}, nil
}

// Query 查询IP地址
func (p *QQwryProvider) Query(ip string) (*Location, error) {
	if p.db == nil {
		return nil, ErrDatabaseNotFound
	}

	// QQwry只支持IPv4，检测IPv6地址
	addr, parseErr := netip.ParseAddr(ip)
	if parseErr != nil {
		return nil, ErrInvalidIP
	}
	if addr.Is6() {
		// IPv6地址，QQwry不支持
		return nil, ErrNoData
	}

	// 执行查询（内置实现已线程安全）
	result, err := p.db.Find(ip)
	if err != nil {
		return nil, ErrNoData
	}

	// 解析地理位置信息
	location := &Location{
		IP:     result.IP,
		ISP:    result.Area,
		Source: "qqwry",
	}

	// 使用长破折号分隔地理位置
	parts := strings.Split(result.Country, "–")

	// 根据分段数量填充字段
	if len(parts) > 0 {
		location.Country = strings.TrimSpace(parts[0])
	}
	if len(parts) > 1 {
		location.Province = strings.TrimSpace(parts[1])
	}
	if len(parts) > 2 {
		location.City = strings.TrimSpace(parts[2])
	}
	if len(parts) > 3 {
		location.District = strings.TrimSpace(parts[3])
	}

	return location, nil
}

// Close 关闭数据库连接
func (p *QQwryProvider) Close() error {
	// QQwry 不需要显式关闭
	return nil
}
