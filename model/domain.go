package models

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
