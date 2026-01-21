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
		Name       string `form:"name" binding:"required"`
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
		value = hs.makeTXTStr(reqObj.Name)
	case "file":
		_, value = hs.makeFile(reqObj.Name)
	}

	type RespObj struct {
		Name       string `json:"name"`
		VerifyType string `json:"verify_type"`
		Value      string `json:"value"`
		ReqID      string `json:"req_id"`
	}

	respObj := RespObj{
		Name:       reqObj.Name,
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
		Name  string `form:"name" binding:"required"`
		ReqID string `form:"req_id" binding:"required"`
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
		finish = hs.checkDNSRecords(reqObj.Name)
	case "file":
		finish = hs.checkFileUpload(reqObj.Name)
	}

	type RespObj struct {
		Name   string `json:"name"`
		Status string `json:"status"` // pending | verified | failed
		ReqID  string `json:"req_id"`
	}

	respObj := RespObj{
		Name:   reqObj.Name,
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

func (hs *HubServer) makeTXTStr(name string) string {
	// 生成 TXT 记录验证字符串
	return name + "example TXT record" // hash(domain, ts, etc...)
}

func (hs *HubServer) makeFile(name string) (file, value string) {
	// 生成 CNAME 记录验证字符串
	return "example.com", name + "cname_value"
}

func (hs *HubServer) checkDNSRecords(name string) bool {
	// 检查 DNS 记录
	return true
}

func (hs *HubServer) checkFileUpload(name string) bool {
	// 检查文件上传
	return true
}
