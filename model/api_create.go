package model

type CreateDomainRequest struct {
	Name   string            `json:"name"`
	UserID string            `json:"user_id"`
	Cache  DomainCacheConfig `json:"cache"`
	Proxy  DomainProxyConfig `json:"proxy"`
	ACL    DomainACLConfig   `json:"acl"`
}

type CreateDomainResponse struct {
	Name       string `json:"name"`
	JobID      string `json:"job_id"`
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}
