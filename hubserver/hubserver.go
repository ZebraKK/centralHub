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

func (hs *HubServer) getOwnership(domain string) (string, error) {

	// db query
	// hs.ownershipDB.get(domain)
	return "", nil
}
