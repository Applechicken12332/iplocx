package iplocx

// Location 统一的IP地理位置信息结构
type Location struct {
	IP        string  `json:"ip"`        // IP地址
	Country   string  `json:"country"`   // 国家
	Province  string  `json:"province"`  // 省/州
	City      string  `json:"city"`      // 市
	District  string  `json:"district"`  // 区/县
	ISP       string  `json:"isp"`       // 运营商
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
	TimeZone  string  `json:"timezone"`  // 时区
	Source    string  `json:"source"`    // 数据来源 (qqwry/geolite2/combined)
}

// HasDetailedInfo 判断是否有详细的省市信息
func (l *Location) HasDetailedInfo() bool {
	// 如果有省份或城市信息，认为是详细信息
	return l.Province != "" || l.City != ""
}

// IsEmpty 判断位置信息是否为空
func (l *Location) IsEmpty() bool {
	return l.Country == "" && l.Province == "" && l.City == ""
}

// GetDetailScore 获取地理信息的详细程度分数
// 分数越高，地理信息越详细
// 评分规则：国家(1分) + 省/州(2分) + 市(4分) + 区/县(8分)
func (l *Location) GetDetailScore() int {
	score := 0
	if l.Country != "" {
		score += 1
	}
	if l.Province != "" {
		score += 2
	}
	if l.City != "" {
		score += 4
	}
	if l.District != "" {
		score += 8
	}
	return score
}
