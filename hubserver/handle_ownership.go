package hubserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"centralHub/logger"
	"centralHub/model"
	"centralHub/utils"
)

/*
	支持用户域名所有权检查的交互接口(ownership)
	dns txt 记录验证
	文件上传验证
*/

// HandleOwnershipCheck 处理域名所有权验证请求
func (hs *HubServer) HandleOwnershipCheck(c *gin.Context) {
	// 获取请求ID并创建日志对象
	reqid, _ := c.Get("reqid")
	rlog := logger.WithReqID(reqid.(string))

	// 解析请求数据
	var reqObj model.OwnershipVerificationRequest
	if err := c.ShouldBindJSON(&reqObj); err != nil {
		rlog.Error().Err(err).Msg("Failed to bind request data")
		c.JSON(model.CodeBadRequest, model.NewErrorResponse(
			model.CodeBadRequest,
			fmt.Sprintf("Invalid request format: %s", err.Error()),
		))
		return
	}

	// 设置用户ID（实际应从认证中间件获取）
	if reqObj.UserID == "" {
		reqObj.UserID = c.GetString("user_id")
		if reqObj.UserID == "" {
			reqObj.UserID = "anonymous" // 默认值，实际应该要求认证
		}
	}

	// 记录审计日志
	rlog.Info().
		Str("domain", reqObj.Name).
		Str("verify_type", string(reqObj.VerifyType)).
		Str("user_id", reqObj.UserID).
		Str("client_ip", c.ClientIP()).
		Msg("Domain ownership verification requested")

	// 验证域名格式
	if !utils.IsValidDomainName(reqObj.Name) {
		rlog.Error().Str("domain", reqObj.Name).Msg("Invalid domain name format")
		c.JSON(model.CodeBadRequest, model.NewErrorResponse(
			model.CodeBadRequest,
			"Invalid domain name format",
		))
		return
	}

	// 验证验证类型
	if reqObj.VerifyType != model.DNSVerification && reqObj.VerifyType != model.FileVerification {
		rlog.Error().Str("type", string(reqObj.VerifyType)).Msg("Invalid verification type")
		c.JSON(model.CodeBadRequest, model.NewErrorResponse(
			model.CodeBadRequest,
			"Invalid verification type, must be 'dns' or 'file'",
		))
		return
	}

	// 生成验证值
	var value string
	switch reqObj.VerifyType {
	case model.DNSVerification:
		value = hs.generateTXTVerificationValue(reqObj.Name, reqObj.UserID)
	case model.FileVerification:
		value = hs.generateFileVerificationValue(reqObj.Name, reqObj.UserID)
	}

	// 生成请求ID
	requestID := uuid.New().String()

	// 计算过期时间
	expireAt := time.Now().Add(24 * time.Hour).Unix()

	// 创建验证记录
	record := model.OwnershipRecord{
		ID:         requestID,
		Name:       reqObj.Name,
		UserID:     reqObj.UserID,
		VerifyType: reqObj.VerifyType,
		Value:      value,
		Status:     model.StatusPending,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
		ExpireAt:   expireAt,
	}

	// 保存验证记录
	ctx := c.Request.Context()
	if err := hs.ownershipStore.InsertVerification(ctx, record); err != nil {
		rlog.Error().Err(err).Msg("Failed to save verification record")
		c.JSON(model.CodeServerError, model.NewErrorResponse(
			model.CodeServerError,
			"Failed to save verification record",
		))
		return
	}

	// 构建响应
	resp := model.OwnershipVerificationResponse{
		Name:       reqObj.Name,
		VerifyType: reqObj.VerifyType,
		Value:      value,
		ReqID:      requestID,
		ExpireAt:   expireAt,
	}

	// 返回响应
	c.JSON(model.CodeSuccess, model.NewSuccessResponse(resp))
}

// HandleOwnershipStats 处理域名所有权验证统计信息查询
func (hs *HubServer) HandleOwnershipStats(c *gin.Context) {
	// 获取请求ID并创建日志对象
	reqid, _ := c.Get("reqid")
	rlog := logger.WithReqID(reqid.(string))

	// 记录审计日志
	rlog.Info().
		Str("client_ip", c.ClientIP()).
		Msg("Domain ownership verification statistics requested")

	// 获取统计信息
	ctx := c.Request.Context()
	stats, err := hs.ownershipStore.GetVerificationStatistics(ctx)
	if err != nil {
		rlog.Error().Err(err).Msg("Failed to get verification statistics")
		c.JSON(model.CodeServerError, model.NewErrorResponse(
			model.CodeServerError,
			"Failed to get verification statistics",
		))
		return
	}

	// 构建响应
	resp := model.OwnershipStatsResponse{
		Total:    0,
		Pending:  stats[string(model.StatusPending)],
		Verified: stats[string(model.StatusVerified)],
		Failed:   stats[string(model.StatusFailed)],
		Expired:  stats[string(model.StatusExpired)],
	}

	// 计算总数
	for _, count := range stats {
		resp.Total += count
	}

	// 返回响应
	c.JSON(model.CodeSuccess, model.NewSuccessResponse(resp))
}

// HandleOwnershipVerify 处理域名所有权验证结果查询
func (hs *HubServer) HandleOwnershipVerify(c *gin.Context) {
	// 获取请求ID并创建日志对象
	reqid, _ := c.Get("reqid")
	rlog := logger.WithReqID(reqid.(string))

	// 解析请求数据
	var reqObj model.OwnershipVerifyRequest
	if err := c.ShouldBindJSON(&reqObj); err != nil {
		rlog.Error().Err(err).Msg("Failed to bind request data")
		c.JSON(model.CodeBadRequest, model.NewErrorResponse(
			model.CodeBadRequest,
			fmt.Sprintf("Invalid request format: %s", err.Error()),
		))
		return
	}

	// 记录审计日志
	rlog.Info().
		Str("domain", reqObj.Name).
		Str("req_id", reqObj.ReqID).
		Str("client_ip", c.ClientIP()).
		Msg("Domain ownership verification check requested")

	// 查询验证记录
	ctx := c.Request.Context()
	record, err := hs.ownershipStore.FindVerificationByID(ctx, reqObj.ReqID)
	if err != nil {
		rlog.Error().Err(err).Str("req_id", reqObj.ReqID).Msg("Failed to find verification record")
		c.JSON(model.CodeNotFound, model.NewErrorResponse(
			model.CodeNotFound,
			"Verification record not found",
		))
		return
	}

	// 验证域名是否匹配
	if record.Name != reqObj.Name {
		rlog.Error().
			Str("request_domain", reqObj.Name).
			Str("record_domain", record.Name).
			Msg("Domain name mismatch")
		c.JSON(model.CodeBadRequest, model.NewErrorResponse(
			model.CodeBadRequest,
			"Domain name does not match verification record",
		))
		return
	}

	// 检查记录是否过期
	now := time.Now().Unix()
	if record.ExpireAt < now {
		// 更新记录状态为过期
		if err := hs.ownershipStore.UpdateVerificationStatus(ctx, record.ID, model.StatusExpired); err != nil {
			rlog.Error().Err(err).Str("req_id", reqObj.ReqID).Msg("Failed to update verification status")
		}

		rlog.Info().Str("req_id", reqObj.ReqID).Msg("Verification record expired")
		c.JSON(model.CodeBadRequest, model.NewErrorResponse(
			model.CodeBadRequest,
			"Verification record expired",
		))
		return
	}

	// 如果记录状态不是待验证，直接返回当前状态
	if record.Status != model.StatusPending {
		resp := model.OwnershipVerifyResponse{
			Name:   record.Name,
			Status: record.Status,
			ReqID:  record.ID,
			UserID: record.UserID,
		}
		c.JSON(model.CodeSuccess, model.NewSuccessResponse(resp))
		return
	}

	// 执行验证
	verified := false
	switch record.VerifyType {
	case model.DNSVerification:
		verified = hs.verifyDNSTXTRecord(ctx, &rlog, record.Name, record.Value)
	case model.FileVerification:
		verified = hs.verifyFileUpload(ctx, &rlog, record.Name, record.Value)
	}

	// 更新验证状态
	newStatus := model.StatusPending
	if verified {
		newStatus = model.StatusVerified
	}

	// 只有状态变更时才更新数据库
	if newStatus != record.Status {
		if err := hs.ownershipStore.UpdateVerificationStatus(ctx, record.ID, newStatus); err != nil {
			rlog.Error().Err(err).Str("req_id", reqObj.ReqID).Msg("Failed to update verification status")
			c.JSON(model.CodeServerError, model.NewErrorResponse(
				model.CodeServerError,
				"Failed to update verification status",
			))
			return
		}
	}

	// 构建响应
	resp := model.OwnershipVerifyResponse{
		Name:   record.Name,
		Status: newStatus,
		ReqID:  record.ID,
		UserID: record.UserID,
	}

	// 返回响应
	c.JSON(model.CodeSuccess, model.NewSuccessResponse(resp))
}

// generateTXTVerificationValue 生成TXT记录验证值
func (hs *HubServer) generateTXTVerificationValue(domain, userID string) string {
	// 生成验证字符串：centralHub-verify={hash(domain + userID + timestamp + salt)}
	timestamp := time.Now().Unix()
	salt := hs.config.Server.Mode // 使用服务器模式作为盐值
	data := fmt.Sprintf("%s:%s:%d:%s", domain, userID, timestamp, salt)

	// 计算SHA-256哈希
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	// 返回格式化的验证值
	return fmt.Sprintf("centralHub-verify=%s", hashStr)
}

// generateFileVerificationValue 生成文件验证值
func (hs *HubServer) generateFileVerificationValue(domain, userID string) string {
	// 生成验证字符串：类似TXT记录，但格式不同
	timestamp := time.Now().Unix()
	salt := hs.config.Server.Mode
	data := fmt.Sprintf("%s:%s:%d:%s:file", domain, userID, timestamp, salt)

	// 计算SHA-256哈希
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	// 返回格式化的验证值（包含更多信息，便于验证）
	return fmt.Sprintf("centralHub-file-verify=%s\ndomain=%s\ntimestamp=%d",
		hashStr, domain, timestamp)
}

// verifyDNSTXTRecord 验证DNS TXT记录
func (hs *HubServer) verifyDNSTXTRecord(ctx context.Context, rlog *zerolog.Logger, domain, expectedValue string) bool {
	if hs.dnsClient == nil {
		rlog.Error().Msg("DNS client is nil")
		return false
	}

	// 记录验证开始
	rlog.Info().
		Str("domain", domain).
		Msg("Verifying DNS TXT record")

	// 验证TXT记录
	verified, err := hs.dnsClient.VerifyTXTRecord(domain, expectedValue)
	if err != nil {
		rlog.Error().
			Err(err).
			Str("domain", domain).
			Msg("DNS TXT record verification failed")
		return false
	}

	if verified {
		rlog.Info().
			Str("domain", domain).
			Msg("DNS TXT record verified successfully")
	} else {
		rlog.Info().
			Str("domain", domain).
			Msg("DNS TXT record verification pending")
	}

	return verified
}

// verifyFileUpload 验证文件上传
func (hs *HubServer) verifyFileUpload(ctx context.Context, rlog *zerolog.Logger, domain, expectedValue string) bool {
	// 记录验证开始
	rlog.Info().
		Str("domain", domain).
		Msg("Verifying file upload")

	// 构建验证文件URL
	verifyURL := fmt.Sprintf("http://%s/.well-known/centralHub-verify.txt", domain)

	// 使用HTTP客户端获取文件内容并验证
	if hs.httpClient == nil {
		rlog.Error().Msg("HTTP client is nil")
		return false
	}

	verified, err := hs.httpClient.VerifyFileContent(ctx, verifyURL, expectedValue)
	if err != nil {
		rlog.Error().
			Err(err).
			Str("domain", domain).
			Str("url", verifyURL).
			Msg("File upload verification failed")
		return false
	}

	if verified {
		rlog.Info().
			Str("domain", domain).
			Str("url", verifyURL).
			Msg("File upload verification succeeded")
	} else {
		rlog.Info().
			Str("domain", domain).
			Str("url", verifyURL).
			Msg("File upload verification failed: content does not match")
	}

	return verified
}
