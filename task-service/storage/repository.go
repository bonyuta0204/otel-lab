package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	taskpb "github.com/bonyuta0204/otel-lab/proto/taskpb"
	"github.com/bonyuta0204/otel-lab/task-service/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskRepository struct {
	db *PostgresDB
}

func NewTaskRepository(db *PostgresDB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.Task, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "TaskRepository.CreateTask")
	defer span.End()

	span.SetAttributes(
		attribute.String("task.title", req.Title),
		attribute.String("task.assignee_id", req.AssigneeId),
	)

	query := `
		INSERT INTO tasks (title, description, assignee_id)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, status, assignee_id, created_at, updated_at
	`

	var task taskpb.Task
	var createdAt, updatedAt time.Time
	var status int32

	err := r.db.DB().QueryRowContext(ctx, query, req.Title, req.Description, req.AssigneeId).Scan(
		&task.Id,
		&task.Title,
		&task.Description,
		&status,
		&task.AssigneeId,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create task")
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	task.Status = taskpb.TaskStatus(status)
	task.CreatedAt = timestamppb.New(createdAt)
	task.UpdatedAt = timestamppb.New(updatedAt)

	span.SetAttributes(attribute.String("task.id", task.Id))

	return &task, nil
}

func (r *TaskRepository) GetTask(ctx context.Context, id string) (*taskpb.Task, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "TaskRepository.GetTask")
	defer span.End()

	span.SetAttributes(attribute.String("task.id", id))

	query := `
		SELECT id, title, description, status, assignee_id, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	var task taskpb.Task
	var createdAt, updatedAt time.Time
	var status int32

	err := r.db.DB().QueryRowContext(ctx, query, id).Scan(
		&task.Id,
		&task.Title,
		&task.Description,
		&status,
		&task.AssigneeId,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			span.SetStatus(codes.Error, "Task not found")
			return nil, fmt.Errorf("task not found")
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get task")
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	task.Status = taskpb.TaskStatus(status)
	task.CreatedAt = timestamppb.New(createdAt)
	task.UpdatedAt = timestamppb.New(updatedAt)

	return &task, nil
}

func (r *TaskRepository) ListTasks(ctx context.Context, req *taskpb.ListTasksRequest) ([]*taskpb.Task, int32, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "TaskRepository.ListTasks")
	defer span.End()

	span.SetAttributes(
		attribute.Int("page.size", int(req.PageSize)),
		attribute.Int("page.number", int(req.PageNumber)),
		attribute.String("filter.assignee_id", req.AssigneeId),
		attribute.Int("filter.status", int(req.Status)),
	)

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if req.AssigneeId != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND assignee_id = $%d", argCount)
		args = append(args, req.AssigneeId)
	}

	if req.Status != taskpb.TaskStatus_TODO {
		argCount++
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, int32(req.Status))
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks %s", whereClause)
	var totalCount int32
	err := r.db.DB().QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to count tasks")
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Main query with pagination
	offset := (req.PageNumber - 1) * req.PageSize
	argCount++
	limitClause := fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argCount)
	args = append(args, req.PageSize)

	argCount++
	offsetClause := fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	query := fmt.Sprintf(`
		SELECT id, title, description, status, assignee_id, created_at, updated_at
		FROM tasks %s%s%s
	`, whereClause, limitClause, offsetClause)

	rows, err := r.db.DB().QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to list tasks")
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*taskpb.Task
	for rows.Next() {
		var task taskpb.Task
		var createdAt, updatedAt time.Time
		var status int32

		err := rows.Scan(
			&task.Id,
			&task.Title,
			&task.Description,
			&status,
			&task.AssigneeId,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to scan task")
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
		}

		task.Status = taskpb.TaskStatus(status)
		task.CreatedAt = timestamppb.New(createdAt)
		task.UpdatedAt = timestamppb.New(updatedAt)

		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Row iteration error")
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	span.SetAttributes(attribute.Int("result.count", len(tasks)))

	return tasks, totalCount, nil
}

func (r *TaskRepository) UpdateTask(ctx context.Context, req *taskpb.UpdateTaskRequest) (*taskpb.Task, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "TaskRepository.UpdateTask")
	defer span.End()

	span.SetAttributes(attribute.String("task.id", req.Id))

	query := `
		UPDATE tasks 
		SET title = $2, description = $3, status = $4, assignee_id = $5
		WHERE id = $1
		RETURNING id, title, description, status, assignee_id, created_at, updated_at
	`

	var task taskpb.Task
	var createdAt, updatedAt time.Time
	var status int32

	err := r.db.DB().QueryRowContext(ctx, query,
		req.Id,
		req.Title,
		req.Description,
		int32(req.Status),
		req.AssigneeId,
	).Scan(
		&task.Id,
		&task.Title,
		&task.Description,
		&status,
		&task.AssigneeId,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			span.SetStatus(codes.Error, "Task not found")
			return nil, fmt.Errorf("task not found")
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update task")
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	task.Status = taskpb.TaskStatus(status)
	task.CreatedAt = timestamppb.New(createdAt)
	task.UpdatedAt = timestamppb.New(updatedAt)

	return &task, nil
}

func (r *TaskRepository) DeleteTask(ctx context.Context, id string) error {
	ctx, span := tracing.GetTracer().Start(ctx, "TaskRepository.DeleteTask")
	defer span.End()

	span.SetAttributes(attribute.String("task.id", id))

	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.DB().ExecContext(ctx, query, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete task")
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		span.SetStatus(codes.Error, "Task not found")
		return fmt.Errorf("task not found")
	}

	return nil
}