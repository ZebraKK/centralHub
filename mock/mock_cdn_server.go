package mock

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"centralHub/logger"
	"centralHub/model"
)

// MockCDNServer implements a mock CDN API server
type MockCDNServer struct {
	BaseMockServer
	domains        map[string]*model.CDNDomain
	purgeHistory   map[string][]string
	preloadHistory map[string][]string
	mu             sync.RWMutex
}

// NewMockCDNServer creates a new MockCDNServer instance
func NewMockCDNServer(cfg MockConfig) *MockCDNServer {
	s := &MockCDNServer{
		BaseMockServer: *NewBaseMockServer("cdn", cfg),
		domains:        make(map[string]*model.CDNDomain),
		purgeHistory:   make(map[string][]string),
		preloadHistory: make(map[string][]string),
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures the API routes for the mock CDN server
func (s *MockCDNServer) setupRoutes() {
	// Base server routes (already includes delay, error simulation, etc.)
	s.Router.POST("/api/cdn/domains", s.handleCreateDomain)
	s.Router.GET("/api/cdn/domains/:domain", s.handleGetDomainConfig)
	s.Router.PUT("/api/cdn/domains/:domain", s.handleUpdateDomainConfig)
	s.Router.DELETE("/api/cdn/domains/:domain", s.handleDeleteDomain)
	s.Router.POST("/api/cdn/purge", s.handlePurgeCache)
	s.Router.POST("/api/cdn/preload", s.handlePreloadContent)

	// Management API routes
	mgmt := s.Router.Group("/manage/cdn")
	mgmt.GET("/domains", s.handleListDomains)
	mgmt.GET("/purge_history", s.handleGetPurgeHistory)
	mgmt.GET("/preload_history", s.handleGetPreloadHistory)
	mgmt.POST("/reset", s.handleReset)
	mgmt.POST("/load_default_data", s.handleLoadDefaultData)
}

// handleCreateDomain handles CDN domain creation requests
func (s *MockCDNServer) handleCreateDomain(c *gin.Context) {
	var request struct {
		Domain string           `json:"domain"`
		Config *model.CDNDomain `json:"config"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if request.Domain == "" || request.Config == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Domain and config are required",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if domain already exists
	if _, exists := s.domains[request.Domain]; exists {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "Domain already exists",
		})
		return
	}

	// Set the domain name in the config
	request.Config.Name = request.Domain

	// Store the domain
	s.domains[request.Domain] = request.Config

	logger.RunLogger.Info().
		Str("domain", request.Domain).
		Msg("MockCDNServer: Created CDN domain")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Domain created successfully",
	})
}

// handleGetDomainConfig handles retrieving a domain's CDN configuration
func (s *MockCDNServer) handleGetDomainConfig(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Domain parameter is required",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if domain exists
	config, exists := s.domains[domain]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Domain not found",
		})
		return
	}

	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("MockCDNServer: Retrieved CDN domain config")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// handleUpdateDomainConfig handles updating a domain's CDN configuration
func (s *MockCDNServer) handleUpdateDomainConfig(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Domain parameter is required",
		})
		return
	}

	var request struct {
		Config *model.CDNDomain `json:"config"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if request.Config == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Config is required",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if domain exists
	_, exists := s.domains[domain]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Domain not found",
		})
		return
	}

	// Preserve the domain name
	request.Config.Name = domain

	// Update the domain
	s.domains[domain] = request.Config

	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("MockCDNServer: Updated CDN domain config")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Domain updated successfully",
	})
}

// handleDeleteDomain handles deleting a CDN domain
func (s *MockCDNServer) handleDeleteDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Domain parameter is required",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if domain exists
	_, exists := s.domains[domain]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Domain not found",
		})
		return
	}

	// Delete the domain
	delete(s.domains, domain)

	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("MockCDNServer: Deleted CDN domain")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Domain deleted successfully",
	})
}

// handlePurgeCache handles cache purge requests
func (s *MockCDNServer) handlePurgeCache(c *gin.Context) {
	var request struct {
		Domain string   `json:"domain"`
		Paths  []string `json:"paths"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if request.Domain == "" || len(request.Paths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Domain and paths are required",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if domain exists
	_, exists := s.domains[request.Domain]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Domain not found",
		})
		return
	}

	// Add to purge history
	s.purgeHistory[request.Domain] = append(s.purgeHistory[request.Domain], request.Paths...)

	logger.RunLogger.Info().
		Str("domain", request.Domain).
		Interface("paths", request.Paths).
		Msg("MockCDNServer: Purged CDN cache")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cache purged successfully",
		"task_id": strconv.FormatInt(time.Now().UnixNano(), 10),
	})
}

// handlePreloadContent handles content preload requests
func (s *MockCDNServer) handlePreloadContent(c *gin.Context) {
	var request struct {
		Domain string   `json:"domain"`
		Paths  []string `json:"paths"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if request.Domain == "" || len(request.Paths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Domain and paths are required",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if domain exists
	_, exists := s.domains[request.Domain]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Domain not found",
		})
		return
	}

	// Add to preload history
	s.preloadHistory[request.Domain] = append(s.preloadHistory[request.Domain], request.Paths...)

	logger.RunLogger.Info().
		Str("domain", request.Domain).
		Interface("paths", request.Paths).
		Msg("MockCDNServer: Preloaded CDN content")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Content preloaded successfully",
		"task_id": strconv.FormatInt(time.Now().UnixNano(), 10),
	})
}

// handleListDomains handles listing all registered domains
func (s *MockCDNServer) handleListDomains(c *gin.Context) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	domains := make([]*model.CDNDomain, 0, len(s.domains))
	for _, domain := range s.domains {
		domains = append(domains, domain)
	}

	logger.RunLogger.Info().
		Int("count", len(domains)).
		Msg("MockCDNServer: Listed CDN domains")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    domains,
	})
}

// handleGetPurgeHistory handles retrieving purge history
func (s *MockCDNServer) handleGetPurgeHistory(c *gin.Context) {
	domain := c.Query("domain")

	s.mu.RLock()
	defer s.mu.RUnlock()

	var history map[string][]string
	if domain != "" {
		history = map[string][]string{
			domain: s.purgeHistory[domain],
		}
	} else {
		history = s.purgeHistory
	}

	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("MockCDNServer: Retrieved purge history")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}

// handleGetPreloadHistory handles retrieving preload history
func (s *MockCDNServer) handleGetPreloadHistory(c *gin.Context) {
	domain := c.Query("domain")

	s.mu.RLock()
	defer s.mu.RUnlock()

	var history map[string][]string
	if domain != "" {
		history = map[string][]string{
			domain: s.preloadHistory[domain],
		}
	} else {
		history = s.preloadHistory
	}

	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("MockCDNServer: Retrieved preload history")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}

// handleReset handles resetting the server state
func (s *MockCDNServer) handleReset(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.domains = make(map[string]*model.CDNDomain)
	s.purgeHistory = make(map[string][]string)
	s.preloadHistory = make(map[string][]string)

	logger.RunLogger.Info().Msg("MockCDNServer: Reset server state")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Server state reset successfully",
	})
}

// handleLoadDefaultData handles loading default test data
func (s *MockCDNServer) handleLoadDefaultData(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create some default CDN domains
	domains := []string{
		"example.com",
		"test.example.com",
		"demo.example.org",
	}

	for _, domainName := range domains {
		domain := &model.CDNDomain{
			Name: domainName,
			Cache: model.DomainCacheConfig{
				GlobalCacheTime: 3600,
				CacheKeyRule: model.CacheKeyRule{
					CacheKeyHost:   "include",
					CacheKeyQuery:  "include",
					CacheKeyHead:   "include",
					CacheKeyScheme: "exclude",
				},
			},
			Proxy: model.DomainProxyConfig{
				Source: model.SourceConfig{
					Addr:          "origin." + domainName,
					Weight:        100,
					LbMode:        model.WholeMode,
					Host:          domainName,
					MaxFails:      3,
					FailTimeoutMs: 10000,
				},
				SourceHost:      domainName,
				SourceURLScheme: model.HTTPSScheme,
			},
		}
		s.domains[domainName] = domain
	}

	logger.RunLogger.Info().Msg("MockCDNServer: Loaded default data")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Default data loaded successfully",
		"domains": domains,
	})
}

// GetDomainConfig retrieves a domain's CDN configuration (for internal use)
func (s *MockCDNServer) GetDomainConfig(domain string) *model.CDNDomain {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.domains[domain]
}

// HasDomain checks if a domain exists in the server
func (s *MockCDNServer) HasDomain(domain string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.domains[domain]
	return exists
}

// AddDomain adds a domain to the server (for internal use)
func (s *MockCDNServer) AddDomain(domain *model.CDNDomain) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.domains[domain.Name] = domain
}

// GetPurgeHistory retrieves the purge history for a domain (for internal use)
func (s *MockCDNServer) GetPurgeHistory(domain string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.purgeHistory[domain]
}

// GetPreloadHistory retrieves the preload history for a domain (for internal use)
func (s *MockCDNServer) GetPreloadHistory(domain string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.preloadHistory[domain]
}
