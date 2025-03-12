package orchestrator

import "sync"

type Node struct {
	Value     float64
	Operation string
	Left      *Node
	Right     *Node
	Task      *Task
}

type ExpressionRequest struct {
	Expression string `json:"expression"`
}

type TaskResult struct {
	ID           string  `json:"id"`
	Result       float64 `json:"result"`
	ExpressionID string  `json:"expression_id"`
}

type Expression struct {
	ID     string  `json:"id"`
	Expr   string  `json:"expression"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
	Tasks  []*Task `json:"-"`
	mu     sync.Mutex
}

type Task struct {
	ID            string  `json:"id"`
	ExpressionID  string  `json:"expression_id"`
	Operation     string  `json:"operation"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Result        float64 `json:"result,omitempty"`
	Status        string  `json:"status"`
	OperationTime int     `json:"operation_time"`
}
