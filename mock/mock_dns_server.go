package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"centralHub/logger"
)

// DNSRecord represents a DNS record
type DNSRecord struct {
	ID         string `json:"id"`
	Domain     string `json:"domain"`
	RecordType string `json:"record_type"` // A, CNAME, TXT, etc.
	Value      string `json:"value"`
	TTL        int    `json:"ttl"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

// MockDNSServer DNS mock service
type MockDNSServer struct {
	*BaseMockServer
	dnsRecords      map[string][]DNSRecord
	recordsMutex    sync.RWMutex
	recordIDCounter int64
}

// NewMockDNSServer creates a new DNS mock server
func NewMockDNSServer(config MockConfig) *MockDNSServer {
	baseServer := NewBaseMockServer("dns", config)
	server := &MockDNSServer{
		BaseMockServer:  baseServer,
		dnsRecords:      make(map[string][]DNSRecord),
		recordIDCounter: 1000,
	}

	// Set up DNS routes
	server.setupRoutes()

	return server
}

// setupRoutes sets up the routes for the DNS mock server
func (s *MockDNSServer) setupRoutes() {
	// DNS API
	dns := s.Router.Group("/dns")
	{
		// Get all records for a domain
		dns.GET("/records/:domain", s.handleGetRecords)

		// Get a specific record
		dns.GET("/record/:domain/:id", s.handleGetRecord)

		// Add a record
		dns.POST("/record", s.handleAddRecord)

		// Update a record
		dns.PUT("/record/:id", s.handleUpdateRecord)

		// Delete a record
		dns.DELETE("/record/:domain/:id", s.handleDeleteRecord)

		// Verify TXT record
		dns.GET("/verify/txt", s.handleVerifyTXTRecord)

		// Admin API
		admin := dns.Group("/admin")
		{
			// Get all domains and their records
			admin.GET("/domains", s.handleGetAllDomains)

			// Get records for all domains
			admin.GET("/records", s.handleGetAllRecords)

			// Import records
			admin.POST("/import", s.handleImportRecords)

			// Export records
			admin.GET("/export", s.handleExportRecords)

			// Clear all records
			admin.DELETE("/records", s.handleClearAllRecords)
		}
	}
}

// handleGetRecords handles the request to get all records for a domain
func (s *MockDNSServer) handleGetRecords(c *gin.Context) {
	domain := strings.ToLower(c.Param("domain"))
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain parameter is required"})
		return
	}

	// Get records
	s.recordsMutex.RLock()
	records, exists := s.dnsRecords[domain]
	s.recordsMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusOK, []DNSRecord{})
		return
	}

	c.JSON(http.StatusOK, records)
}

// handleGetRecord handles the request to get a specific record
func (s *MockDNSServer) handleGetRecord(c *gin.Context) {
	domain := strings.ToLower(c.Param("domain"))
	id := c.Param("id")
	if domain == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain and id parameters are required"})
		return
	}

	// Get records
	s.recordsMutex.RLock()
	records, exists := s.dnsRecords[domain]
	s.recordsMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	// Find specific record
	for _, record := range records {
		if record.ID == id {
			c.JSON(http.StatusOK, record)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
}

// handleAddRecord handles the request to add a record
func (s *MockDNSServer) handleAddRecord(c *gin.Context) {
	var record DNSRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate record
	if record.Domain == "" || record.RecordType == "" || record.Value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain, record_type, and value are required"})
		return
	}

	// Normalize domain
	record.Domain = strings.ToLower(strings.TrimSpace(record.Domain))

	// Set defaults
	if record.TTL <= 0 {
		record.TTL = 600 // Default TTL: 10 minutes
	}

	// Generate ID and timestamps
	now := time.Now().Unix()
	s.recordIDCounter++
	record.ID = fmt.Sprintf("%d", s.recordIDCounter)
	record.CreatedAt = now
	record.UpdatedAt = now

	// Add record
	s.recordsMutex.Lock()
	if _, exists := s.dnsRecords[record.Domain]; !exists {
		s.dnsRecords[record.Domain] = make([]DNSRecord, 0)
	}
	s.dnsRecords[record.Domain] = append(s.dnsRecords[record.Domain], record)
	s.recordsMutex.Unlock()

	logger.RunLogger.Info().
		Str("domain", record.Domain).
		Str("type", record.RecordType).
		Str("value", record.Value).
		Int("ttl", record.TTL).
		Msg("Added DNS record")

	c.JSON(http.StatusOK, record)
}

// handleUpdateRecord handles the request to update a record
func (s *MockDNSServer) handleUpdateRecord(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	var updateRecord DNSRecord
	if err := c.ShouldBindJSON(&updateRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find and update record
	s.recordsMutex.Lock()
	defer s.recordsMutex.Unlock()

	found := false
	for domain, records := range s.dnsRecords {
		for i, record := range records {
			if record.ID == id {
				// Update fields
				if updateRecord.Value != "" {
					record.Value = updateRecord.Value
				}
				if updateRecord.TTL > 0 {
					record.TTL = updateRecord.TTL
				}
				record.UpdatedAt = time.Now().Unix()

				// Update record in the slice
				s.dnsRecords[domain][i] = record

				logger.RunLogger.Info().
					Str("domain", record.Domain).
					Str("type", record.RecordType).
					Str("id", record.ID).
					Msg("Updated DNS record")

				c.JSON(http.StatusOK, record)
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
	}
}

// handleDeleteRecord handles the request to delete a record
func (s *MockDNSServer) handleDeleteRecord(c *gin.Context) {
	domain := strings.ToLower(c.Param("domain"))
	id := c.Param("id")
	if domain == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain and id parameters are required"})
		return
	}

	// Delete record
	s.recordsMutex.Lock()
	defer s.recordsMutex.Unlock()

	records, exists := s.dnsRecords[domain]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		return
	}

	for i, record := range records {
		if record.ID == id {
			// Remove record from slice
			s.dnsRecords[domain] = append(records[:i], records[i+1:]...)

			logger.RunLogger.Info().
				Str("domain", domain).
				Str("id", id).
				Msg("Deleted DNS record")

			c.JSON(http.StatusOK, gin.H{"message": "record deleted"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
}

// handleVerifyTXTRecord handles the request to verify a TXT record
func (s *MockDNSServer) handleVerifyTXTRecord(c *gin.Context) {
	domain := strings.ToLower(c.Query("domain"))
	expectedValue := c.Query("value")
	if domain == "" || expectedValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain and value parameters are required"})
		return
	}

	// Get records
	s.recordsMutex.RLock()
	records, exists := s.dnsRecords[domain]
	s.recordsMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusOK, gin.H{"verified": false, "message": "domain not found"})
		return
	}

	// Check for matching TXT record
	for _, record := range records {
		if record.RecordType == "TXT" && record.Value == expectedValue {
			c.JSON(http.StatusOK, gin.H{"verified": true})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"verified": false, "message": "matching TXT record not found"})
}

// handleGetAllDomains handles the request to get all domains
func (s *MockDNSServer) handleGetAllDomains(c *gin.Context) {
	s.recordsMutex.RLock()
	defer s.recordsMutex.RUnlock()

	domains := make([]string, 0, len(s.dnsRecords))
	for domain := range s.dnsRecords {
		domains = append(domains, domain)
	}

	c.JSON(http.StatusOK, domains)
}

// handleGetAllRecords handles the request to get all records
func (s *MockDNSServer) handleGetAllRecords(c *gin.Context) {
	s.recordsMutex.RLock()
	defer s.recordsMutex.RUnlock()

	allRecords := make([]DNSRecord, 0)
	for _, records := range s.dnsRecords {
		allRecords = append(allRecords, records...)
	}

	c.JSON(http.StatusOK, allRecords)
}

// handleImportRecords handles the request to import records
func (s *MockDNSServer) handleImportRecords(c *gin.Context) {
	var records []DNSRecord
	if err := c.ShouldBindJSON(&records); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Import records
	s.recordsMutex.Lock()
	defer s.recordsMutex.Unlock()

	// Reset ID counter if needed
	maxID := s.recordIDCounter
	for _, record := range records {
		// Try to parse record ID as int64
		if id, err := parseInt64(record.ID); err == nil && id > maxID {
			maxID = id
		}
	}
	s.recordIDCounter = maxID

	// Group records by domain
	for _, record := range records {
		// Normalize domain
		record.Domain = strings.ToLower(strings.TrimSpace(record.Domain))

		// Set defaults
		if record.TTL <= 0 {
			record.TTL = 600
		}

		// Generate new ID if missing
		if record.ID == "" {
			s.recordIDCounter++
			record.ID = fmt.Sprintf("%d", s.recordIDCounter)
		}

		// Set timestamps if missing
		now := time.Now().Unix()
		if record.CreatedAt <= 0 {
			record.CreatedAt = now
		}
		if record.UpdatedAt <= 0 {
			record.UpdatedAt = now
		}

		// Add to records map
		if _, exists := s.dnsRecords[record.Domain]; !exists {
			s.dnsRecords[record.Domain] = make([]DNSRecord, 0)
		}
		s.dnsRecords[record.Domain] = append(s.dnsRecords[record.Domain], record)
	}

	logger.RunLogger.Info().
		Int("count", len(records)).
		Msg("Imported DNS records")

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Imported %d DNS records", len(records)),
		"count":   len(records),
	})
}

// handleExportRecords handles the request to export records
func (s *MockDNSServer) handleExportRecords(c *gin.Context) {
	s.recordsMutex.RLock()
	defer s.recordsMutex.RUnlock()

	allRecords := make([]DNSRecord, 0)
	for _, records := range s.dnsRecords {
		allRecords = append(allRecords, records...)
	}

	c.JSON(http.StatusOK, allRecords)
}

// handleClearAllRecords handles the request to clear all records
func (s *MockDNSServer) handleClearAllRecords(c *gin.Context) {
	s.recordsMutex.Lock()
	count := 0
	for _, records := range s.dnsRecords {
		count += len(records)
	}
	s.dnsRecords = make(map[string][]DNSRecord)
	s.recordsMutex.Unlock()

	logger.RunLogger.Info().
		Int("count", count).
		Msg("Cleared all DNS records")

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Cleared %d DNS records", count),
		"count":   count,
	})
}

// LoadDefaultDNSRecords loads default DNS records
func (s *MockDNSServer) LoadDefaultDNSRecords() error {
	now := time.Now().Unix()

	// Default DNS records
	defaultRecords := []DNSRecord{
		{
			ID:         "1001",
			Domain:     "example.com",
			RecordType: "A",
			Value:      "192.0.2.1",
			TTL:        3600,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "1002",
			Domain:     "example.com",
			RecordType: "MX",
			Value:      "10 mail.example.com",
			TTL:        3600,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "1003",
			Domain:     "example.com",
			RecordType: "TXT",
			Value:      "v=spf1 ip4:192.0.2.0/24 ~all",
			TTL:        3600,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "1004",
			Domain:     "www.example.com",
			RecordType: "CNAME",
			Value:      "example.com",
			TTL:        3600,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "1005",
			Domain:     "example.org",
			RecordType: "A",
			Value:      "192.0.2.2",
			TTL:        3600,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "1006",
			Domain:     "example.net",
			RecordType: "A",
			Value:      "192.0.2.3",
			TTL:        3600,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	// Import records
	s.recordsMutex.Lock()
	defer s.recordsMutex.Unlock()

	// Reset records
	s.dnsRecords = make(map[string][]DNSRecord)
	s.recordIDCounter = 1006 // Set to the highest ID in the default records

	// Group records by domain
	for _, record := range defaultRecords {
		domain := record.Domain
		if _, exists := s.dnsRecords[domain]; !exists {
			s.dnsRecords[domain] = make([]DNSRecord, 0)
		}
		s.dnsRecords[domain] = append(s.dnsRecords[domain], record)
	}

	logger.RunLogger.Info().
		Int("count", len(defaultRecords)).
		Msg("Loaded default DNS records")

	return nil
}

// LoadDNSRecordsFromFile loads DNS records from a file
func (s *MockDNSServer) LoadDNSRecordsFromFile(path string) error {
	// Load from file
	data, err := loadFromFile(path)
	if err != nil {
		return fmt.Errorf("load from file: %w", err)
	}

	// Unmarshal records
	var records []DNSRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("unmarshal records: %w", err)
	}

	// Import records
	s.recordsMutex.Lock()
	defer s.recordsMutex.Unlock()

	// Reset records
	s.dnsRecords = make(map[string][]DNSRecord)

	// Reset ID counter
	s.recordIDCounter = 1000

	// Group records by domain
	for _, record := range records {
		// Normalize domain
		record.Domain = strings.ToLower(strings.TrimSpace(record.Domain))

		// Try to parse record ID as int64
		if id, err := parseInt64(record.ID); err == nil && id > s.recordIDCounter {
			s.recordIDCounter = id
		}

		// Add to records map
		if _, exists := s.dnsRecords[record.Domain]; !exists {
			s.dnsRecords[record.Domain] = make([]DNSRecord, 0)
		}
		s.dnsRecords[record.Domain] = append(s.dnsRecords[record.Domain], record)
	}

	logger.RunLogger.Info().
		Int("count", len(records)).
		Str("path", path).
		Msg("Loaded DNS records from file")

	return nil
}

// SaveDNSRecordsToFile saves DNS records to a file
func (s *MockDNSServer) SaveDNSRecordsToFile(path string) error {
	s.recordsMutex.RLock()
	defer s.recordsMutex.RUnlock()

	// Convert to slice
	records := make([]DNSRecord, 0)
	for _, domainRecords := range s.dnsRecords {
		records = append(records, domainRecords...)
	}

	// Marshal records
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal records: %w", err)
	}

	// Save to file
	if err := saveToFile(path, data); err != nil {
		return fmt.Errorf("save to file: %w", err)
	}

	logger.RunLogger.Info().
		Int("count", len(records)).
		Str("path", path).
		Msg("Saved DNS records to file")

	return nil
}

// parseInt64 parses a string as int64
func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
