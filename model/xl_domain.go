package model

// xunli管理平台 Domain 配置结构

type XLDomain struct {
	ID           string       `bson:"_id,omitempty"`
	PlatformInfo PlatformInfo `json:"platformInfo" bson:"platformInfo"`
	DomainConfig CDNDomain    `json:"domainConfig" bson:"domainConfig"`
}

type PlatformInfo struct {
	UserID      string
	ICPInfo     *ICPInfo `json:"icpInfo,omitempty" bson:"icpInfo,omitempty"` // ICP备案信息
	CreateAt    int64
	UpdateAt    int64
	ReverseName string `json:"reverseName" bson:"reverseName"` // 域名反查, 用来做域名后缀搜索索引
	Cname       string `json:"cname" bson:"cname"`             // centralhub 分配的cname记录,提供给用户接入CDN
	//CnameBackup string `json:"cnameBackup" bson:"cnameBackup"` // centralhub 分配的备用cname记录,提供给用户接入CDN
	//CnameAbroad  string `json:"cnameAbroad" bson:"cnameAbroad"` // centralhub 分配的海外cname记录,提供给用户接入CDN
	// line --> 线路
}

// ICPInfo ICP备案信息
type ICPInfo struct {
	Verified     bool   `json:"verified" bson:"verified"`         // 是否已验证
	ICPNumber    string `json:"icpNumber" bson:"icpNumber"`       // 备案号
	Status       string `json:"status" bson:"status"`             // 备案状态
	Owner        string `json:"owner" bson:"owner"`               // 备案主体
	VerifiedAt   int64  `json:"verifiedAt" bson:"verifiedAt"`     // 验证时间（Unix时间戳）
	CachedAt     int64  `json:"cachedAt" bson:"cachedAt"`         // 缓存时间（Unix时间戳）
}

// 綫路管理
// 綫路定義
