package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// NewAgent - Creates new agent with specified constants
func NewAgent(cntGoroutines, pingTime int) *Agent {
	return &Agent{
		cntGoroutines: cntGoroutines,
		pingTime:      pingTime,
	}
}

// Run - starts N workers with WaitGroup in each
func (a *Agent) Run() {
	ping = a.pingTime

	log.Printf("Starting %v workers", a.cntGoroutines)

	var wg sync.WaitGroup
	for i := 0; i < a.cntGoroutines; i++ {
		wg.Add(1)
		go worker(&wg)
	}
	wg.Wait()
}

// worker - main logic. Get task, calculate it and send task to orchestrator
func worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		task, err := getTask()
		if err != nil || task == nil {
			time.Sleep(time.Duration(ping) * time.Millisecond)
			continue
		}

		result := calculate(task.Task)

		log.Printf("Get task: %v %v %v. Result: %v", task.Task.Arg1, task.Task.Operation, task.Task.Arg2, result)

		err = sendTask(task.Task, result)
		if err != nil {
			log.Printf("Error sending task: %v", err)
		}
	}
}

// getTask - gets task from orchestrator
func getTask() (*TaskRequest, error) {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, err
	}

	var res *TaskRequest
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res, nil
}

// calculate - wait for operation time and return calculation of two arguments
func calculate(task Task) float64 {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		return task.Arg1 / task.Arg2
	default:
		return 0
	}
}

// sendTask - sends ready task to orchestrator
func sendTask(task Task, result float64) error {
	data, _ := json.Marshal(map[string]interface{}{"id": task.ID, "result": result, "expression_id": task.ExpressionID})

	_, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	return nil
}
