package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"centralHub/logger"
	"centralHub/model"
	"centralHub/utils"
)

// MongoDB 集合名称常量
const (
	OwnershipCollectionName = "ownership_verifications"
)

// 错误定义
var (
	ErrEmptyVerificationID        = errors.New("verification ID cannot be empty")
	ErrEmptyDomainForVerification = errors.New("domain name cannot be empty for verification")
	ErrInvalidVerificationType    = errors.New("invalid verification type")
)

// OwnershipStore 域名所有权验证存储结构体
type OwnershipStore struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
	timeout    int
}

// NewOwnershipStore 创建域名所有权验证存储实例
func NewOwnershipStore(client *mongo.Client, db *mongo.Database, timeout int) *OwnershipStore {
	collection := db.Collection(OwnershipCollectionName)

	logger.RunLogger.Info().
		Str("collection", OwnershipCollectionName).
		Msg("Ownership verification store initialized")

	return &OwnershipStore{
		client:     client,
		db:         db,
		collection: collection,
		timeout:    timeout,
	}
}

// validateOwnershipRecord 验证所有权验证记录
func validateOwnershipRecord(record model.OwnershipRecord) error {
	// 验证域名名称
	if record.Name == "" {
		return ErrEmptyDomainForVerification
	}

	// 验证域名格式
	if !utils.IsValidDomainName(record.Name) {
		return ErrInvalidDomainName
	}

	// 验证用户ID
	if record.UserID == "" {
		return ErrEmptyUserID
	}

	// 验证验证类型
	if record.VerifyType != model.DNSVerification && record.VerifyType != model.FileVerification {
		return ErrInvalidVerificationType
	}

	return nil
}

// InsertVerification 插入验证记录
func (os *OwnershipStore) InsertVerification(ctx context.Context, record model.OwnershipRecord) error {
	if os.collection == nil {
		return errors.New("MongoDB collection is nil")
	}

	// 验证记录
	if err := validateOwnershipRecord(record); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", record.Name).
			Str("type", string(record.VerifyType)).
			Msg("Ownership verification record validation failed")
		return fmt.Errorf("ownership verification record validation failed: %w", err)
	}

	// 创建超时上下文
	timeout := time.Duration(os.timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("domain", record.Name).
		Str("type", string(record.VerifyType)).
		Msg("Inserting ownership verification record")

	// 执行插入操作
	_, err := os.collection.InsertOne(timeoutCtx, record)
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", record.Name).
			Msg("Insert ownership verification record failed")
		return fmt.Errorf("insert ownership verification record failed: %w", err)
	}

	logger.RunLogger.Debug().
		Str("domain", record.Name).
		Str("id", record.ID).
		Msg("Ownership verification record inserted successfully")
	return nil
}

// FindVerificationByID 根据ID查找验证记录
func (os *OwnershipStore) FindVerificationByID(ctx context.Context, id string) (*model.OwnershipRecord, error) {
	if os.collection == nil {
		return nil, errors.New("MongoDB collection is nil")
	}

	if id == "" {
		return nil, ErrEmptyVerificationID
	}

	// 创建超时上下文
	timeout := time.Duration(os.timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("id", id).
		Msg("Finding ownership verification record by ID")

	// 执行查询操作
	var record model.OwnershipRecord
	err := os.collection.FindOne(timeoutCtx, bson.M{"_id": id}).Decode(&record)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			logger.RunLogger.Debug().
				Str("id", id).
				Msg("Ownership verification record not found")
			return nil, fmt.Errorf("ownership verification record not found: %w", err)
		}

		logger.RunLogger.Error().
			Err(err).
			Str("id", id).
			Msg("Find ownership verification record failed")
		return nil, fmt.Errorf("find ownership verification record failed: %w", err)
	}

	logger.RunLogger.Debug().
		Str("id", id).
		Msg("Ownership verification record found")
	return &record, nil
}

// FindVerificationByDomainAndType 根据域名和验证类型查找最新的验证记录
func (os *OwnershipStore) FindVerificationByDomainAndType(
	ctx context.Context,
	domain string,
	verifyType model.VerificationType,
) (*model.OwnershipRecord, error) {
	if os.collection == nil {
		return nil, errors.New("MongoDB collection is nil")
	}

	if domain == "" {
		return nil, ErrEmptyDomainForVerification
	}

	// 创建超时上下文
	timeout := time.Duration(os.timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("domain", domain).
		Str("type", string(verifyType)).
		Msg("Finding ownership verification record by domain and type")

	// 创建查询选项
	findOptions := options.FindOne().
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // 按创建时间降序，获取最新记录

	// 执行查询操作
	var record model.OwnershipRecord
	err := os.collection.FindOne(
		timeoutCtx,
		bson.M{"name": domain, "verify_type": verifyType},
		findOptions,
	).Decode(&record)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			logger.RunLogger.Debug().
				Str("domain", domain).
				Str("type", string(verifyType)).
				Msg("Ownership verification record not found")
			return nil, fmt.Errorf("ownership verification record not found: %w", err)
		}

		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain).
			Str("type", string(verifyType)).
			Msg("Find ownership verification record failed")
		return nil, fmt.Errorf("find ownership verification record failed: %w", err)
	}

	logger.RunLogger.Debug().
		Str("domain", domain).
		Str("type", string(verifyType)).
		Str("id", record.ID).
		Msg("Ownership verification record found")
	return &record, nil
}

// UpdateVerificationStatus 更新验证记录状态
func (os *OwnershipStore) UpdateVerificationStatus(
	ctx context.Context,
	id string,
	status model.VerificationStatus,
) error {
	if os.collection == nil {
		return errors.New("MongoDB collection is nil")
	}

	if id == "" {
		return ErrEmptyVerificationID
	}

	// 创建超时上下文
	timeout := time.Duration(os.timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("id", id).
		Str("status", string(status)).
		Msg("Updating ownership verification status")

	// 执行更新操作
	update := bson.M{
		"status":     status,
		"updated_at": time.Now().Unix(),
	}

	result, err := os.collection.UpdateOne(
		timeoutCtx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)

	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("id", id).
			Msg("Update ownership verification status failed")
		return fmt.Errorf("update ownership verification status failed: %w", err)
	}

	if result.MatchedCount == 0 {
		logger.RunLogger.Warn().
			Str("id", id).
			Msg("No ownership verification record matched for update")
		return fmt.Errorf("ownership verification record not found for update")
	}

	logger.RunLogger.Debug().
		Str("id", id).
		Str("status", string(status)).
		Int64("matched", result.MatchedCount).
		Int64("modified", result.ModifiedCount).
		Msg("Ownership verification status updated successfully")
	return nil
}

// ListVerificationsByDomain 列出域名的所有验证记录
func (os *OwnershipStore) ListVerificationsByDomain(
	ctx context.Context,
	domain string,
	skip,
	limit int64,
) ([]model.OwnershipRecord, error) {
	if os.collection == nil {
		return nil, errors.New("MongoDB collection is nil")
	}

	if domain == "" {
		return nil, ErrEmptyDomainForVerification
	}

	// 创建超时上下文
	timeout := time.Duration(os.timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("domain", domain).
		Int64("skip", skip).
		Int64("limit", limit).
		Msg("Listing ownership verification records by domain")

	// 创建查询选项
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // 按创建时间降序

	// 执行查询操作
	cursor, err := os.collection.Find(
		timeoutCtx,
		bson.M{"name": domain},
		findOptions,
	)

	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain).
			Msg("List ownership verification records failed")
		return nil, fmt.Errorf("list ownership verification records failed: %w", err)
	}
	defer cursor.Close(timeoutCtx)

	// 解码结果
	var records []model.OwnershipRecord
	if err := cursor.All(timeoutCtx, &records); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain).
			Msg("Decode ownership verification records failed")
		return nil, fmt.Errorf("decode ownership verification records failed: %w", err)
	}

	logger.RunLogger.Debug().
		Str("domain", domain).
		Int("count", len(records)).
		Msg("Ownership verification records listed successfully")
	return records, nil
}

// CleanupExpiredVerifications 清理过期的验证记录
func (os *OwnershipStore) CleanupExpiredVerifications(ctx context.Context) (int64, error) {
	if os.collection == nil {
		return 0, errors.New("MongoDB collection is nil")
	}

	// 创建超时上下文
	timeout := time.Duration(os.timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 当前时间戳
	now := time.Now().Unix()

	// 记录操作日志
	logger.RunLogger.Info().
		Int64("now", now).
		Msg("Cleaning up expired ownership verification records")

	// 执行删除操作
	result, err := os.collection.DeleteMany(
		timeoutCtx,
		bson.M{
			"expire_at": bson.M{"$lt": now},
			"status": bson.M{"$in": []model.VerificationStatus{
				model.StatusPending,
				model.StatusExpired,
			}},
		},
	)

	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Msg("Cleanup expired ownership verification records failed")
		return 0, fmt.Errorf("cleanup expired ownership verification records failed: %w", err)
	}

	logger.RunLogger.Debug().
		Int64("deleted", result.DeletedCount).
		Msg("Expired ownership verification records cleaned up successfully")
	return result.DeletedCount, nil
}

// ScheduleCleanup 定期清理过期的验证记录
func (os *OwnershipStore) ScheduleCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				// 执行清理操作
				count, err := os.CleanupExpiredVerifications(ctx)
				if err != nil {
					logger.RunLogger.Error().
						Err(err).
						Msg("Scheduled cleanup of expired verification records failed")
				} else {
					logger.RunLogger.Info().
						Int64("deleted", count).
						Msg("Scheduled cleanup of expired verification records completed")
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	logger.RunLogger.Info().
		Dur("interval", interval).
		Msg("Scheduled cleanup of expired verification records started")
}

// UpdateExpiredVerifications 更新过期的验证记录状态
func (os *OwnershipStore) UpdateExpiredVerifications(ctx context.Context) (int64, error) {
	if os.collection == nil {
		return 0, errors.New("MongoDB collection is nil")
	}

	// 创建超时上下文
	timeout := time.Duration(os.timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 当前时间戳
	now := time.Now().Unix()

	// 记录操作日志
	logger.RunLogger.Info().
		Int64("now", now).
		Msg("Updating expired ownership verification records")

	// 执行更新操作
	result, err := os.collection.UpdateMany(
		timeoutCtx,
		bson.M{
			"expire_at": bson.M{"$lt": now},
			"status":    model.StatusPending,
		},
		bson.M{
			"$set": bson.M{
				"status":     model.StatusExpired,
				"updated_at": now,
			},
		},
	)

	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Msg("Update expired ownership verification records failed")
		return 0, fmt.Errorf("update expired ownership verification records failed: %w", err)
	}

	logger.RunLogger.Debug().
		Int64("modified", result.ModifiedCount).
		Msg("Expired ownership verification records updated successfully")
	return result.ModifiedCount, nil
}

// GetVerificationStatistics 获取验证记录统计信息
func (os *OwnershipStore) GetVerificationStatistics(ctx context.Context) (map[string]int64, error) {
	if os.collection == nil {
		return nil, errors.New("MongoDB collection is nil")
	}

	// 创建超时上下文
	timeout := time.Duration(os.timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Msg("Getting ownership verification statistics")

	// 执行聚合操作
	pipeline := mongo.Pipeline{
		{
			{"$group", bson.D{
				{"_id", "$status"},
				{"count", bson.D{{"$sum", 1}}},
			}},
		},
	}

	cursor, err := os.collection.Aggregate(timeoutCtx, pipeline)
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Msg("Get ownership verification statistics failed")
		return nil, fmt.Errorf("get ownership verification statistics failed: %w", err)
	}
	defer cursor.Close(timeoutCtx)

	// 解析结果
	var results []struct {
		ID    model.VerificationStatus `bson:"_id"`
		Count int64                    `bson:"count"`
	}
	if err := cursor.All(timeoutCtx, &results); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Msg("Decode ownership verification statistics failed")
		return nil, fmt.Errorf("decode ownership verification statistics failed: %w", err)
	}

	// 构建统计结果
	stats := make(map[string]int64)
	for _, result := range results {
		stats[string(result.ID)] = result.Count
	}

	logger.RunLogger.Debug().
		Interface("stats", stats).
		Msg("Ownership verification statistics retrieved successfully")
	return stats, nil
}

// GetLatestVerificationByDomain 获取域名最新的验证记录
func (os *OwnershipStore) GetLatestVerificationByDomain(
	ctx context.Context,
	domain string,
) (*model.OwnershipRecord, error) {
	if os.collection == nil {
		return nil, errors.New("MongoDB collection is nil")
	}

	if domain == "" {
		return nil, ErrEmptyDomainForVerification
	}

	// 创建超时上下文
	timeout := time.Duration(os.timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Getting latest ownership verification record by domain")

	// 创建查询选项
	findOptions := options.FindOne().
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // 按创建时间降序，获取最新记录

	// 执行查询操作
	var record model.OwnershipRecord
	err := os.collection.FindOne(
		timeoutCtx,
		bson.M{"name": domain},
		findOptions,
	).Decode(&record)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			logger.RunLogger.Debug().
				Str("domain", domain).
				Msg("No ownership verification record found")
			return nil, fmt.Errorf("no ownership verification record found: %w", err)
		}

		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain).
			Msg("Get latest ownership verification record failed")
		return nil, fmt.Errorf("get latest ownership verification record failed: %w", err)
	}

	logger.RunLogger.Debug().
		Str("domain", domain).
		Str("id", record.ID).
		Str("status", string(record.Status)).
		Msg("Latest ownership verification record found")
	return &record, nil
}

// GetVerificationsByStatus 获取指定状态的验证记录
func (os *OwnershipStore) GetVerificationsByStatus(
	ctx context.Context,
	status model.VerificationStatus,
	skip, limit int64,
) ([]model.OwnershipRecord, error) {
	if os.collection == nil {
		return nil, errors.New("MongoDB collection is nil")
	}

	// 创建超时上下文
	timeout := time.Duration(os.timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("status", string(status)).
		Int64("skip", skip).
		Int64("limit", limit).
		Msg("Getting ownership verification records by status")

	// 创建查询选项
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // 按创建时间降序

	// 执行查询操作
	cursor, err := os.collection.Find(
		timeoutCtx,
		bson.M{"status": status},
		findOptions,
	)

	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("status", string(status)).
			Msg("Get ownership verification records by status failed")
		return nil, fmt.Errorf("get ownership verification records by status failed: %w", err)
	}
	defer cursor.Close(timeoutCtx)

	// 解码结果
	var records []model.OwnershipRecord
	if err := cursor.All(timeoutCtx, &records); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("status", string(status)).
			Msg("Decode ownership verification records failed")
		return nil, fmt.Errorf("decode ownership verification records failed: %w", err)
	}

	logger.RunLogger.Debug().
		Str("status", string(status)).
		Int("count", len(records)).
		Msg("Ownership verification records retrieved successfully")
	return records, nil
}

// CountVerificationsByDomain 计算域名的验证记录总数
func (os *OwnershipStore) CountVerificationsByDomain(ctx context.Context, domain string) (int64, error) {
	if os.collection == nil {
		return 0, errors.New("MongoDB collection is nil")
	}

	if domain == "" {
		return 0, ErrEmptyDomainForVerification
	}

	// 创建超时上下文
	timeout := time.Duration(os.timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录操作日志
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Counting ownership verification records by domain")

	// 执行计数操作
	count, err := os.collection.CountDocuments(timeoutCtx, bson.M{"name": domain})
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain).
			Msg("Count ownership verification records failed")
		return 0, fmt.Errorf("count ownership verification records failed: %w", err)
	}

	logger.RunLogger.Debug().
		Str("domain", domain).
		Int64("count", count).
		Msg("Ownership verification records counted successfully")
	return count, nil
}
