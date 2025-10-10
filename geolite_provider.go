package iplocx

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

// GeoLiteProvider GeoLite2数据库查询提供者
type GeoLiteProvider struct {
	db *geoip2.Reader
}

// NewGeoLiteProvider 创建GeoLite2查询提供者
func NewGeoLiteProvider(dbPath string) (*GeoLiteProvider, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &GeoLiteProvider{db: db}, nil
}

// Query 查询IP地址
func (p *GeoLiteProvider) Query(ip string) (*Location, error) {
	if p.db == nil {
		return nil, ErrDatabaseNotFound
	}

	// 解析IP地址（v1 使用 net.IP）
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, ErrInvalidIP
	}

	// 查询城市信息
	record, err := p.db.City(parsedIP)
	if err != nil {
		return nil, err
	}

	// 构建位置信息
	location := &Location{
		IP:      ip,
		Country: record.Country.Names["zh-CN"],
		City:    record.City.Names["zh-CN"],
		Source:  "geolite2",
	}

	// 国家名称备选方案：Country中文 -> RegisteredCountry中文 -> Country英文 -> RegisteredCountry英文
	if location.Country == "" {
		location.Country = record.RegisteredCountry.Names["zh-CN"]
	}
	if location.Country == "" {
		location.Country = record.Country.Names["en"]
	}
	if location.Country == "" {
		location.Country = record.RegisteredCountry.Names["en"]
	}

	// 英文城市名作为备选
	if location.City == "" {
		location.City = record.City.Names["en"]
	}

	// 省/州信息
	if len(record.Subdivisions) > 0 {
		location.Province = record.Subdivisions[0].Names["zh-CN"]
		if location.Province == "" {
			location.Province = record.Subdivisions[0].Names["en"]
		}
	}

	// 经纬度信息
	location.Latitude = record.Location.Latitude
	location.Longitude = record.Location.Longitude

	// 时区信息
	location.TimeZone = record.Location.TimeZone

	return location, nil
}

// Close 关闭数据库连接
func (p *GeoLiteProvider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
