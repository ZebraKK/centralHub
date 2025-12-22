package hubserver

import "github.com/gin-gonic/gin"

func (hs *HubServer) preCreateCheck() {
	// 请求，任务检测( 防止重复提交？ 排队？ 不同请求？ 覆盖？)

	// 域名所有权检查

	// 域名有效性检查(备案)

	// 域名检查
}

func (hs *HubServer) HandleCreate(c *gin.Context) {

	//parse form data

	hs.preCreateCheck()
	// task pipeline
	//task := hs.Workflow.PushTask()

	// write http response , taskId
	//
	// error
}
