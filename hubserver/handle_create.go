package hubserver

import (
	"centralHub/logger"
	"centralHub/model"
	"centralHub/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// validateCreateRequest 验证域名创建请求
func (hs *HubServer) validateCreateRequest(rlog *zerolog.Logger, reqObj *model.CreateDomainRequest) error {
	// 验证域名名称
	if reqObj.Name == "" {
		rlog.Error().Msg("Domain name is empty")
		return errors.New("domain name cannot be empty")
	}

	// 验证域名格式
	if !utils.IsValidDomainName(reqObj.Name) {
		rlog.Error().Str("domain", reqObj.Name).Msg("Invalid domain name format")
		return errors.New("invalid domain name format")
	}

	// 验证缓存配置
	if reqObj.Cache.GlobalCacheTime < -1 {
		rlog.Error().Int("cacheTime", reqObj.Cache.GlobalCacheTime).Msg("Invalid global cache time")
		return errors.New("global cache time must be >= -1")
	}

	// 验证回源配置
	if reqObj.Proxy.Source.Addr == "" {
		rlog.Error().Msg("Source address is empty")
		return errors.New("source address cannot be empty")
	}

	// 验证回源协议
	if reqObj.Proxy.SourceURLScheme != model.HTTPScheme &&
		reqObj.Proxy.SourceURLScheme != model.HTTPSScheme &&
		reqObj.Proxy.SourceURLScheme != model.FollowScheme {
		rlog.Error().Str("scheme", string(reqObj.Proxy.SourceURLScheme)).Msg("Invalid source URL scheme")
		return errors.New("invalid source URL scheme")
	}

	return nil
}

// preCreateCheck 域名创建前的检查
func (hs *HubServer) preCreateCheck(ctx context.Context, rlog *zerolog.Logger, reqObj *model.CreateDomainRequest) (*model.CDNDomain, *model.ICPInfo, error) {
	// 请求参数验证
	if err := hs.validateCreateRequest(rlog, reqObj); err != nil {
		return nil, nil, err
	}

	// 检查域名是否已存在
	name := reqObj.Name
	existingDomain, err := hs.domainStore.FindByName(ctx, name)
	if err == nil && existingDomain != nil {
		rlog.Warn().Str("domain", name).Msg("Domain already exists")
		return nil, nil, errors.New("domain already exists")
	}

	// 域名所有权检查
	owner, err := hs.getOwnership(name)
	if err != nil {
		rlog.Error().Err(err).Str("domain", name).Msg("Failed to check domain ownership")
		return nil, nil, fmt.Errorf("ownership check failed: %w", err)
	}

	if owner != "" && owner != reqObj.UserID {
		rlog.Warn().
			Str("domain", name).
			Str("owner", owner).
			Str("requester", reqObj.UserID).
			Msg("Domain is owned by another user")
		return nil, nil, errors.New("domain is owned by another user")
	}

	// ICP备案验证（带缓存）
	icpInfo, err := hs.verifyICPWithCache(ctx, rlog, name)
	if err != nil {
		rlog.Error().Err(err).Str("domain", name).Msg("ICP verification failed")
		return nil, nil, fmt.Errorf("ICP verification failed: %w", err)
	}

	// 记录ICP验证结果
	if icpInfo.Verified {
		rlog.Info().
			Str("domain", name).
			Str("icp_number", icpInfo.ICPNumber).
			Str("owner", icpInfo.Owner).
			Msg("Domain ICP verification passed")
	} else {
		rlog.Warn().
			Str("domain", name).
			Str("status", icpInfo.Status).
			Msg("Domain ICP verification failed or not required")
	}

	// 构建CDN域名配置
	cdnDomain := &model.CDNDomain{
		UID:   reqObj.UserID,
		Name:  name,
		Cache: reqObj.Cache,
		Proxy: reqObj.Proxy,
		ACL:   reqObj.ACL,
	}

	rlog.Debug().
		Str("domain", name).
		Str("user", reqObj.UserID).
		Msg("Domain pre-create check passed")

	return cdnDomain, icpInfo, nil
}

// HandleCreate 处理域名创建请求
func (hs *HubServer) HandleCreate(c *gin.Context) {
	// 获取请求ID并创建日志对象
	reqid, _ := c.Get("reqid")
	rlog := logger.WithReqID(reqid.(string))

	// 解析请求数据
	var reqObj model.CreateDomainRequest
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
		Str("user", reqObj.UserID).
		Str("client_ip", c.ClientIP()).
		Msg("Domain creation requested")

	// 域名创建前检查（包括ICP验证）
	ctx := c.Request.Context()
	toCreateDomain, icpInfo, err := hs.preCreateCheck(ctx, &rlog, &reqObj)
	if err != nil {
		rlog.Error().Err(err).Str("domain", reqObj.Name).Msg("Domain pre-create check failed")
		c.JSON(model.CodeBadRequest, model.NewErrorResponse(
			model.CodeBadRequest,
			fmt.Sprintf("Pre-create check failed: %s", err.Error()),
		))
		return
	}

	// 创建域名记录
	xlDomain := model.XLDomain{
		PlatformInfo: model.PlatformInfo{
			UserID:   reqObj.UserID,
			ICPInfo:  icpInfo, // 保存ICP信息
			CreateAt: time.Now().Unix(),
			UpdateAt: time.Now().Unix(),
		},
		DomainConfig: *toCreateDomain,
	}

	// 生成唯一ID
	xlDomain.ID = uuid.New().String()

	// 启动工作流
	taskId := hs.workflow.PushTask()

	// 异步执行域名创建工作流
	go func() {
		// 创建CNAME记录并再次检查ICP状态
		cname, icpRecord, err := hs.workflow.CreateDomain(ctx, &rlog, toCreateDomain)
		if err != nil {
			rlog.Error().
				Err(err).
				Str("domain", reqObj.Name).
				Str("task_id", taskId).
				Msg("Failed to create domain in workflow")
			return
		}

		// 更新域名记录中的CNAME
		xlDomain.PlatformInfo.Cname = cname

		// 如果工作流返回了ICP记录，更新域名的ICP信息
		if icpRecord != nil {
			xlDomain.PlatformInfo.ICPInfo = &model.ICPInfo{
				Verified:  true,
				ICPNumber: icpRecord.ICPNumber,
				Owner:     icpRecord.Owner,
				Status:    icpRecord.Status,
			}
		}

		// 保存域名记录到数据库
		if err := hs.domainStore.Insert(context.Background(), xlDomain); err != nil {
			rlog.Error().
				Err(err).
				Str("domain", reqObj.Name).
				Str("task_id", taskId).
				Msg("Failed to save domain record")
			return
		}

		rlog.Info().
			Str("domain", reqObj.Name).
			Str("cname", cname).
			Str("task_id", taskId).
			Msg("Domain created successfully")
	}()

	// 返回成功响应
	resp := model.CreateDomainResponse{
		Name:       reqObj.Name,
		JobID:      taskId,
		StatusCode: model.CodeSuccess,
		StatusMsg:  "Domain creation task submitted successfully",
	}

	c.JSON(model.CodeSuccess, model.NewSuccessResponse(resp))
}
