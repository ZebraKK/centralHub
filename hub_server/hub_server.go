package hubserver

import (
	"centralHub/workflow"
)

type HubServer struct {
	workflow *workflow.Workflow
}

func NewHubServer() *HubServer {
	return &HubServer{
		workflow: workflow.NewWorkflow(),
	}
}
