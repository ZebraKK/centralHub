package hubserver

import "github.com/gin-gonic/gin"

/*
	支持用户域名所有权检查的交互接口(ownership)
	dns txt 记录
		cname 记录
	节点服务器file upload 验证
*/

func (hs *HubServer) HandleOwnershipCheck(c *gin.Context) {

	//
	// 用户提交域名，验证形式，发起验证
	// 响应对应验证类型的 值 和此次验证请求的任务ID

	type ReqObj struct {
		Domain     string `form:"domain" binding:"required"`
		VerifyType string `form:"verify_type" binding:"required"` // dns | file
	}
	var reqObj ReqObj
	//parse form data
	if err := c.ShouldBind(&reqObj); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var value string
	switch reqObj.VerifyType {
	case "dns":
		value = hs.makeTXTStr(reqObj.Domain)
	case "file":
		_, value = hs.makeFile(reqObj.Domain)
	}

	type RespObj struct {
		Domain     string `json:"domain"`
		VerifyType string `json:"verify_type"`
		Value      string `json:"value"`
		ReqID      string `json:"req_id"`
	}

	respObj := RespObj{
		Domain:     reqObj.Domain,
		VerifyType: reqObj.VerifyType,
		Value:      value,
		ReqID:      "example_req_id", // same as workflow task id
	}

	// db save reqID , domain, verifyType, value, status(pending)
	//hs.db.Save(&RespObj)

	c.JSON(200, respObj)

}

func (hs *HubServer) HandleOwnershipVerify(c *gin.Context) {
	// 提交域名，验证请求ID
	// 检查对应验证结果 或者当前进度

	type ReqObj struct {
		Domain string `form:"domain" binding:"required"`
		ReqID  string `form:"req_id" binding:"required"`
	}
	var reqObj ReqObj
	//parse form data
	if err := c.ShouldBind(&reqObj); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	verifyType := "dns" // from db reqID

	finish := false
	switch verifyType {
	case "dns":
		finish = hs.checkDNSRecords(reqObj.Domain)
	case "file":
		finish = hs.checkFileUpload(reqObj.Domain)
	}

	type RespObj struct {
		Domain string `json:"domain"`
		Status string `json:"status"` // pending | verified | failed
		ReqID  string `json:"req_id"`
	}

	respObj := RespObj{
		Domain: reqObj.Domain,
		Status: "pending",
		ReqID:  reqObj.ReqID,
	}
	if finish {
		respObj.Status = "verified"
		// update db

	} else {
		respObj.Status = "pending"
	}
	c.JSON(200, respObj)

}

func (hs *HubServer) makeTXTStr(domain string) string {
	// 生成 TXT 记录验证字符串
	return domain + "example TXT record" // hash(domain, ts, etc...)
}

func (hs *HubServer) makeFile(domain string) (name, value string) {
	// 生成 CNAME 记录验证字符串
	return "example.com", domain + "cname_value"
}

func (hs *HubServer) checkDNSRecords(domain string) bool {
	// 检查 DNS 记录
	return true
}

func (hs *HubServer) checkFileUpload(domain string) bool {
	// 检查文件上传
	return true
}
