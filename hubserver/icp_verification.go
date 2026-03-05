package hubserver

import (
	"centralHub/model"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	// ICPCacheDuration ICP验证结果缓存时间（24小时）
	ICPCacheDuration = 24 * time.Hour
)

// verifyICPWithCache 验证域名ICP备案状态（带缓存）
func (hs *HubServer) verifyICPWithCache(ctx context.Context, rlog *zerolog.Logger, domain string) (*model.ICPInfo, error) {
	// 首先检查缓存
	icpInfo, err := hs.getICPFromCache(ctx, domain)
	if err == nil && icpInfo != nil {
		// 检查缓存是否过期
		cachedAt := time.Unix(icpInfo.CachedAt, 0)
		if time.Since(cachedAt) < ICPCacheDuration {
			rlog.Debug().
				Str("domain", domain).
				Str("icp_number", icpInfo.ICPNumber).
				Time("cached_at", cachedAt).
				Msg("Using cached ICP verification result")
			return icpInfo, nil
		}
		rlog.Debug().
			Str("domain", domain).
			Time("cached_at", cachedAt).
			Msg("ICP cache expired, re-verifying")
	}

	// 缓存不存在或已过期，查询ICP备案信息
	icpInfo, err = hs.queryICPInfo(ctx, rlog, domain)
	if err != nil {
		rlog.Warn().
			Err(err).
			Str("domain", domain).
			Msg("Failed to query ICP information")

		// 根据配置决定是否允许未备案域名创建
		if hs.config.Features.RequireICP {
			return nil, fmt.Errorf("ICP verification failed: %w", err)
		}

		// 返回未验证的ICP信息
		return &model.ICPInfo{
			Verified:   false,
			Status:     "未验证",
			VerifiedAt: time.Now().Unix(),
			CachedAt:   time.Now().Unix(),
		}, nil
	}

	return icpInfo, nil
}

// getICPFromCache 从缓存中获取ICP备案信息
func (hs *HubServer) getICPFromCache(ctx context.Context, domain string) (*model.ICPInfo, error) {
	// 查询域名记录
	xlDomain, err := hs.domainStore.FindByName(ctx, domain)
	if err != nil {
		return nil, err
	}

	// 返回ICP信息
	return xlDomain.PlatformInfo.ICPInfo, nil
}

// queryICPInfo 查询域名ICP备案信息
func (hs *HubServer) queryICPInfo(ctx context.Context, rlog *zerolog.Logger, domain string) (*model.ICPInfo, error) {
	// 提取主域名（如果是子域名）
	mainDomain := extractMainDomain(domain)

	rlog.Info().
		Str("domain", domain).
		Str("main_domain", mainDomain).
		Msg("Querying ICP information")

	// 调用ICP客户端查询
	record, err := hs.icpClient.QueryICP(ctx, mainDomain)
	if err != nil {
		return nil, fmt.Errorf("query ICP: %w", err)
	}

	// 检查是否已备案
	if record.ICPNumber == "" || record.Status != "已备案" {
		rlog.Warn().
			Str("domain", domain).
			Str("status", record.Status).
			Msg("Domain is not registered with ICP")

		return &model.ICPInfo{
			Verified:   false,
			ICPNumber:  record.ICPNumber,
			Status:     record.Status,
			Owner:      record.Owner,
			VerifiedAt: time.Now().Unix(),
			CachedAt:   time.Now().Unix(),
		}, nil
	}

	// 已备案，返回ICP信息
	rlog.Info().
		Str("domain", domain).
		Str("icp_number", record.ICPNumber).
		Str("owner", record.Owner).
		Msg("Domain is registered with ICP")

	return &model.ICPInfo{
		Verified:   true,
		ICPNumber:  record.ICPNumber,
		Status:     record.Status,
		Owner:      record.Owner,
		VerifiedAt: time.Now().Unix(),
		CachedAt:   time.Now().Unix(),
	}, nil
}

// extractMainDomain 提取主域名
func extractMainDomain(domain string) string {
	// 转换为小写
	domain = strings.ToLower(domain)

	// 去掉协议前缀
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	// 去掉端口
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	// 去掉路径
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// 分割域名
	parts := strings.Split(domain, ".")
	if len(parts) <= 2 {
		return domain
	}

	// 对于三级或更高级域名，返回后两段（如 www.example.com -> example.com）
	// 注意：这是简化逻辑，实际应该考虑公共后缀列表（如 .co.uk）
	return strings.Join(parts[len(parts)-2:], ".")
}
