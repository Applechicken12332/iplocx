package iplocx

import "errors"

var (
	// ErrDatabaseNotFound 数据库文件未找到
	ErrDatabaseNotFound = errors.New("database file not found")

	// ErrInvalidIP 无效的IP地址
	ErrInvalidIP = errors.New("invalid IP address")

	// ErrNoData 未找到数据
	ErrNoData = errors.New("no data found for this IP")

	// ErrNoProvider 没有可用的查询提供者
	ErrNoProvider = errors.New("no provider available")
)
