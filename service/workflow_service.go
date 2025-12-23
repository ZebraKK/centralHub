package service

type WorkflowService struct {
}

func NewWorkflowService() *WorkflowService {
	return &WorkflowService{}
}

func (ws *WorkflowService) PushTask() string {

	return "task-12345"
}
