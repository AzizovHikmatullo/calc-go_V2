package orchestrator

import (
	"encoding/json"
	"errors"
	"github.com/AzizovHikmatullo/calc-go_V2/pkg"
	"github.com/AzizovHikmatullo/calc-go_V2/pkg/calc"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Orchestrator struct {
}

// NewOrchestrator - return empty orchestrator
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{}
}

var (
	taskQueue      = make(chan *Task, 100)
	expressions    = make(map[string]*Expression)
	mu             sync.Mutex
	operationTimes = map[string]int{
		"+": 1000,
		"-": 1000,
		"*": 2000,
		"/": 3000,
	}
)

// Run - register all handlers and allow CORS. Starts server
func (o *Orchestrator) Run() {
	loadEnv()

	r := mux.NewRouter()

	r.HandleFunc("/api/v1/calculate", calculateHandler).Methods("POST")
	r.HandleFunc("/api/v1/expressions", getExpressionsHandler).Methods("GET")
	r.HandleFunc("/api/v1/expressions/{id}", getExpressionHandler).Methods("GET")
	r.HandleFunc("/internal/task", sendTaskHandler).Methods("GET")
	r.HandleFunc("/internal/task", getTaskHandler).Methods("POST")

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:8081"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	handler := corsHandler.Handler(r)

	log.Println("Starting server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal("Failed to start server")
	}
}

// loadEnv - loads consts from env
func loadEnv() {
	operationTimes["+"] = pkg.GetEnvIntWithDefault("TIME_ADDITION_MS", 1000)
	operationTimes["-"] = pkg.GetEnvIntWithDefault("TIME_SUBTRACTION_MS", 1000)
	operationTimes["*"] = pkg.GetEnvIntWithDefault("TIME_MULTIPLICATION_MS", 2000)
	operationTimes["/"] = pkg.GetEnvIntWithDefault("TIME_DIVISION_MS", 3000)
}

// calculateHandler - accepts expression from user and returns expressionID
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	var req ExpressionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}

	exprID := uuid.New().String()

	expression := &Expression{
		ID:     exprID,
		Expr:   req.Expression,
		Status: "processing",
	}

	mu.Lock()
	expressions[exprID] = expression
	mu.Unlock()

	go processExpression(expression)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(map[string]string{"id": exprID}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// processExpression - gets expression and create AST from that. Then evaluates AST
func processExpression(expr *Expression) {
	tokens, err := calc.Tokenize(expr.Expr)
	if err != nil {
		expr.setStatus("error")
		log.Println("Error tokenizing expression:", err)
		return
	}

	postfix, err := calc.InfixToPostfix(tokens)
	if err != nil || len(postfix) == 0 {
		expr.setStatus("error")
		log.Println("Error converting to postfix:", err)
		return
	}

	root, err := buildExpressionTree(postfix)
	if err != nil {
		expr.setStatus("error")
		log.Println("Error building expression tree:", err)
		return
	}

	result, err := evaluateNode(root, expr)
	if err != nil {
		expr.setStatus("error")
		log.Println("Error evaluating expression:", err)
		return
	}

	expr.setResult(result)
	expr.setStatus("completed")
	log.Printf("Complete expression: %v. Status: %v. Result: %v", expr.Expr, expr.Status, result)
}

// buildExpressionTree - builds AST from RPN
func buildExpressionTree(postfix []string) (*Node, error) {
	var stack []*Node

	for _, token := range postfix {
		if calc.IsNumber(token) {
			num, _ := strconv.ParseFloat(token, 64)
			stack = append(stack, &Node{Value: num})
		} else if calc.IsOperator(token) {
			if len(stack) < 2 {
				return nil, errors.New("invalid expression")
			}
			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, &Node{Operation: token, Left: left, Right: right})
		} else {
			return nil, errors.New("invalid token")
		}
	}

	if len(stack) != 1 {
		return nil, errors.New("invalid expression")
	}

	return stack[0], nil
}

// evaluateNode - recursive function, that creates task for every operation in AST
func evaluateNode(node *Node, expr *Expression) (float64, error) {
	if node.Operation == "" {
		return node.Value, nil
	}

	var left, right float64
	var err error

	left, err = evaluateNode(node.Left, expr)
	if err != nil {
		return 0, err
	}

	right, err = evaluateNode(node.Right, expr)
	if err != nil {
		return 0, err
	}

	task := &Task{
		ID:            uuid.New().String(),
		ExpressionID:  expr.ID,
		Operation:     node.Operation,
		Arg1:          left,
		Arg2:          right,
		Status:        "queued",
		OperationTime: operationTimes[node.Operation],
	}

	expr.addTask(task)
	taskQueue <- task

	resultChan := make(chan float64)
	go func() {
		for {
			taskStatus := getTaskStatus(task.ID, expr.ID)
			if taskStatus == "completed" {
				resultChan <- getTaskResult(task.ID, expr.ID)
				close(resultChan)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil
	}
}

// getTaskStatus - gets status of task in expression
func getTaskStatus(taskID, exprID string) string {
	mu.Lock()
	expr, exists := expressions[exprID]
	mu.Unlock()

	if !exists {
		return ""
	}

	expr.mu.Lock()
	defer expr.mu.Unlock()

	for _, task := range expr.Tasks {
		if task.ID == taskID {
			return task.Status
		}
	}
	return ""
}

// getTaskStatus - gets result of task in expression
func getTaskResult(taskID, exprID string) float64 {
	mu.Lock()
	expr, exists := expressions[exprID]
	mu.Unlock()

	if !exists {
		return 0
	}

	expr.mu.Lock()
	defer expr.mu.Unlock()

	for _, task := range expr.Tasks {
		if task.ID == taskID && task.Status == "completed" {
			return task.Result
		}
	}
	return 0
}

func (e *Expression) setStatus(status string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Status = status
}

func (e *Expression) setResult(result float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Result = result
}

func (e *Expression) addTask(task *Task) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Tasks = append(e.Tasks, task)
}

// getExpressionsHandler - creates list with all expressions and return that
func getExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var exprList []*Expression
	for _, expr := range expressions {
		expr.mu.Lock()
		exprCopy := &Expression{
			ID:     expr.ID,
			Expr:   expr.Expr,
			Status: expr.Status,
			Result: expr.Result,
			Tasks:  expr.Tasks,
		}
		exprList = append(exprList, exprCopy)
		expr.mu.Unlock()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	log.Println("Get all expressions")

	if err := json.NewEncoder(w).Encode(map[string]interface{}{"expressions": exprList}); err != nil {
		http.Error(w, "Error encoding expressions", http.StatusInternalServerError)
		return
	}
}

// getExpressionHandler - gets expression with ID
func getExpressionHandler(w http.ResponseWriter, r *http.Request) {
	exprID := mux.Vars(r)["id"]

	mu.Lock()
	expr, ok := expressions[exprID]
	mu.Unlock()

	if !ok {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	expr.mu.Lock()
	exprCopy := &Expression{
		ID:     expr.ID,
		Expr:   expr.Expr,
		Status: expr.Status,
		Result: expr.Result,
		Tasks:  expr.Tasks,
	}
	expr.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	log.Printf("Get expression: %v", expr.Expr)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{"expression": exprCopy}); err != nil {
		http.Error(w, "Error encoding expression", http.StatusInternalServerError)
		return
	}
}

// sendTaskHandler - internal function for agent. Send one task from queue
func sendTaskHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case task := <-taskQueue:
		json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
	default:
		http.Error(w, "No tasks available", http.StatusNotFound)
	}
}

// getTaskHandler - internal function for agent. Gets task from agent and checks if all expression is completed
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	var taskResult TaskResult

	if err := json.NewDecoder(r.Body).Decode(&taskResult); err != nil {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	mu.Lock()
	expr, ok := expressions[taskResult.ExpressionID]
	mu.Unlock()

	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	expr.mu.Lock()
	defer expr.mu.Unlock()

	for _, task := range expr.Tasks {
		if task.ID == taskResult.ID {
			task.Result = taskResult.Result
			task.Status = "completed"
			break
		}
	}

	allCompleted := true
	var finalResult float64
	for _, task := range expr.Tasks {
		if task.Status != "completed" {
			allCompleted = false
			break
		}
		finalResult = task.Result
	}

	if allCompleted {
		expr.Result = finalResult
		expr.Status = "completed"
	}

	w.WriteHeader(http.StatusOK)
}
