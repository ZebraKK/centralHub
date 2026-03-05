package hubserver

import (
	"centralHub/logger"
	"centralHub/model"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// QueryParams 查询参数结构体
type QueryParams struct {
	Name     string // 域名名称
	UserID   string // 用户ID
	Page     int    // 页码
	PageSize int    // 每页大小
}

// validateQueryParams 验证查询参数
func validateQueryParams(params *QueryParams) error {
	// 验证分页参数
	if params.Page < 0 {
		return errors.New("page must be >= 0")
	}
	if params.PageSize <= 0 {
		return errors.New("page_size must be > 0")
	}
	if params.PageSize > 100 {
		return errors.New("page_size must be <= 100")
	}

	return nil
}

// parseQueryParams 解析查询参数
func parseQueryParams(c *gin.Context) (*QueryParams, error) {
	params := &QueryParams{
		Name:     c.Query("name"),
		UserID:   c.Query("user_id"),
		Page:     0,
		PageSize: 10, // 默认每页10条
	}

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, fmt.Errorf("invalid page parameter: %w", err)
		}
		params.Page = page
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid page_size parameter: %w", err)
		}
		params.PageSize = pageSize
	}

	// 验证参数
	if err := validateQueryParams(params); err != nil {
		return nil, err
	}

	return params, nil
}

// HandleQuery 处理域名查询请求
func (hs *HubServer) HandleQuery(c *gin.Context) {
	// 获取请求ID并创建日志对象
	reqid, _ := c.Get("reqid")
	rlog := logger.WithReqID(reqid.(string))

	// 解析查询参数
	params, err := parseQueryParams(c)
	if err != nil {
		rlog.Error().Err(err).Msg("Failed to parse query parameters")
		c.JSON(model.CodeBadRequest, model.NewErrorResponse(
			model.CodeBadRequest,
			fmt.Sprintf("Invalid query parameters: %s", err.Error()),
		))
		return
	}

	// 记录审计日志
	rlog.Info().
		Str("name", params.Name).
		Str("user_id", params.UserID).
		Int("page", params.Page).
		Int("page_size", params.PageSize).
		Str("client_ip", c.ClientIP()).
		Msg("Domain query requested")

	// 创建查询上下文
	ctx := c.Request.Context()

	// 查询域名总数
	var total int64
	var err1 error
	if params.Name != "" {
		// 如果指定了域名名称，则查询特定域名
		domain, err := hs.queryDomainByName(ctx, params.Name, params.UserID)
		if err != nil {
			rlog.Error().Err(err).Str("name", params.Name).Msg("Failed to query domain")
			c.JSON(model.CodeServerError, model.NewErrorResponse(
				model.CodeServerError,
				fmt.Sprintf("Failed to query domain: %s", err.Error()),
			))
			return
		}

		if domain == nil {
			// 域名不存在
			c.JSON(model.CodeSuccess, model.NewPageResponse(
				[]model.DomainResponse{},
				0,
				params.Page,
				params.PageSize,
			))
			return
		}

		// 转换为响应格式
		resp := convertToDomainResponse(domain)
		c.JSON(model.CodeSuccess, model.NewPageResponse(
			[]model.DomainResponse{*resp},
			1,
			params.Page,
			params.PageSize,
		))
		return
	}

	// 查询域名总数
	total, err1 = hs.domainStore.CountDomains(ctx)
	if err1 != nil {
		rlog.Error().Err(err1).Msg("Failed to count domains")
		c.JSON(model.CodeServerError, model.NewErrorResponse(
			model.CodeServerError,
			"Failed to count domains",
		))
		return
	}

	// 如果没有域名，直接返回空列表
	if total == 0 {
		c.JSON(model.CodeSuccess, model.NewPageResponse(
			[]model.DomainResponse{},
			0,
			params.Page,
			params.PageSize,
		))
		return
	}

	// 计算分页参数
	skip := int64(params.Page * params.PageSize)
	limit := int64(params.PageSize)

	// 查询域名列表
	domains, err := hs.domainStore.ListDomains(ctx, skip, limit)
	if err != nil {
		rlog.Error().Err(err).Msg("Failed to list domains")
		c.JSON(model.CodeServerError, model.NewErrorResponse(
			model.CodeServerError,
			"Failed to list domains",
		))
		return
	}

	// 转换为响应格式
	respList := make([]model.DomainResponse, 0, len(domains))
	for _, domain := range domains {
		resp := convertToDomainResponse(&domain)
		respList = append(respList, *resp)
	}

	// 返回分页响应
	c.JSON(model.CodeSuccess, model.NewPageResponse(
		respList,
		total,
		params.Page,
		params.PageSize,
	))
}

// queryDomainByName 根据域名名称查询域名
func (hs *HubServer) queryDomainByName(ctx context.Context, name, userID string) (*model.XLDomain, error) {
	// 查询域名
	domain, err := hs.domainStore.FindByName(ctx, name)
	if err != nil {
		// 如果是找不到域名的错误，返回nil
		if errors.Is(err, errors.New("domain not found")) {
			return nil, nil
		}
		return nil, err
	}

	// 如果指定了用户ID，则验证域名所有权
	if userID != "" && domain.PlatformInfo.UserID != userID {
		return nil, nil // 不返回不属于该用户的域名
	}

	return domain, nil
}

// convertToDomainResponse 将XLDomain转换为DomainResponse
func convertToDomainResponse(domain *model.XLDomain) *model.DomainResponse {
	// 获取域名状态
	status := "active"
	if domain.DomainConfig.Status.Disabled {
		status = "disabled"
	}

	// 创建响应
	return &model.DomainResponse{
		ID:         domain.ID,
		DomainName: domain.DomainConfig.Name,
		Status:     status,
		Owner:      domain.PlatformInfo.UserID,
		ICPInfo:    domain.PlatformInfo.ICPInfo, // 添加ICP信息
		CreatedAt:  domain.PlatformInfo.CreateAt,
		UpdatedAt:  domain.PlatformInfo.UpdateAt,
	}
}
