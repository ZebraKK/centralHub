package workflow

import (
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"centralHub/model"
)

// 域名cname采用拼接方式： xxx+随机+.www
// dnspod subdomain长度限制50, www长度4, xxx长度自定义, 随机码长度[1,13]
// 随机码长度[1,13], .www 长度4
const dnspodMaxSubdomainLen = 50

func (wf *Workflow) makeCname(c *gin.Context, obj model.XLDomain) string {
	domainName := "" // placeholder, from obj
	uuid := uuid.New()
	var cnamePrefix string
	var start, end = 0, len(domainName)
	var subDomainLen = dnspodMaxSubdomainLen - 4 - len(uuid.String())
	if len(domainName) > subDomainLen {
		cnamePrefix = uuid.String()
	} else {
		if strings.HasPrefix(domainName, ".") { // 泛域名
			start = 1
		}
		cnamePrefix = domainName[start:end]
		// dnspod最低用户权限不支持subdomain的域名层次超过3层，这里普通域名设置成1层，泛域名设置为2层
		cnamePrefix = strings.Replace(cnamePrefix, ".", "-", -1)
		cnamePrefix = cnamePrefix + "-" + uuid.String()
	}
	if strings.HasPrefix(domainName, ".") { // 泛域名
		cnamePrefix = cnamePrefix + ".www"
	}

	cnameSuffix := ".xldns.com" // placeholder, to define

	return cnamePrefix + cnameSuffix
}

func (wf *Workflow) createVendorDomain(c *gin.Context, obj model.XLDomain) string {
	// 1, 确定要使用的vendor
	vendors := []string{"mock-vendor"}

	// 2, 调用vendor的接口创建域名
	var wg sync.WaitGroup
	for _, v := range vendors {
		wg.Add(1)
		go func(vendor string) {
			defer wg.Done()

			vendorClt := wf.getVendorClient(vendor)
			_ = vendorClt.CreateDomain(c, obj)
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
1, make Cname
2, create vendor domain
3,
*/
func (wf *Workflow) CreateDomain(c *gin.Context, obj model.XLDomain) string {

	cname := wf.makeCname(c, obj)

	_ = wf.createVendorDomain(c, obj)

	return cname
}
