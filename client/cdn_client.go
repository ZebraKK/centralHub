package client

import (
	"context"
	"fmt"
	"time"

	"centralHub/logger"
	"centralHub/model"

	volc "github.com/volcengine/volc-sdk-golang/service/cdn"
)

// CDNClientInterface defines the interface for CDN operations
type CDNClientInterface interface {
	// CreateCDNDomain creates a new CDN domain configuration
	CreateCDNDomain(ctx context.Context, domain string, config *model.CDNDomain) error

	// GetCDNDomainConfig retrieves the CDN configuration for a domain
	GetCDNDomainConfig(ctx context.Context, domain string) (*model.CDNDomain, error)

	// UpdateCDNDomainConfig updates the CDN configuration for a domain
	UpdateCDNDomainConfig(ctx context.Context, domain string, config *model.CDNDomain) error

	// DeleteCDNDomain removes a domain from CDN
	DeleteCDNDomain(ctx context.Context, domain string) error

	// PurgeCDNCache purges the CDN cache for specified paths
	PurgeCDNCache(ctx context.Context, domain string, paths []string) error

	// PreloadCDNContent preloads content into CDN cache
	PreloadCDNContent(ctx context.Context, domain string, paths []string) error
}

// RealCDNClient implements CDNClientInterface using Volcengine CDN service
type RealCDNClient struct {
	instance   *volc.CDN
	apiURL     string
	apiKey     string
	apiTimeout time.Duration
}

// NewRealCDNClient creates a new instance of RealCDNClient
func NewRealCDNClient(timeout time.Duration, apiURL, apiKey string) CDNClientInterface {
	ins := volc.NewInstance()

	// Note: Configuration would need to be handled according to the actual Volcengine SDK API
	// This is a placeholder until we implement the real SDK integration

	return &RealCDNClient{
		instance:   ins,
		apiURL:     apiURL,
		apiKey:     apiKey,
		apiTimeout: timeout,
	}
}

// CreateCDNDomain implements CDNClientInterface.CreateCDNDomain
func (c *RealCDNClient) CreateCDNDomain(ctx context.Context, domain string, config *model.CDNDomain) error {
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Creating CDN domain")

	// TODO: Convert our model.CDNDomain to Volcengine's format
	// TODO: Call Volcengine API

	return nil
}

// GetCDNDomainConfig implements CDNClientInterface.GetCDNDomainConfig
func (c *RealCDNClient) GetCDNDomainConfig(ctx context.Context, domain string) (*model.CDNDomain, error) {
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Getting CDN domain config")

	// TODO: Call Volcengine API
	// TODO: Convert Volcengine response to our model.CDNDomain

	return &model.CDNDomain{
		Name: domain,
	}, nil
}

// UpdateCDNDomainConfig implements CDNClientInterface.UpdateCDNDomainConfig
func (c *RealCDNClient) UpdateCDNDomainConfig(ctx context.Context, domain string, config *model.CDNDomain) error {
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Updating CDN domain config")

	// TODO: Convert our model.CDNDomain to Volcengine's format
	// TODO: Call Volcengine API

	return nil
}

// DeleteCDNDomain implements CDNClientInterface.DeleteCDNDomain
func (c *RealCDNClient) DeleteCDNDomain(ctx context.Context, domain string) error {
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Deleting CDN domain")

	// TODO: Call Volcengine API

	return nil
}

// PurgeCDNCache implements CDNClientInterface.PurgeCDNCache
func (c *RealCDNClient) PurgeCDNCache(ctx context.Context, domain string, paths []string) error {
	logger.RunLogger.Info().
		Str("domain", domain).
		Interface("paths", paths).
		Msg("Purging CDN cache")

	// TODO: Call Volcengine API

	return nil
}

// PreloadCDNContent implements CDNClientInterface.PreloadCDNContent
func (c *RealCDNClient) PreloadCDNContent(ctx context.Context, domain string, paths []string) error {
	logger.RunLogger.Info().
		Str("domain", domain).
		Interface("paths", paths).
		Msg("Preloading CDN content")

	// TODO: Call Volcengine API

	return nil
}

// MockCDNClient implements CDNClientInterface for testing
type MockCDNClient struct {
	serverURL string
	client    *HTTPClient
}

// NewMockCDNClient creates a new instance of MockCDNClient
func NewMockCDNClient(serverURL string) CDNClientInterface {
	httpClient := NewHTTPClient(
		WithTimeout(5 * time.Second),
	)

	return &MockCDNClient{
		serverURL: serverURL,
		client:    httpClient,
	}
}

// CreateCDNDomain implements CDNClientInterface.CreateCDNDomain for MockCDNClient
func (c *MockCDNClient) CreateCDNDomain(ctx context.Context, domain string, config *model.CDNDomain) error {
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Mock: Creating CDN domain")

	url := c.serverURL + "/api/cdn/domains"
	body := map[string]interface{}{
		"domain": domain,
		"config": config,
	}
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	err := c.client.PostJSON(ctx, url, body, nil, headers)
	return err
}

// GetCDNDomainConfig implements CDNClientInterface.GetCDNDomainConfig for MockCDNClient
func (c *MockCDNClient) GetCDNDomainConfig(ctx context.Context, domain string) (*model.CDNDomain, error) {
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Mock: Getting CDN domain config")

	url := c.serverURL + "/api/cdn/domains/" + domain
	headers := map[string]string{
		"Accept": "application/json",
	}

	var result struct {
		Success bool             `json:"success"`
		Message string           `json:"message"`
		Data    *model.CDNDomain `json:"data"`
	}

	err := c.client.GetJSON(ctx, url, &result, headers)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("server returned error: %s", result.Message)
	}

	return result.Data, nil
}

// UpdateCDNDomainConfig implements CDNClientInterface.UpdateCDNDomainConfig for MockCDNClient
func (c *MockCDNClient) UpdateCDNDomainConfig(ctx context.Context, domain string, config *model.CDNDomain) error {
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Mock: Updating CDN domain config")

	url := c.serverURL + "/api/cdn/domains/" + domain
	body := map[string]interface{}{
		"config": config,
	}
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	err := c.client.PutJSON(ctx, url, body, nil, headers)
	return err
}

// DeleteCDNDomain implements CDNClientInterface.DeleteCDNDomain for MockCDNClient
func (c *MockCDNClient) DeleteCDNDomain(ctx context.Context, domain string) error {
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Mock: Deleting CDN domain")

	url := c.serverURL + "/api/cdn/domains/" + domain
	headers := map[string]string{}

	resp, err := c.client.Delete(ctx, url, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("delete request failed with status %d", resp.StatusCode)
	}

	return nil
}

// PurgeCDNCache implements CDNClientInterface.PurgeCDNCache for MockCDNClient
func (c *MockCDNClient) PurgeCDNCache(ctx context.Context, domain string, paths []string) error {
	logger.RunLogger.Info().
		Str("domain", domain).
		Interface("paths", paths).
		Msg("Mock: Purging CDN cache")

	url := c.serverURL + "/api/cdn/purge"
	body := map[string]interface{}{
		"domain": domain,
		"paths":  paths,
	}
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	err := c.client.PostJSON(ctx, url, body, nil, headers)
	return err
}

// PreloadCDNContent implements CDNClientInterface.PreloadCDNContent for MockCDNClient
func (c *MockCDNClient) PreloadCDNContent(ctx context.Context, domain string, paths []string) error {
	logger.RunLogger.Info().
		Str("domain", domain).
		Interface("paths", paths).
		Msg("Mock: Preloading CDN content")

	url := c.serverURL + "/api/cdn/preload"
	body := map[string]interface{}{
		"domain": domain,
		"paths":  paths,
	}
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	err := c.client.PostJSON(ctx, url, body, nil, headers)
	return err
}
