package model

// reference:

type DomainStatusConfig struct {
	Disabled        bool   `json:"disabled" bson:"disabled"` // 是否禁用
	Reason          string `json:"reason" bson:"reason"`     // 禁用原因
	DisabledExpires int64  `json:"disabledExpires" bson:"disabledExpires"`
	DisabledMsg     string `json:"disabledMsg" bson:"disabledMsg"`
}

type CacheKeyRule struct {
	CacheKeyHost   string `json:"cacheKeyHost" bson:"cacheKeyHost"`     // 是否将 Host 头作为缓存键的一部分
	CacheKeyQuery  string `json:"cacheKeyQuery" bson:"cacheKeyQuery"`   // 是否将查询参数作为缓存键的一部分
	CacheKeyHead   string `json:"cacheKeyHead" bson:"cacheKeyHead"`     // 是否将路径作为缓存键的一部分
	CacheKeyScheme string `json:"cacheKeyScheme" bson:"cacheKeyScheme"` // 是否将协议作为缓存键的一部分 http/https
}

// 缓存配置
type DomainCacheConfig struct {
	GlobalCacheTime int          `json:"globalCacheTime" bson:"globalCacheTime"` // 默认缓存时间。-1 表示不缓存，下同。
	CacheKeyRule    CacheKeyRule `json:"cacheKeyRule" bson:"cacheKeyRule"`
	// to add
	// 缓存规则,目录匹配，前置匹配，后缀匹配 ---> 改写放到CacheKeyRule里
	// 响应码缓存
	// 去问号缓存
	// 忽略参数缓存
}

type LbMode string

const (
	WholeMode   LbMode = "whole"        // 作为默认选择模式。即 v4/v6 相等比例，正常的遍历选取。 向前兼容。即: 没有配置或者配置值为空 与配置 whole 效果相同
	V6PriorMode LbMode = "v6prior"      // v6优先选择,即 v4:v6比例 1:99
	V4PriorMode LbMode = "balance-99-1" // v4优先选择,即 v4:v6比例 99:1
	// balance-xx-xx， 指定v4-v6的比例. 比如 "balance-9-1"，v4:v6 9:1
	// 指定的值不能为 0
	// 代码上要处理下xx，此处做定义示例
	//FollowMode LbMode = "follow"           // 跟随请求client的ip version （预留模式，修改影响大暂未开发支持）
)

type AuthExtra struct {
	Headers []string `json:"headers,omitempty" bson:"headers"`
}

type Auth struct {
	Vendor          string `json:"vendor,omitempty" bson:"vendor"`
	VendorID        string `json:"vendor_id,omitempty" bson:"vendor_id"`
	AccessKeyId     string `json:"accessKeyId,omitempty" bson:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret,omitempty" bson:"accessKeySecret"`
	*AuthExtra
}

type HeaderOp string

const (
	HeaderOpAdd HeaderOp = "add"
	HeaderOpSet HeaderOp = "set"
	HeaderOpDel HeaderOp = "del"
)

type HeaderControl struct {
	Op    HeaderOp `json:"op"    bson:"op"`
	Key   string   `json:"key"   bson:"key"`
	Value string   `json:"value" bson:"value"`
}

// 回源节点配置
type SourceConfig struct {
	Addr            string `json:"addr" bson:"addr"`
	Weight          int    `json:"weight" bson:"weight"`
	LbMode          LbMode `json:"lbMode" bson:"lbMode"`
	Backup          bool   `json:"backup" bson:"backup"`
	Host            string `json:"host" bson:"host"`
	MaxFails        int    `json:"maxFails" bson:"maxFails"`
	FailTimeoutMs   int    `json:"failTimeoutMs" bson:"failTimeoutMs"` // failTimeout 内失败次数达到 maxFails 认为 node 不可用。
	PunishTimeoutMs int64  `json:"punishTimeoutMs" bson:"punishTimeoutMs"`
	EnableAuth      bool   `json:"enableAuth,omitempty" bson:"enableAuth"`
	Auth            Auth   `json:"auth,omitempty" bson:"auth"`
	//EnableRawUri    bool            `json:"enableRawUri,omitempty" bson:"enableRawUri"`     // 是否继续(URLRewrites) 改写前的URI
	HeaderControls []HeaderControl `json:"headerControls,omitempty" bson:"headerControls"` // 回源请求头增删改
}

type URLSchemeType string

const (
	HTTPScheme   URLSchemeType = "http"
	HTTPSScheme  URLSchemeType = "https"
	FollowScheme URLSchemeType = "follow"
)

// 回源配置
type DomainProxyConfig struct {
	Source          SourceConfig  `json:"source" bson:"source"`
	SourceHost      string        `json:"sourceHost" bson:"sourceHost"` // 回源 Host 头，不配置则使用域名 Host 头
	SourceURLScheme URLSchemeType `json:"sourceURLScheme" bson:"sourceURLScheme"`
	// to add
	// URLRewrites
	// 参数忽略
	// 回源超时
}

// 访问控制配置
type DomainACLConfig struct {
	// to add
	// 黑白名单
	// 防盗链
	// 鉴权
	// referer

	// scheme  // http or https or must https
	// IPV4 / IPV6 only

}

// https http2 证书配置
// 证书配置 存放，存放另外DB
type DomainCertConfig struct {
	// to add
	// https http2 证书配置
	CertID   string `json:"certID" bson:"certID"`
	FreeCert bool   `json:"freeCert" bson:"freeCert"`
}

// CDNDomain represents a CDN domain configuration
// 域名在cdn上的域名相关配置
type CDNDomain struct {
	UID  string
	Name string
	// 泛子域名的父域名，ParentName 不为空表示 Name 是泛子域名。
	// 泛子域名只能修改回源配置，其他配置使用父域名的配置。
	ParentName string
	Status     DomainStatusConfig
	Cache      DomainCacheConfig
	Proxy      DomainProxyConfig
	ACL        DomainACLConfig
	// to add
}

// cname 记录

// cdn 提供商

// 线路 -->

// 备案

// cache 缓存

// source 源

// 访问：黑白名单 防盗链 鉴权

// https http2 证书

// 302 配置
