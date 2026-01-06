package hubserver

import (
	"centralHub/logger"

	"github.com/gin-gonic/gin"
)

func (hs *HubServer) preCreateCheck() {
	// 请求，任务检测( 防止重复提交？ 排队？ 不同请求？ 覆盖？)

	// 域名所有权检查

	// 域名有效性检查(备案)

	// 域名检查
}

func (hs *HubServer) HandleCreate(c *gin.Context) {

	reqid, _ := c.Get("reqid")
	rlog := logger.WithReqID(reqid.(string))

	obj := struct { // todo request struct define
		DomainName string `form:"domain_name" binding:"required"`
		Owner      string `form:"owner" binding:"required"`
	}{}
	//parse form data
	if err := c.ShouldBind(&obj); err != nil {
		rlog.Error().Err(err).Msg("Failed to bind request data")
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	rlog.Info().Str("domain", obj.DomainName).Str("owner", obj.Owner).Msg("Start create domain task")

	hs.preCreateCheck()
	// task pipeline
	taskId := "okay"
	//taskId := hs.workflow.PushTask()
	hs.workflow.CreateDomain(c, obj)

	// build Cname  source Cname
	// midsrc
	// provider CDN configure
	// double-check(test)
	//

	// json response todo
	c.JSON(200, gin.H{"task_id": taskId})
	// write http response , taskId
	//
	// error
}
