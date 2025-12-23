package hubserver

import "github.com/gin-gonic/gin"

func (hs *HubServer) preCreateCheck() {
	// 请求，任务检测( 防止重复提交？ 排队？ 不同请求？ 覆盖？)

	// 域名所有权检查

	// 域名有效性检查(备案)

	// 域名检查
}

func (hs *HubServer) HandleCreate(c *gin.Context) {

	obj := struct {
		DomainName string `form:"domain_name" binding:"required"`
		Owner      string `form:"owner" binding:"required"`
	}{}
	//parse form data
	c.ShouldBind(&obj) // gin框架功能

	hs.preCreateCheck()
	// task pipeline
	taskId := hs.workflow.PushTask()

	// json response todo
	c.JSON(200, gin.H{"task_id": taskId})
	// write http response , taskId
	//
	// error
}
