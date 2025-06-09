package server

import (
	"context"

	taskpb "github.com/bonyuta0204/otel-lab/proto/taskpb"
	"github.com/bonyuta0204/otel-lab/task-service/storage"
	"github.com/bonyuta0204/otel-lab/task-service/tracing"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	repo *storage.TaskRepository
}

func NewTaskServer(repo *storage.TaskRepository) *TaskServer {
	return &TaskServer{
		repo: repo,
	}
}

func (s *TaskServer) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.Task, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "TaskServer.CreateTask")
	defer span.End()

	span.SetAttributes(
		attribute.String("task.title", req.Title),
		attribute.String("task.assignee_id", req.AssigneeId),
	)

	if req.Title == "" {
		span.SetStatus(otelcodes.Error, "Title is required")
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	task, err := s.repo.CreateTask(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, "Failed to create task")
		return nil, status.Error(codes.Internal, "failed to create task")
	}

	span.SetAttributes(attribute.String("task.id", task.Id))

	return task, nil
}

func (s *TaskServer) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.Task, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "TaskServer.GetTask")
	defer span.End()

	span.SetAttributes(attribute.String("task.id", req.Id))

	if req.Id == "" {
		span.SetStatus(otelcodes.Error, "Task ID is required")
		return nil, status.Error(codes.InvalidArgument, "task id is required")
	}

	task, err := s.repo.GetTask(ctx, req.Id)
	if err != nil {
		span.RecordError(err)
		if err.Error() == "task not found" {
			span.SetStatus(otelcodes.Error, "Task not found")
			return nil, status.Error(codes.NotFound, "task not found")
		}
		span.SetStatus(otelcodes.Error, "Failed to get task")
		return nil, status.Error(codes.Internal, "failed to get task")
	}

	return task, nil
}

func (s *TaskServer) ListTasks(ctx context.Context, req *taskpb.ListTasksRequest) (*taskpb.ListTasksResponse, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "TaskServer.ListTasks")
	defer span.End()

	span.SetAttributes(
		attribute.Int("page.size", int(req.PageSize)),
		attribute.Int("page.number", int(req.PageNumber)),
		attribute.String("filter.assignee_id", req.AssigneeId),
		attribute.Int("filter.status", int(req.Status)),
	)

	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageNumber <= 0 {
		req.PageNumber = 1
	}

	tasks, totalCount, err := s.repo.ListTasks(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, "Failed to list tasks")
		return nil, status.Error(codes.Internal, "failed to list tasks")
	}

	span.SetAttributes(attribute.Int("result.count", len(tasks)))

	return &taskpb.ListTasksResponse{
		Tasks:      tasks,
		TotalCount: totalCount,
	}, nil
}

func (s *TaskServer) UpdateTask(ctx context.Context, req *taskpb.UpdateTaskRequest) (*taskpb.Task, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "TaskServer.UpdateTask")
	defer span.End()

	span.SetAttributes(attribute.String("task.id", req.Id))

	if req.Id == "" {
		span.SetStatus(otelcodes.Error, "Task ID is required")
		return nil, status.Error(codes.InvalidArgument, "task id is required")
	}

	task, err := s.repo.UpdateTask(ctx, req)
	if err != nil {
		span.RecordError(err)
		if err.Error() == "task not found" {
			span.SetStatus(otelcodes.Error, "Task not found")
			return nil, status.Error(codes.NotFound, "task not found")
		}
		span.SetStatus(otelcodes.Error, "Failed to update task")
		return nil, status.Error(codes.Internal, "failed to update task")
	}

	return task, nil
}

func (s *TaskServer) DeleteTask(ctx context.Context, req *taskpb.DeleteTaskRequest) (*emptypb.Empty, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "TaskServer.DeleteTask")
	defer span.End()

	span.SetAttributes(attribute.String("task.id", req.Id))

	if req.Id == "" {
		span.SetStatus(otelcodes.Error, "Task ID is required")
		return nil, status.Error(codes.InvalidArgument, "task id is required")
	}

	err := s.repo.DeleteTask(ctx, req.Id)
	if err != nil {
		span.RecordError(err)
		if err.Error() == "task not found" {
			span.SetStatus(otelcodes.Error, "Task not found")
			return nil, status.Error(codes.NotFound, "task not found")
		}
		span.SetStatus(otelcodes.Error, "Failed to delete task")
		return nil, status.Error(codes.Internal, "failed to delete task")
	}

	return &emptypb.Empty{}, nil
}