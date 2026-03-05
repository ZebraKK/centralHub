package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"centralHub/config"
	"centralHub/logger"
	models "centralHub/model"
	"centralHub/utils"
)

// MongoDB 集合名称常量
const (
	DomainCollectionName = "domains"
)

// 连接相关常量
const (
	DefaultConnectTimeout = 10 * time.Second
	DefaultPingTimeout    = 5 * time.Second
	MaxConnectRetries     = 3
	RetryInterval         = 2 * time.Second
)

// 错误定义
var (
	ErrNilConfig     = errors.New("nil configuration")
	ErrConnectFailed = errors.New("failed to connect to MongoDB")
	ErrPingFailed    = errors.New("failed to ping MongoDB")
)

// DomainStore 域名存储结构体
type DomainStore struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
	config     *config.MongoDBConfig
}

// GetClient 获取MongoDB客户端
func (ds *DomainStore) GetClient() *mongo.Client {
	return ds.client
}

// GetDB 获取MongoDB数据库
func (ds *DomainStore) GetDB() *mongo.Database {
	return ds.db
}

// NewDomainStore 创建域名存储实例
func NewDomainStore(cfg *config.Config) (*DomainStore, error) {
	if cfg == nil {
		logger.RunLogger.Error().Msg("Configuration is nil")
		return nil, ErrNilConfig
	}

	mongoConfig := cfg.Database.MongoDB
	logger.RunLogger.Debug().
		Str("uri", maskURI(mongoConfig.URI)).
		Str("database", mongoConfig.Database).
		Int("timeout", mongoConfig.Timeout).
		Msg("Initializing MongoDB connection")

	// 创建连接选项
	clientOptions := createClientOptions(mongoConfig)

	// 连接到 MongoDB（带重试）
	client, err := connectWithRetry(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectFailed, err)
	}

	// 获取数据库和集合
	db := client.Database(mongoConfig.Database)
	collection := db.Collection(DomainCollectionName)

	logger.RunLogger.Info().
		Str("database", mongoConfig.Database).
		Str("collection", DomainCollectionName).
		Msg("MongoDB connection established successfully")

	return &DomainStore{
		client:     client,
		db:         db,
		collection: collection,
		config:     &mongoConfig,
	}, nil
}

// createClientOptions 创建 MongoDB 客户端选项
func createClientOptions(mongoConfig config.MongoDBConfig) *options.ClientOptions {
	// 设置连接超时
	timeout := DefaultConnectTimeout
	if mongoConfig.Timeout > 0 {
		timeout = time.Duration(mongoConfig.Timeout) * time.Second
	}

	// 创建客户端选项
	clientOptions := options.Client().
		ApplyURI(mongoConfig.URI).
		SetConnectTimeout(timeout).
		SetServerSelectionTimeout(timeout).
		SetMaxPoolSize(100).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(30 * time.Minute)

	return clientOptions
}

// connectWithRetry 带重试的 MongoDB 连接
func connectWithRetry(clientOptions *options.ClientOptions) (*mongo.Client, error) {
	var client *mongo.Client
	var err error

	for attempt := 1; attempt <= MaxConnectRetries; attempt++ {
		// 创建上下文（带超时）
		ctx, cancel := context.WithTimeout(context.Background(), DefaultConnectTimeout)
		defer cancel()

		// 尝试连接
		client, err = mongo.Connect(ctx, clientOptions)
		if err == nil {
			// 连接成功，尝试 ping
			pingCtx, pingCancel := context.WithTimeout(context.Background(), DefaultPingTimeout)
			defer pingCancel()

			if err = client.Ping(pingCtx, readpref.Primary()); err == nil {
				// Ping 成功，返回客户端
				return client, nil
			}

			logger.RunLogger.Warn().
				Err(err).
				Int("attempt", attempt).
				Msg("Connected to MongoDB but ping failed")
		}

		logger.RunLogger.Warn().
			Err(err).
			Int("attempt", attempt).
			Int("max_attempts", MaxConnectRetries).
			Msg("Failed to connect to MongoDB, retrying...")

		// 如果不是最后一次尝试，等待后重试
		if attempt < MaxConnectRetries {
			time.Sleep(RetryInterval)
		}
	}

	return nil, fmt.Errorf("after %d attempts: %w", MaxConnectRetries, err)
}

// maskURI 掩盖 URI 中的敏感信息（用于日志）
func maskURI(uri string) string {
	// 简单实现，实际可能需要更复杂的正则表达式
	return uri // 在实际生产环境中应该实现掩盖密码的逻辑
}

// Close 关闭数据库连接
func (ds *DomainStore) Close() error {
	if ds.client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultConnectTimeout)
	defer cancel()

	if err := ds.client.Disconnect(ctx); err != nil {
		logger.RunLogger.Error().Err(err).Msg("Failed to disconnect MongoDB client")
		return err
	}

	logger.RunLogger.Info().Msg("MongoDB client disconnected")
	return nil
}

// 验证错误
var (
	ErrEmptyDomainName   = errors.New("domain name cannot be empty")
	ErrInvalidDomainName = errors.New("invalid domain name format")
	ErrEmptyUserID       = errors.New("user ID cannot be empty")
	ErrEmptyID           = errors.New("domain ID cannot be empty")
)

// validateDomain 验证域名数据
func validateDomain(domain models.XLDomain) error {
	// 验证域名名称
	if domain.DomainConfig.Name == "" {
		return ErrEmptyDomainName
	}

	// 验证域名格式（简单验证，实际可能需要更复杂的正则表达式）
	if !utils.IsValidDomainName(domain.DomainConfig.Name) {
		return ErrInvalidDomainName
	}

	// 验证用户ID
	if domain.PlatformInfo.UserID == "" {
		return ErrEmptyUserID
	}

	// 验证域名ID
	if domain.ID == "" {
		return ErrEmptyID
	}

	return nil
}

// Insert 插入域名记录
func (ds *DomainStore) Insert(ctx context.Context, domain models.XLDomain) error {
	if ds.collection == nil {
		return errors.New("MongoDB collection is nil")
	}

	// 验证域名数据
	if err := validateDomain(domain); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain.DomainConfig.Name).
			Msg("Domain validation failed")
		return fmt.Errorf("domain validation failed: %w", err)
	}

	// 创建超时上下文
	timeoutCtx, cancel := createTimeoutContext(ctx, ds.config.Timeout)
	defer cancel()

	// 检查域名是否已存在
	existingDomain, err := ds.FindByName(ctx, domain.DomainConfig.Name)
	if err == nil && existingDomain != nil {
		logger.RunLogger.Warn().
			Str("domain", domain.DomainConfig.Name).
			Msg("Domain already exists")
		return fmt.Errorf("domain already exists: %s", domain.DomainConfig.Name)
	}

	// 记录操作日志
	logger.RunLogger.Info().
		Str("domain", domain.DomainConfig.Name).
		Msg("Inserting domain")

	// 执行插入操作
	_, err = ds.collection.InsertOne(timeoutCtx, domain)
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain.DomainConfig.Name).
			Msg("Insert domain failed")
		return fmt.Errorf("insert domain failed: %w", err)
	}

	logger.RunLogger.Debug().
		Str("domain", domain.DomainConfig.Name).
		Msg("Domain inserted successfully")
	return nil
}

// FindByID 根据ID查找域名
func (ds *DomainStore) FindByID(ctx context.Context, id string) (*models.XLDomain, error) {
	if ds.collection == nil {
		return nil, errors.New("MongoDB collection is nil")
	}

	// 创建超时上下文
	timeoutCtx, cancel := createTimeoutContext(ctx, ds.config.Timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("id", id).
		Msg("Finding domain by ID")

	// 执行查询操作
	var domain models.XLDomain
	err := ds.collection.FindOne(timeoutCtx, bson.M{"_id": id}).Decode(&domain)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			logger.RunLogger.Debug().
				Str("id", id).
				Msg("Domain not found")
			return nil, fmt.Errorf("domain not found: %w", err)
		}

		logger.RunLogger.Error().
			Err(err).
			Str("id", id).
			Msg("Find domain failed")
		return nil, fmt.Errorf("find domain failed: %w", err)
	}

	logger.RunLogger.Debug().
		Str("id", id).
		Msg("Domain found")
	return &domain, nil
}

// Update 更新域名记录
func (ds *DomainStore) Update(ctx context.Context, id string, update bson.M) error {
	if ds.collection == nil {
		return errors.New("MongoDB collection is nil")
	}

	// 验证ID
	if id == "" {
		return ErrEmptyID
	}

	// 验证更新内容
	if len(update) == 0 {
		return errors.New("update cannot be empty")
	}

	// 检查是否包含域名名称更新，如果有则验证格式
	if domainConfig, ok := update["domainConfig"].(map[string]interface{}); ok {
		if name, ok := domainConfig["name"].(string); ok {
			if name == "" {
				return ErrEmptyDomainName
			}
			if !utils.IsValidDomainName(name) {
				return ErrInvalidDomainName
			}
		}
	}

	// 创建超时上下文
	timeoutCtx, cancel := createTimeoutContext(ctx, ds.config.Timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("id", id).
		Interface("update", update).
		Msg("Updating domain")

	// 执行更新操作
	result, err := ds.collection.UpdateOne(timeoutCtx, bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("id", id).
			Msg("Update domain failed")
		return fmt.Errorf("update domain failed: %w", err)
	}

	if result.MatchedCount == 0 {
		logger.RunLogger.Warn().
			Str("id", id).
			Msg("No domain matched for update")
		return fmt.Errorf("domain not found for update")
	}

	logger.RunLogger.Debug().
		Str("id", id).
		Int64("matched", result.MatchedCount).
		Int64("modified", result.ModifiedCount).
		Msg("Domain updated successfully")
	return nil
}

// Delete 删除域名记录
func (ds *DomainStore) Delete(ctx context.Context, id string) error {
	if ds.collection == nil {
		return errors.New("MongoDB collection is nil")
	}

	// 创建超时上下文
	timeoutCtx, cancel := createTimeoutContext(ctx, ds.config.Timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("id", id).
		Msg("Deleting domain")

	// 执行删除操作
	result, err := ds.collection.DeleteOne(timeoutCtx, bson.M{"_id": id})
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("id", id).
			Msg("Delete domain failed")
		return fmt.Errorf("delete domain failed: %w", err)
	}

	if result.DeletedCount == 0 {
		logger.RunLogger.Warn().
			Str("id", id).
			Msg("No domain matched for deletion")
		return fmt.Errorf("domain not found for deletion")
	}

	logger.RunLogger.Debug().
		Str("id", id).
		Int64("deleted", result.DeletedCount).
		Msg("Domain deleted successfully")
	return nil
}

// FindByName 根据域名查找记录
func (ds *DomainStore) FindByName(ctx context.Context, name string) (*models.XLDomain, error) {
	if ds.collection == nil {
		return nil, errors.New("MongoDB collection is nil")
	}

	// 创建超时上下文
	timeoutCtx, cancel := createTimeoutContext(ctx, ds.config.Timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("name", name).
		Msg("Finding domain by name")

	// 执行查询操作
	var domain models.XLDomain
	err := ds.collection.FindOne(timeoutCtx, bson.M{"domainConfig.name": name}).Decode(&domain)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			logger.RunLogger.Debug().
				Str("name", name).
				Msg("Domain not found")
			return nil, fmt.Errorf("domain not found: %w", err)
		}

		logger.RunLogger.Error().
			Err(err).
			Str("name", name).
			Msg("Find domain by name failed")
		return nil, fmt.Errorf("find domain by name failed: %w", err)
	}

	logger.RunLogger.Debug().
		Str("name", name).
		Msg("Domain found")
	return &domain, nil
}

// ListDomains 列出所有域名（支持分页）
func (ds *DomainStore) ListDomains(ctx context.Context, skip, limit int64) ([]models.XLDomain, error) {
	if ds.collection == nil {
		return nil, errors.New("MongoDB collection is nil")
	}

	// 创建超时上下文
	timeoutCtx, cancel := createTimeoutContext(ctx, ds.config.Timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Int64("skip", skip).
		Int64("limit", limit).
		Msg("Listing domains")

	// 创建查询选项
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "platformInfo.createAt", Value: -1}}) // 按创建时间降序

	// 执行查询操作
	cursor, err := ds.collection.Find(timeoutCtx, bson.M{}, findOptions)
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Msg("List domains failed")
		return nil, fmt.Errorf("list domains failed: %w", err)
	}
	defer cursor.Close(timeoutCtx)

	// 解码结果
	var domains []models.XLDomain
	if err := cursor.All(timeoutCtx, &domains); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Msg("Decode domains failed")
		return nil, fmt.Errorf("decode domains failed: %w", err)
	}

	logger.RunLogger.Debug().
		Int("count", len(domains)).
		Msg("Domains listed successfully")
	return domains, nil
}

// CountDomains 计算域名总数
func (ds *DomainStore) CountDomains(ctx context.Context) (int64, error) {
	if ds.collection == nil {
		return 0, errors.New("MongoDB collection is nil")
	}

	// 创建超时上下文
	timeoutCtx, cancel := createTimeoutContext(ctx, ds.config.Timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().Msg("Counting domains")

	// 执行计数操作
	count, err := ds.collection.CountDocuments(timeoutCtx, bson.M{})
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Msg("Count domains failed")
		return 0, fmt.Errorf("count domains failed: %w", err)
	}

	logger.RunLogger.Debug().
		Int64("count", count).
		Msg("Domains counted successfully")
	return count, nil
}

// createTimeoutContext 创建带超时的上下文
func createTimeoutContext(ctx context.Context, timeoutSeconds int) (context.Context, context.CancelFunc) {
	timeout := DefaultConnectTimeout
	if timeoutSeconds > 0 {
		timeout = time.Duration(timeoutSeconds) * time.Second
	}
	return context.WithTimeout(ctx, timeout)
}
