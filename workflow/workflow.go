package workflow

import (
	"centralHub/client"
)

type VendorClient interface {
	// 定义 VendorClient 接口的方法
	GetVendorName(params ...interface{}) string
	CreateDomain(params ...interface{}) error
}

type Workflow struct {
	vendorClients map[string]VendorClient
}

func NewWorkflow() *Workflow {
	cltDict := make(map[string]VendorClient)
	cltDict["mock-vendor"] = client.NewMockClient()
	return &Workflow{
		vendorClients: cltDict,
	}
}

// 待后续集成 https://github.com/ZebraKK/workflow
func (ws *Workflow) PushTask() string {

	return "task-12345"
}

func (wf *Workflow) getVendorClient(vendor string) VendorClient {
	clt, ok := wf.vendorClients[vendor]
	if !ok {
		return nil
	}
	return clt
}
