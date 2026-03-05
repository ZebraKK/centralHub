package workflow

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"centralHub/model"
)

// CNAME 相关常量
const (
	// DNSPod subdomain长度限制
	dnspodMaxSubdomainLen = 50
	// CNAME 后缀
	cnameSuffix = ".xldns.com"
	// 最大重试次数
	maxRetries = 3
	// 特殊字符替换规则
	invalidCharsRegex = `[^a-zA-Z0-9\-]`
	// 最小随机字符串长度
	minRandomLength = 6
	// 最大随机字符串长度
	maxRandomLength = 12
)

// validateCname 验证CNAME格式
func validateCname(cname string) error {
	// 检查长度
	if len(cname) < 5 || len(cname) > 255 {
		return fmt.Errorf("invalid CNAME length: %d", len(cname))
	}

	// 检查格式
	if !strings.HasSuffix(cname, cnameSuffix) {
		return fmt.Errorf("invalid CNAME suffix: %s", cname)
	}

	// 检查字符
	matched, err := regexp.MatchString(`^[a-zA-Z0-9\-\.]+$`, cname)
	if err != nil || !matched {
		return fmt.Errorf("invalid CNAME characters: %s", cname)
	}

	return nil
}

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// makeCname 生成域名的CNAME记录
// 域名cname采用拼接方式： xxx+随机+.www
// 支持特殊域名处理，确保唯一性和格式正确
func (wf *Workflow) makeCname(rlog *zerolog.Logger, domainName string) string {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 生成随机字符串长度
	randomLength := rand.Intn(maxRandomLength-minRandomLength+1) + minRandomLength

	// 生成随机字符串
	randomStr := generateRandomString(randomLength)

	// 处理域名前缀
	var cnamePrefix string
	var start, end = 0, len(domainName)
	var subDomainLen = dnspodMaxSubdomainLen - 4 - randomLength

	// 检查域名长度
	if len(domainName) > subDomainLen {
		// 域名太长，使用随机字符串作为前缀
		cnamePrefix = randomStr
	} else {
		// 处理泛域名
		if strings.HasPrefix(domainName, ".") {
			start = 1
		}

		// 提取域名部分
		cnamePrefix = domainName[start:end]

		// 替换无效字符
		re := regexp.MustCompile(invalidCharsRegex)
		cnamePrefix = re.ReplaceAllString(cnamePrefix, "-")

		// 确保不以连字符开头或结尾
		cnamePrefix = strings.Trim(cnamePrefix, "-")

		// 添加随机字符串
		cnamePrefix = cnamePrefix + "-" + randomStr
	}

	// 处理泛域名
	if strings.HasPrefix(domainName, ".") {
		cnamePrefix = cnamePrefix + ".www"
	}

	// 组合完整CNAME
	cname := cnamePrefix + cnameSuffix

	// 验证CNAME格式
	if err := validateCname(cname); err != nil {
		rlog.Error().Err(err).Str("domain", domainName).Msg("Invalid CNAME generated")
		// 生成备用CNAME
		cname = "domain-" + randomStr + cnameSuffix
	}

	// 记录生成的CNAME
	rlog.Info().
		Str("domain", domainName).
		Str("cname", cname).
		Msg("Generated CNAME")

	return cname
}

// createVendorDomain 调用供应商接口创建域名
func (wf *Workflow) createVendorDomain(ctx context.Context, rlog *zerolog.Logger, obj *model.CDNDomain) string {
	rlog.Debug().Msg("Start create vendor domain")
	// 1, 确定要使用的vendor
	vendors := []string{"mock-vendor"}

	// 2, 调用vendor的接口创建域名
	var wg sync.WaitGroup
	for _, v := range vendors {
		wg.Add(1)
		go func(vendor string) {
			defer wg.Done()

			vendorClt := wf.getVendorClient(vendor)
			_ = vendorClt.CreateDomain(ctx, obj)
		}(v)
	}
	// 3, 返回vendor的域名
	wg.Wait()

	// placeholder
	// 三方对接, 是异步任务, 回调或者轮询
	return ""
}

/*
CreateDomain 创建域名的工作流
1. 检查ICP备案状态
2. 生成CNAME
3. 创建vendor域名
4. 创建DNS记录
5. 创建CDN配置
*/
func (wf *Workflow) CreateDomain(ctx context.Context, rlog *zerolog.Logger, obj *model.CDNDomain) (string, *model.ICPRecord, error) {
	// 步骤1：检查ICP备案状态
	rlog.Info().Str("domain", obj.Name).Msg("Checking ICP status")
	icpRecord, err := wf.CheckICPStatus(ctx, obj.Name)
	if err != nil {
		rlog.Error().Err(err).Str("domain", obj.Name).Msg("Failed to check ICP status")
		// ICP检查失败不阻止创建流程，只记录警告
		rlog.Warn().Str("domain", obj.Name).Msg("Proceeding without ICP verification")
		icpRecord = nil
	} else {
		rlog.Info().
			Str("domain", obj.Name).
			Str("icp_number", icpRecord.ICPNumber).
			Str("owner", icpRecord.Owner).
			Msg("ICP status checked successfully")
	}

	// 步骤2：生成CNAME
	rlog.Info().Str("domain", obj.Name).Msg("Generating CNAME")
	cname := wf.makeCname(rlog, obj.Name)

	// 步骤3：创建vendor域名（目前是占位符）
	rlog.Info().Str("domain", obj.Name).Msg("Creating vendor domain")
	_ = wf.createVendorDomain(ctx, rlog, obj)

	return cname, icpRecord, nil
}
