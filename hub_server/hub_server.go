package hubserver

import (
	"centralHub/service"
)

type HubServer struct {
	workflow *service.WorkflowService
}

func NewHubServer() *HubServer {
	return &HubServer{
		workflow: service.NewWorkflowService(),
	}
}
