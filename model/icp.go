package model

// ICPResponse 外部 ICP 查询服务响应
type ICPResponse struct {
	Code      int     `json:"code"`      // 状态码，0为成功
	Message   string  `json:"message"`   // 提示信息
	Data      ICPData `json:"data"`      // 备案数据
	RequestId string  `json:"requestId"` // 请求ID
}

// ICPData 外部 ICP 查询服务返回的备案数据
type ICPData struct {
	Domain         string `json:"domain"`         // 域名
	Company        string `json:"company"`        // 备案主体
	IcpNumber      string `json:"icpNumber"`      // 备案号
	NatureName     string `json:"natureName"`     // 主体性质（企业/个人）
	UpdateTime     string `json:"updateTime"`     // 备案更新时间
	HomeUrl        string `json:"homeUrl"`        // 网站首页
	ServiceContent string `json:"serviceContent"` // 服务内容
}

// ICPRecord 内部使用的 ICP 备案记录
type ICPRecord struct {
	Domain       string `json:"domain"`        // 域名
	ICPNumber    string `json:"icp_number"`    // 备案号
	Owner        string `json:"owner"`         // 备案主体
	Type         string `json:"type"`          // 主体性质（企业/个人）
	Status       string `json:"status"`        // 备案状态
	ApprovalDate string `json:"approval_date"` // 备案批准日期
	UpdateTime   string `json:"update_time"`   // 更新时间
	HomeURL      string `json:"home_url"`      // 网站首页
	Description  string `json:"description"`   // 网站描述
}

// ICPQueryRequest ICP 查询请求
type ICPQueryRequest struct {
	Domain string `json:"domain" binding:"required"` // 域名
}

// ICPQueryResponse ICP 查询响应
type ICPQueryResponse struct {
	Domain       string `json:"domain"`        // 域名
	ICPNumber    string `json:"icp_number"`    // 备案号
	Owner        string `json:"owner"`         // 备案主体
	Type         string `json:"type"`          // 主体性质（企业/个人）
	Status       string `json:"status"`        // 备案状态
	ApprovalDate string `json:"approval_date"` // 备案批准日期
}

// ConvertICPDataToRecord 将外部 ICP 数据转换为内部记录
func ConvertICPDataToRecord(data ICPData) ICPRecord {
	return ICPRecord{
		Domain:       data.Domain,
		ICPNumber:    data.IcpNumber,
		Owner:        data.Company,
		Type:         data.NatureName,
		Status:       "已备案", // 默认状态
		ApprovalDate: data.UpdateTime,
		UpdateTime:   data.UpdateTime,
		HomeURL:      data.HomeUrl,
		Description:  data.ServiceContent,
	}
}

// ConvertRecordToQueryResponse 将内部记录转换为查询响应
func ConvertRecordToQueryResponse(record ICPRecord) ICPQueryResponse {
	return ICPQueryResponse{
		Domain:       record.Domain,
		ICPNumber:    record.ICPNumber,
		Owner:        record.Owner,
		Type:         record.Type,
		Status:       record.Status,
		ApprovalDate: record.ApprovalDate,
	}
}
