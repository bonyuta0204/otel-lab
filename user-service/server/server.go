package server

import (
	"context"

	userpb "github.com/bonyuta0204/otel-lab/proto/userpb"
	"github.com/bonyuta0204/otel-lab/user-service/storage"
	"github.com/bonyuta0204/otel-lab/user-service/tracing"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	userpb.UnimplementedUserServiceServer
	repo *storage.UserRepository
}

func NewUserServer(repo *storage.UserRepository) *UserServer {
	return &UserServer{
		repo: repo,
	}
}

func (s *UserServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.User, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "UserServer.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.name", req.Name),
		attribute.String("user.email", req.Email),
	)

	if req.Name == "" {
		span.SetStatus(otelcodes.Error, "Name is required")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	if req.Email == "" {
		span.SetStatus(otelcodes.Error, "Email is required")
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	user, err := s.repo.CreateUser(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, "Failed to create user")
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	span.SetAttributes(attribute.String("user.id", user.Id))

	return user, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.User, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "UserServer.GetUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", req.Id))

	if req.Id == "" {
		span.SetStatus(otelcodes.Error, "User ID is required")
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	user, err := s.repo.GetUser(ctx, req.Id)
	if err != nil {
		span.RecordError(err)
		if err.Error() == "user not found" {
			span.SetStatus(otelcodes.Error, "User not found")
			return nil, status.Error(codes.NotFound, "user not found")
		}
		span.SetStatus(otelcodes.Error, "Failed to get user")
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return user, nil
}

func (s *UserServer) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "UserServer.ListUsers")
	defer span.End()

	span.SetAttributes(
		attribute.Int("page.size", int(req.PageSize)),
		attribute.Int("page.number", int(req.PageNumber)),
	)

	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageNumber <= 0 {
		req.PageNumber = 1
	}

	users, totalCount, err := s.repo.ListUsers(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, "Failed to list users")
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	span.SetAttributes(attribute.Int("result.count", len(users)))

	return &userpb.ListUsersResponse{
		Users:      users,
		TotalCount: totalCount,
	}, nil
}

func (s *UserServer) GetUsersByIds(ctx context.Context, req *userpb.GetUsersByIdsRequest) (*userpb.GetUsersByIdsResponse, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "UserServer.GetUsersByIds")
	defer span.End()

	span.SetAttributes(
		attribute.StringSlice("user.ids", req.Ids),
		attribute.Int("ids.count", len(req.Ids)),
	)

	if len(req.Ids) == 0 {
		span.SetStatus(otelcodes.Error, "User IDs are required")
		return nil, status.Error(codes.InvalidArgument, "user ids are required")
	}

	users, err := s.repo.GetUsersByIds(ctx, req.Ids)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, "Failed to get users by IDs")
		return nil, status.Error(codes.Internal, "failed to get users by ids")
	}

	span.SetAttributes(attribute.Int("result.count", len(users)))

	return &userpb.GetUsersByIdsResponse{
		Users: users,
	}, nil
}