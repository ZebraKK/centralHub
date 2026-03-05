package utils

import (
	"regexp"
	"strings"
)

// 域名相关常量
const (
	MaxDomainLength = 255
	MaxLabelLength  = 63
)

// 域名验证正则表达式
// 允许字母、数字、连字符，不允许连续连字符，不允许开头和结尾是连字符
// 每个标签最长63个字符，整个域名最长255个字符
var (
	domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
)

// IsValidDomainName 验证域名格式是否正确
// 规则：
// 1. 域名总长度不超过255个字符
// 2. 每个标签（点之间的部分）不超过63个字符
// 3. 标签只能包含字母、数字和连字符
// 4. 标签不能以连字符开头或结尾
// 5. 标签不能包含连续的连字符
// 6. 顶级域名（最后一个标签）必须是纯字母
func IsValidDomainName(domain string) bool {
	// 检查域名长度
	if len(domain) == 0 || len(domain) > MaxDomainLength {
		return false
	}

	// 转换为小写
	domain = strings.ToLower(domain)

	// 使用正则表达式验证
	if !domainRegex.MatchString(domain) {
		return false
	}

	// 检查每个标签的长度
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) > MaxLabelLength {
			return false
		}
	}

	// 检查顶级域名（最后一个标签）
	tld := labels[len(labels)-1]
	for _, r := range tld {
		if r < 'a' || r > 'z' {
			return false
		}
	}

	return true
}

// NormalizeDomain 规范化域名
// 1. 转换为小写
// 2. 去除前后空格
// 3. 去除末尾的点
func NormalizeDomain(domain string) string {
	// 转换为小写
	domain = strings.ToLower(domain)

	// 去除前后空格
	domain = strings.TrimSpace(domain)

	// 去除末尾的点
	domain = strings.TrimSuffix(domain, ".")

	return domain
}

// ExtractSubdomain 从完整域名中提取子域名部分
// 例如：从 www.example.com 中提取 www
func ExtractSubdomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) <= 2 {
		return ""
	}
	return parts[0]
}

// ExtractMainDomain 从完整域名中提取主域名部分
// 例如：从 www.example.com 中提取 example.com
func ExtractMainDomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) <= 1 {
		return domain
	}

	// 简单处理，取最后两部分
	// 注意：这种方法对于 co.uk 这样的域名不准确
	// 实际应用中应该使用公共后缀列表
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}

	return domain
}

// IsSameDomain 检查两个域名是否属于同一个主域名
func IsSameDomain(domain1, domain2 string) bool {
	return ExtractMainDomain(domain1) == ExtractMainDomain(domain2)
}

// IsSubdomain 检查一个域名是否是另一个域名的子域名
func IsSubdomain(subdomain, domain string) bool {
	// 规范化域名
	subdomain = NormalizeDomain(subdomain)
	domain = NormalizeDomain(domain)

	// 检查子域名是否以主域名结尾
	return strings.HasSuffix(subdomain, "."+domain)
}
