package model

type Workflow int

const (
	UNKNOWN Workflow = iota
	REVIEW
	RETURN
	RECOMMEND
	ASK_ABOUT
)

type WorkflowFulfillment struct {
	Name      string
	Arguments map[string]string
}
