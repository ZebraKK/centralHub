package workflow

type Workflow struct {
}

func NewWorkflow() *Workflow {
	return &Workflow{}
}

func (ws *Workflow) PushTask() string {

	return "task-12345"
}
