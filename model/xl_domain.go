package model

// xunli管理平台 Domain 配置结构

type XLDomain struct {
	ID           string       `bson:"_id,omitempty"`
	PlatformInfo PlatformInfo `json:"platformInfo" bson:"platformInfo"`
	DomainConfig CDNDomain    `json:"domainConfig" bson:"domainConfig"`
}

type PlatformInfo struct {
	UserID      string
	CreateAt    int64
	UpdateAt    int64
	ReverseName string `json:"reverseName" bson:"reverseName"` // 域名反查, 用来做域名后缀搜索索引
	Cname       string `json:"cname" bson:"cname"`             // centralhub 分配的cname记录,提供给用户接入CDN
	//CnameBackup string `json:"cnameBackup" bson:"cnameBackup"` // centralhub 分配的备用cname记录,提供给用户接入CDN
	//CnameAbroad  string `json:"cnameAbroad" bson:"cnameAbroad"` // centralhub 分配的海外cname记录,提供给用户接入CDN
	// line --> 线路
}

// 綫路管理
// 綫路定義
