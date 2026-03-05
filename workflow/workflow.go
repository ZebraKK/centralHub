package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"centralHub/client"
	"centralHub/model"
)

type VendorClient interface {
	// 定义 VendorClient 接口的方法
	GetVendorName(params ...interface{}) string
	CreateDomain(params ...interface{}) error
}

// ICPCacheEntry ICP缓存条目
type ICPCacheEntry struct {
	Record    *model.ICPRecord
	Timestamp time.Time
}

type Workflow struct {
	vendorClients map[string]VendorClient
	icpClient     client.ICPClient
	dnsClient     client.DNSClientInterface
	cdnClient     client.CDNClientInterface

	// ICP缓存相关
	icpCache    map[string]*ICPCacheEntry
	icpCacheMu  sync.RWMutex
	icpCacheTTL time.Duration
}

// WorkflowOption 工作流配置选项
type WorkflowOption func(*Workflow)

// WithICPClient 设置ICP客户端
func WithICPClient(client client.ICPClient) WorkflowOption {
	return func(wf *Workflow) {
		wf.icpClient = client
	}
}

// WithDNSClient 设置DNS客户端
func WithDNSClient(client client.DNSClientInterface) WorkflowOption {
	return func(wf *Workflow) {
		wf.dnsClient = client
	}
}

// WithCDNClient 设置CDN客户端
func WithCDNClient(client client.CDNClientInterface) WorkflowOption {
	return func(wf *Workflow) {
		wf.cdnClient = client
	}
}

// WithICPCacheTTL 设置ICP缓存TTL
func WithICPCacheTTL(ttl time.Duration) WorkflowOption {
	return func(wf *Workflow) {
		wf.icpCacheTTL = ttl
	}
}

func NewWorkflow(options ...WorkflowOption) *Workflow {
	cltDict := make(map[string]VendorClient)
	cltDict["mock-vendor"] = client.NewMockClient()

	wf := &Workflow{
		vendorClients: cltDict,
		icpCache:      make(map[string]*ICPCacheEntry),
		icpCacheTTL:   30 * time.Minute, // 默认30分钟缓存
	}

	// 应用选项
	for _, option := range options {
		option(wf)
	}

	return wf
}

// GetICPFromCache 从缓存获取ICP记录
func (wf *Workflow) GetICPFromCache(domain string) (*model.ICPRecord, bool) {
	wf.icpCacheMu.RLock()
	defer wf.icpCacheMu.RUnlock()

	entry, exists := wf.icpCache[domain]
	if !exists {
		return nil, false
	}

	// 检查缓存是否过期
	if time.Since(entry.Timestamp) > wf.icpCacheTTL {
		return nil, false
	}

	return entry.Record, true
}

// SetICPToCache 设置ICP记录到缓存
func (wf *Workflow) SetICPToCache(domain string, record *model.ICPRecord) {
	wf.icpCacheMu.Lock()
	defer wf.icpCacheMu.Unlock()

	wf.icpCache[domain] = &ICPCacheEntry{
		Record:    record,
		Timestamp: time.Now(),
	}
}

// ClearExpiredICPCache 清理过期的ICP缓存
func (wf *Workflow) ClearExpiredICPCache() {
	wf.icpCacheMu.Lock()
	defer wf.icpCacheMu.Unlock()

	now := time.Now()
	for domain, entry := range wf.icpCache {
		if now.Sub(entry.Timestamp) > wf.icpCacheTTL {
			delete(wf.icpCache, domain)
		}
	}
}

// CheckICPStatus 检查域名ICP备案状态
// 先从缓存查找，如果缓存不存在或过期，则调用ICP客户端查询
func (wf *Workflow) CheckICPStatus(ctx context.Context, domain string) (*model.ICPRecord, error) {
	// 尝试从缓存获取
	if record, ok := wf.GetICPFromCache(domain); ok {
		return record, nil
	}

	// 缓存未命中，调用ICP客户端查询
	if wf.icpClient == nil {
		return nil, fmt.Errorf("ICP client not initialized")
	}

	record, err := wf.icpClient.QueryICP(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to query ICP: %w", err)
	}

	// 缓存查询结果
	wf.SetICPToCache(domain, record)

	return record, nil
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
