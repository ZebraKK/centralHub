package model

// VerificationType 域名所有权验证类型
type VerificationType string

const (
	// DNSVerification DNS TXT记录验证
	DNSVerification VerificationType = "dns"
	// FileVerification 文件上传验证
	FileVerification VerificationType = "file"
)

// VerificationStatus 验证状态
type VerificationStatus string

const (
	// StatusPending 等待验证
	StatusPending VerificationStatus = "pending"
	// StatusVerified 验证成功
	StatusVerified VerificationStatus = "verified"
	// StatusFailed 验证失败
	StatusFailed VerificationStatus = "failed"
	// StatusExpired 验证过期
	StatusExpired VerificationStatus = "expired"
)

// OwnershipVerificationRequest 域名所有权验证请求
type OwnershipVerificationRequest struct {
	Name       string           `json:"name" binding:"required"`        // 域名名称
	VerifyType VerificationType `json:"verify_type" binding:"required"` // 验证类型
	UserID     string           `json:"user_id"`                        // 用户ID
}

// OwnershipVerificationResponse 域名所有权验证响应
type OwnershipVerificationResponse struct {
	Name       string           `json:"name"`        // 域名名称
	VerifyType VerificationType `json:"verify_type"` // 验证类型
	Value      string           `json:"value"`       // 验证值
	ReqID      string           `json:"req_id"`      // 请求ID
	ExpireAt   int64            `json:"expire_at"`   // 过期时间
}

// OwnershipVerifyRequest 验证状态查询请求
type OwnershipVerifyRequest struct {
	Name  string `json:"name" binding:"required"`   // 域名名称
	ReqID string `json:"req_id" binding:"required"` // 请求ID
}

// OwnershipVerifyResponse 验证状态查询响应
type OwnershipVerifyResponse struct {
	Name   string             `json:"name"`    // 域名名称
	Status VerificationStatus `json:"status"`  // 验证状态
	ReqID  string             `json:"req_id"`  // 请求ID
	UserID string             `json:"user_id"` // 用户ID
}

// OwnershipStatsResponse 域名所有权验证统计信息响应
type OwnershipStatsResponse struct {
	Total    int64 `json:"total"`    // 总记录数
	Pending  int64 `json:"pending"`  // 待验证记录数
	Verified int64 `json:"verified"` // 已验证记录数
	Failed   int64 `json:"failed"`   // 验证失败记录数
	Expired  int64 `json:"expired"`  // 已过期记录数
}

// OwnershipRecord 域名所有权验证记录
type OwnershipRecord struct {
	ID         string             `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	UserID     string             `bson:"user_id"`
	VerifyType VerificationType   `bson:"verify_type"`
	Value      string             `bson:"value"`
	Status     VerificationStatus `bson:"status"`
	CreatedAt  int64              `bson:"created_at"`
	UpdatedAt  int64              `bson:"updated_at"`
	ExpireAt   int64              `bson:"expire_at"`
}

// DNSVerificationConfig DNS验证配置
type DNSVerificationConfig struct {
	TTL            int   `json:"ttl"`             // TTL时间（秒）
	ValidityPeriod int64 `json:"validity_period"` // 有效期（秒）
	RetryInterval  int   `json:"retry_interval"`  // 重试间隔（秒）
	MaxRetries     int   `json:"max_retries"`     // 最大重试次数
}

// FileVerificationConfig 文件验证配置
type FileVerificationConfig struct {
	ValidityPeriod int64 `json:"validity_period"` // 有效期（秒）
	RetryInterval  int   `json:"retry_interval"`  // 重试间隔（秒）
	MaxRetries     int   `json:"max_retries"`     // 最大重试次数
}

// VerificationConfig 验证配置
type VerificationConfig struct {
	DNS  DNSVerificationConfig  `json:"dns"`
	File FileVerificationConfig `json:"file"`
}

// DefaultVerificationConfig 默认验证配置
var DefaultVerificationConfig = VerificationConfig{
	DNS: DNSVerificationConfig{
		TTL:            300,
		ValidityPeriod: 24 * 3600, // 24小时
		RetryInterval:  60,        // 60秒
		MaxRetries:     10,
	},
	File: FileVerificationConfig{
		ValidityPeriod: 24 * 3600, // 24小时
		RetryInterval:  60,        // 60秒
		MaxRetries:     10,
	},
}
