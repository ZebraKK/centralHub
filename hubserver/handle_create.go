package hubserver

import (
	"centralHub/logger"
	"centralHub/model"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func (hs *HubServer) preCreateCheck(rlog *zerolog.Logger, reqObj *model.CreateDomainRequest) *model.CDNDomain {
	// 请求，任务检测( 防止重复提交？ 排队？ 不同请求？ 覆盖？)

	// 域名所有权检查(ownership)
	//
	name := reqObj.Name
	hs.getOwnership(name)

	// 域名有效性检查(备案) ICP
	// get ICP info from govt API

	// 域名各配置项检查

	return &model.CDNDomain{
		Name: name,
	}

}

func (hs *HubServer) HandleCreate(c *gin.Context) {

	reqid, _ := c.Get("reqid")
	rlog := logger.WithReqID(reqid.(string))

	var reqObj model.CreateDomainRequest
	//parse form data
	if err := c.ShouldBind(&reqObj); err != nil {
		rlog.Error().Err(err).Msg("Failed to bind request data")
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	rlog.Info().Str("domain", reqObj.Name).Msg("Start create domain task")

	toCreateDomain := hs.preCreateCheck(&rlog, &reqObj)
	// task pipeline
	taskId := "okay"
	//taskId := hs.workflow.PushTask()
	hs.workflow.CreateDomain(c, &rlog, toCreateDomain)

	// build Cname  source Cname
	// midsrc
	// provider CDN configure
	// double-check(test)
	//

	resp := model.CreateDomainResponse{
		Name:       reqObj.Name,
		JobID:      taskId,
		StatusCode: 200,
		StatusMsg:  "Domain creation task submitted successfully",
	}
	c.JSON(200, resp)
	//c.JSON(200, gin.H{"task_id": taskId})
	// write http response , taskId
	//
	// error
}
