package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bonyuta0204/otel-lab/api-gateway/tracing"
	userpb "github.com/bonyuta0204/otel-lab/proto/userpb"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
)

type UserHandler struct {
	client userpb.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserHandler() (*UserHandler, error) {
	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "localhost:8082"
	}

	conn, err := grpc.NewClient(
		userServiceAddr,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)

	if err != nil {
		return nil, err
	}

	return &UserHandler{
		client: userpb.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

func (h *UserHandler) Close() error {
	return h.conn.Close()
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.GetTracer().Start(r.Context(), "CreateUser")
	defer span.End()

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to decode request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("user.name", req.Name),
		attribute.String("user.email", req.Email),
	)

	pbReq := &userpb.CreateUserRequest{
		Name:  req.Name,
		Email: req.Email,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := h.client.CreateUser(ctx, pbReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create user")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.String("user.id", user.Id))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.GetTracer().Start(r.Context(), "GetUser")
	defer span.End()

	vars := mux.Vars(r)
	userID := vars["id"]

	span.SetAttributes(attribute.String("user.id", userID))

	pbReq := &userpb.GetUserRequest{Id: userID}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := h.client.GetUser(ctx, pbReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get user")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.GetTracer().Start(r.Context(), "ListUsers")
	defer span.End()

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page_number"))

	if pageSize == 0 {
		pageSize = 10
	}
	if pageNumber == 0 {
		pageNumber = 1
	}

	span.SetAttributes(
		attribute.Int("page.size", pageSize),
		attribute.Int("page.number", pageNumber),
	)

	pbReq := &userpb.ListUsersRequest{
		PageSize:   int32(pageSize),
		PageNumber: int32(pageNumber),
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := h.client.ListUsers(ctx, pbReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to list users")
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("result.count", len(resp.Users)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
