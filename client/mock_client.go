package client

// 万能的 Mock  Client
type MockClient struct {
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (mc *MockClient) GetVendorName(params ...interface{}) string {
	return "vendor-67890"
}

func (mc *MockClient) CreateDomain(params ...interface{}) error {
	// 模拟创建域名的逻辑
	return nil
}
