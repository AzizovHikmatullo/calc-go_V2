package agent

type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
	ExpressionID  string  `json:"expression_id"`
}

type TaskRequest struct {
	Task Task `json:"task"`
}

type Agent struct {
	cntGoroutines int
	pingTime      int
}

var ping int
