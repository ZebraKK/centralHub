package model

// reference:

// xunli Domain 配置

type XLDomain struct {
	ID     string `bson:"_id,omitempty"`
	Name   string `bson:"name"`
	Owner  string `bson:"owner"`
	Status string
}

type XLPlatformInfo struct {
	ReverseName string
	UserID      string
}

type StatusRecord struct {
	CreateAt int64
	UpdateAt int64
}

// 属性： 类型，泛域名

// cname 记录

// cdn 提供商

// 线路 -->

// 备案

// cache 缓存

// source 源

// 访问：黑白名单 防盗链 鉴权

// https http2 证书

// 302 配置
