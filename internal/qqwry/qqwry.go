package qqwry

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	indexLen      = 7
	redirectMode1 = 0x01
	redirectMode2 = 0x02
)

var (
	ErrInvalidIP       = errors.New("invalid IP address")
	ErrDatabaseInvalid = errors.New("invalid database format")
)

// QQwry 纯真IP库查询器（内存版）
type QQwry struct {
	data       []byte // 整个数据库内容
	indexStart uint32 // 索引开始位置
	indexEnd   uint32 // 索引结束位置
}

// Result 查询结果
type Result struct {
	IP      string
	Country string
	Area    string
}

// NewQQwry 创建纯真IP查询器（一次性加载到内存）
func NewQQwry(dbPath string) (*QQwry, error) {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return nil, err
	}

	if len(data) < 8 {
		return nil, ErrDatabaseInvalid
	}

	q := &QQwry{
		data:       data,
		indexStart: binary.LittleEndian.Uint32(data[0:4]),
		indexEnd:   binary.LittleEndian.Uint32(data[4:8]),
	}

	return q, nil
}

// Find 查询IP地址（线程安全）
func (q *QQwry) Find(ip string) (*Result, error) {
	parsedIP := net.ParseIP(ip).To4()
	if parsedIP == nil {
		return nil, ErrInvalidIP
	}

	ipNum := binary.BigEndian.Uint32(parsedIP)
	offset := q.searchIndex(ipNum)
	if offset == 0 {
		return nil, errors.New("IP not found")
	}

	country, area := q.readLocation(offset + 4)

	return &Result{
		IP:      ip,
		Country: q.gbkToUtf8(country),
		Area:    q.gbkToUtf8(area),
	}, nil
}

// searchIndex 二分查找IP索引
func (q *QQwry) searchIndex(ip uint32) uint32 {
	start := q.indexStart
	end := q.indexEnd

	for {
		mid := q.getMiddleOffset(start, end)

		if end-start == indexLen {
			// 最后一个索引
			offset := q.readUInt24(mid + 4)
			buf := q.data[end : end+4]
			if ip < binary.LittleEndian.Uint32(buf) {
				return offset
			}
			return 0
		}

		buf := q.data[mid : mid+4]
		midIP := binary.LittleEndian.Uint32(buf)

		if midIP > ip {
			end = mid
		} else if midIP < ip {
			start = mid
		} else {
			return q.readUInt24(mid + 4)
		}
	}
}

// readLocation 读取位置信息
func (q *QQwry) readLocation(offset uint32) (country, area []byte) {
	mode := q.readMode(offset)

	if mode == redirectMode1 {
		countryOffset := q.readUInt24(offset + 1)
		mode = q.readMode(countryOffset)
		if mode == redirectMode2 {
			c := q.readUInt24(countryOffset + 1)
			country = q.readString(c)
			countryOffset += 4
		} else {
			country = q.readString(countryOffset)
			countryOffset += uint32(len(country) + 1)
		}
		area = q.readArea(countryOffset)
	} else if mode == redirectMode2 {
		countryOffset := q.readUInt24(offset + 1)
		country = q.readString(countryOffset)
		area = q.readArea(offset + 4)
	} else {
		country = q.readString(offset)
		area = q.readArea(offset + uint32(len(country)+1))
	}

	return
}

// readArea 读取区域信息
func (q *QQwry) readArea(offset uint32) []byte {
	mode := q.readMode(offset)
	if mode == redirectMode1 || mode == redirectMode2 {
		areaOffset := q.readUInt24(offset + 1)
		if areaOffset == 0 {
			return []byte("")
		}
		return q.readString(areaOffset)
	}
	return q.readString(offset)
}

// readString 读取以0结尾的字符串
func (q *QQwry) readString(offset uint32) []byte {
	if offset >= uint32(len(q.data)) {
		return []byte("")
	}

	end := offset
	for end < uint32(len(q.data)) && q.data[end] != 0 {
		end++
	}
	return q.data[offset:end]
}

// readMode 读取模式字节
func (q *QQwry) readMode(offset uint32) byte {
	if offset >= uint32(len(q.data)) {
		return 0
	}
	return q.data[offset]
}

// readUInt24 读取3字节小端序整数
func (q *QQwry) readUInt24(offset uint32) uint32 {
	if offset+3 > uint32(len(q.data)) {
		return 0
	}
	buf := q.data[offset : offset+3]
	return uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16
}

// getMiddleOffset 获取中间偏移量
func (q *QQwry) getMiddleOffset(start, end uint32) uint32 {
	records := ((end - start) / indexLen) >> 1
	return start + records*indexLen
}

// gbkToUtf8 GBK转UTF-8
func (q *QQwry) gbkToUtf8(src []byte) string {
	if len(src) == 0 {
		return ""
	}

	reader := transform.NewReader(bytes.NewReader(src), simplifiedchinese.GBK.NewDecoder())
	result, err := io.ReadAll(reader)
	if err != nil {
		return string(src) // 转换失败返回原始字符串
	}
	return string(result)
}
