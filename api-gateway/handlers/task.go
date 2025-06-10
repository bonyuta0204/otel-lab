package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bonyuta0204/otel-lab/api-gateway/tracing"
	taskpb "github.com/bonyuta0204/otel-lab/proto/taskpb"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
)

type TaskHandler struct {
	client taskpb.TaskServiceClient
	conn   *grpc.ClientConn
}

func NewTaskHandler() (*TaskHandler, error) {
	taskServiceAddr := os.Getenv("TASK_SERVICE_ADDR")
	if taskServiceAddr == "" {
		taskServiceAddr = "localhost:8081"
	}

	conn, err := grpc.NewClient(
		taskServiceAddr,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}

	return &TaskHandler{
		client: taskpb.NewTaskServiceClient(conn),
		conn:   conn,
	}, nil
}

func (h *TaskHandler) Close() error {
	return h.conn.Close()
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	AssigneeID  string `json:"assignee_id"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	AssigneeID  string `json:"assignee_id,omitempty"`
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.GetTracer().Start(r.Context(), "CreateTask")
	defer span.End()

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to decode request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("task.title", req.Title),
		attribute.String("task.assignee_id", req.AssigneeID),
	)

	pbReq := &taskpb.CreateTaskRequest{
		Title:       req.Title,
		Description: req.Description,
		AssigneeId:  req.AssigneeID,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	task, err := h.client.CreateTask(ctx, pbReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create task")
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.String("task.id", task.Id))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.GetTracer().Start(r.Context(), "GetTask")
	defer span.End()

	vars := mux.Vars(r)
	taskID := vars["id"]

	span.SetAttributes(attribute.String("task.id", taskID))

	pbReq := &taskpb.GetTaskRequest{Id: taskID}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	task, err := h.client.GetTask(ctx, pbReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get task")
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.GetTracer().Start(r.Context(), "ListTasks")
	defer span.End()

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page_number"))
	assigneeID := r.URL.Query().Get("assignee_id")
	status := r.URL.Query().Get("status")

	if pageSize == 0 {
		pageSize = 10
	}
	if pageNumber == 0 {
		pageNumber = 1
	}

	span.SetAttributes(
		attribute.Int("page.size", pageSize),
		attribute.Int("page.number", pageNumber),
		attribute.String("filter.assignee_id", assigneeID),
		attribute.String("filter.status", status),
	)

	pbReq := &taskpb.ListTasksRequest{
		PageSize:   int32(pageSize),
		PageNumber: int32(pageNumber),
		AssigneeId: assigneeID,
	}

	if status != "" {
		if s, ok := taskpb.TaskStatus_value[status]; ok {
			pbReq.Status = taskpb.TaskStatus(s)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := h.client.ListTasks(ctx, pbReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to list tasks")
		http.Error(w, "Failed to list tasks", http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("result.count", len(resp.Tasks)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.GetTracer().Start(r.Context(), "UpdateTask")
	defer span.End()

	vars := mux.Vars(r)
	taskID := vars["id"]

	span.SetAttributes(attribute.String("task.id", taskID))

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to decode request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pbReq := &taskpb.UpdateTaskRequest{
		Id:          taskID,
		Title:       req.Title,
		Description: req.Description,
		AssigneeId:  req.AssigneeID,
	}

	if req.Status != "" {
		if s, ok := taskpb.TaskStatus_value[req.Status]; ok {
			pbReq.Status = taskpb.TaskStatus(s)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	task, err := h.client.UpdateTask(ctx, pbReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update task")
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.GetTracer().Start(r.Context(), "DeleteTask")
	defer span.End()

	vars := mux.Vars(r)
	taskID := vars["id"]

	span.SetAttributes(attribute.String("task.id", taskID))

	pbReq := &taskpb.DeleteTaskRequest{Id: taskID}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := h.client.DeleteTask(ctx, pbReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete task")
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) GetTaskTrace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	// This would typically query Jaeger API for traces related to this task
	// For now, we'll return a placeholder response
	response := map[string]interface{}{
		"task_id":    taskID,
		"message":    "Trace lookup not yet implemented",
		"jaeger_url": "http://localhost:16686/search?service=task-service&tags={\"task.id\":\"" + taskID + "\"}",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
