package model

type ICPResponse struct {
	Code      int     `json:"code"`      // 状态码，0为成功
	Message   string  `json:"message"`   // 提示信息
	Data      ICPData `json:"data"`      // 备案数据
	RequestId string  `json:"requestId"` // 请求ID
}

type ICPData struct {
	Domain         string `json:"domain"`         // 域名
	Company        string `json:"company"`        // 备案主体
	IcpNumber      string `json:"icpNumber"`      // 备案号
	NatureName     string `json:"natureName"`     // 主体性质（企业/个人）
	UpdateTime     string `json:"updateTime"`     // 备案更新时间
	HomeUrl        string `json:"homeUrl"`        // 网站首页
	ServiceContent string `json:"serviceContent"` // 服务内容
}
