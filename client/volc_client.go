package client

import (
	volc "github.com/volcengine/volc-sdk-golang/service/cdn"
)

type VolcClient struct {
	instance *volc.CDN
}

func NewVolcClient() *VolcClient {
	ins := volc.NewInstance()

	return &VolcClient{
		instance: ins,
	}
}
