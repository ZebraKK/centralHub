package hubserver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"centralHub/client"
	"centralHub/config"
	"centralHub/logger"
	"centralHub/store"
	"centralHub/workflow"
)

// 错误定义
var (
	ErrDomainStoreInit    = errors.New("failed to initialize domain store")
	ErrOwnershipStoreInit = errors.New("failed to initialize ownership store")
	ErrDNSClientInit      = errors.New("failed to initialize DNS client")
	ErrHTTPClientInit     = errors.New("failed to initialize HTTP client")
	ErrNilConfig          = errors.New("nil configuration")
)

// HubServer 主服务结构体
type HubServer struct {
	workflow       *workflow.Workflow
	domainStore    *store.DomainStore
	ownershipStore *store.OwnershipStore
	dnsClient      client.DNSClientInterface
	httpClient     *client.HTTPClient
	icpClient      client.ICPClient
	config         *config.Config
}

// NewHubServer 创建HubServer实例
func NewHubServer(cfg *config.Config) (*HubServer, error) {
	if cfg == nil {
		logger.RunLogger.Error().Msg("Configuration is nil")
		return nil, ErrNilConfig
	}

	// 初始化域名存储
	domainStore, err := store.NewDomainStore(cfg)
	if err != nil {
		logger.RunLogger.Error().Err(err).Msg("Failed to initialize domain store")
		return nil, fmt.Errorf("%w: %v", ErrDomainStoreInit, err)
	}

	// 初始化域名所有权验证存储
	ownershipStore := store.NewOwnershipStore(
		domainStore.GetClient(),
		domainStore.GetDB(),
		cfg.Database.MongoDB.Timeout,
	)
	if ownershipStore == nil {
		logger.RunLogger.Error().Msg("Failed to initialize ownership store")
		return nil, ErrOwnershipStoreInit
	}

	// 初始化DNS客户端
	var dnsClient client.DNSClientInterface
	if cfg.Features.UseMockServices {
		// 使用Mock DNS客户端
		mockServerURL := fmt.Sprintf("http://localhost:%d", cfg.MockServices.DNS.Port)
		dnsClient = client.NewMockDNSClient(mockServerURL)
		logger.RunLogger.Info().
			Str("type", "mock").
			Str("url", mockServerURL).
			Msg("Using Mock DNS client")
	} else {
		// 使用真实DNS客户端
		dnsClient = client.NewRealDNSClient(
			time.Duration(cfg.Server.Timeout)*time.Second,
			cfg.ExternalServices.DNS.APIURL,
			cfg.ExternalServices.DNS.APIKey,
		)
		logger.RunLogger.Info().
			Str("type", "real").
			Msg("Using Real DNS client")
	}

	// 初始化HTTP客户端
	httpClient := client.NewHTTPClient(
		client.WithTimeout(time.Duration(cfg.Server.Timeout) * time.Second),
	)
	if httpClient == nil {
		logger.RunLogger.Error().Msg("Failed to initialize HTTP client")
		return nil, ErrHTTPClientInit
	}

	// 初始化ICP客户端
	var icpClient client.ICPClient
	if cfg.Features.UseMockServices {
		// 使用Mock ICP客户端
		icpConfig := map[string]string{
			"mock_server_url": fmt.Sprintf("http://localhost:%d", cfg.MockServices.ICP.Port),
		}
		icpClient = client.NewICPClient(true, icpConfig, time.Duration(cfg.Server.Timeout)*time.Second)
		logger.RunLogger.Info().
			Str("type", "mock").
			Int("port", cfg.MockServices.ICP.Port).
			Msg("Using Mock ICP client")
	} else {
		// 使用真实ICP客户端
		icpConfig := map[string]string{
			"api_url":    cfg.ExternalServices.ICP.APIURL,
			"api_key":    cfg.ExternalServices.ICP.APIKey,
			"api_secret": cfg.ExternalServices.ICP.APISecret,
		}
		icpClient = client.NewICPClient(false, icpConfig, time.Duration(cfg.Server.Timeout)*time.Second)
		logger.RunLogger.Info().
			Str("type", "real").
			Msg("Using Real ICP client")
	}

	// 初始化CDN客户端
	var cdnClient client.CDNClientInterface
	if cfg.Features.UseMockServices {
		// 使用Mock CDN客户端
		mockServerURL := fmt.Sprintf("http://localhost:%d", cfg.MockServices.CDN.Port)
		cdnClient = client.NewMockCDNClient(mockServerURL)
		logger.RunLogger.Info().
			Str("type", "mock").
			Str("url", mockServerURL).
			Msg("Using Mock CDN client")
	} else {
		// 使用真实CDN客户端
		// 注意：RealCDNClient当前是占位符实现，需要进一步集成Volcengine SDK
		cdnClient = client.NewRealCDNClient(
			time.Duration(cfg.Server.Timeout)*time.Second,
			cfg.ExternalServices.CDN.APIURL,
			cfg.ExternalServices.CDN.APIKey,
		)
		logger.RunLogger.Info().
			Str("type", "real").
			Str("api_url", cfg.ExternalServices.CDN.APIURL).
			Msg("Using Real CDN client (placeholder)")
	}

	// 初始化工作流，传入客户端
	wf := workflow.NewWorkflow(
		workflow.WithICPClient(icpClient),
		workflow.WithDNSClient(dnsClient),
		workflow.WithCDNClient(cdnClient),
		workflow.WithICPCacheTTL(30*time.Minute), // 30分钟缓存
	)

	logger.RunLogger.Info().Msg("HubServer initialized successfully")
	// 创建 HubServer 实例
	hs := &HubServer{
		workflow:       wf,
		domainStore:    domainStore,
		ownershipStore: ownershipStore,
		dnsClient:      dnsClient,
		httpClient:     httpClient,
		icpClient:      icpClient,
		config:         cfg,
	}

	// 启动定期清理过期验证记录的任务
	// 每小时执行一次清理
	ownershipStore.ScheduleCleanup(context.Background(), 1*time.Hour)

	// 立即执行一次过期记录状态更新
	go func() {
		count, err := ownershipStore.UpdateExpiredVerifications(context.Background())
		if err != nil {
			logger.RunLogger.Error().
				Err(err).
				Msg("Initial update of expired verification records failed")
		} else {
			logger.RunLogger.Info().
				Int64("updated", count).
				Msg("Initial update of expired verification records completed")
		}
	}()

	return hs, nil
}

// Close 关闭HubServer，释放资源
func (hs *HubServer) Close() error {
	var errs []error

	// 关闭域名存储
	if hs.domainStore != nil {
		if err := hs.domainStore.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close domain store: %w", err))
		}
	}

	// 如果有错误，返回第一个错误
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// getOwnership 获取域名所有权信息
func (hs *HubServer) getOwnership(name string) (string, error) {
	// 使用域名存储查询所有权信息
	if hs.domainStore == nil {
		return "", errors.New("domain store is nil")
	}

	// 查询域名
	domain, err := hs.domainStore.FindByName(context.Background(), name)
	if err != nil {
		// 如果是找不到域名的错误，返回空字符串
		if errors.Is(err, errors.New("domain not found")) {
			return "", nil
		}
		return "", fmt.Errorf("query domain ownership: %w", err)
	}

	// 返回域名所有者
	return domain.PlatformInfo.UserID, nil
}
