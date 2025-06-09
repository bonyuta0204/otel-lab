package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	userpb "github.com/bonyuta0204/otel-lab/proto/userpb"
	"github.com/bonyuta0204/otel-lab/user-service/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type InMemoryUserStore struct {
	mu    sync.RWMutex
	users map[string]*userpb.User
	cache map[string]*userpb.User // Simple cache simulation
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*userpb.User),
		cache: make(map[string]*userpb.User),
	}
}

type UserRepository struct {
	store *InMemoryUserStore
}

func NewUserRepository(store *InMemoryUserStore) *UserRepository {
	return &UserRepository{store: store}
}

func (r *UserRepository) SeedInitialData(ctx context.Context) {
	ctx, span := tracing.GetTracer().Start(ctx, "UserRepository.SeedInitialData")
	defer span.End()

	users := []*userpb.User{
		{
			Id:        "user-001",
			Name:      "Alice Johnson",
			Email:     "alice@example.com",
			CreatedAt: timestamppb.New(time.Now().Add(-24 * time.Hour)),
		},
		{
			Id:        "user-002",
			Name:      "Bob Smith",
			Email:     "bob@example.com",
			CreatedAt: timestamppb.New(time.Now().Add(-12 * time.Hour)),
		},
		{
			Id:        "user-003",
			Name:      "Charlie Brown",
			Email:     "charlie@example.com",
			CreatedAt: timestamppb.New(time.Now().Add(-6 * time.Hour)),
		},
	}

	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	for _, user := range users {
		r.store.users[user.Id] = user
	}

	span.SetAttributes(attribute.Int("seed.count", len(users)))
}

func (r *UserRepository) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.User, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "UserRepository.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.name", req.Name),
		attribute.String("user.email", req.Email),
	)

	user := &userpb.User{
		Id:        generateUserID(),
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: timestamppb.New(time.Now()),
	}

	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	r.store.users[user.Id] = user

	span.SetAttributes(attribute.String("user.id", user.Id))

	return user, nil
}

func (r *UserRepository) GetUser(ctx context.Context, id string) (*userpb.User, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "UserRepository.GetUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id))

	// Check cache first
	_, cacheSpan := tracing.GetTracer().Start(ctx, "cache.Get")
	r.store.mu.RLock()
	if cached, exists := r.store.cache[id]; exists {
		r.store.mu.RUnlock()
		cacheSpan.SetAttributes(attribute.Bool("cache.hit", true))
		cacheSpan.End()
		return cached, nil
	}
	r.store.mu.RUnlock()
	cacheSpan.SetAttributes(attribute.Bool("cache.hit", false))
	cacheSpan.End()

	// Get from store
	r.store.mu.RLock()
	user, exists := r.store.users[id]
	r.store.mu.RUnlock()

	if !exists {
		span.SetStatus(codes.Error, "User not found")
		return nil, fmt.Errorf("user not found")
	}

	// Update cache
	r.store.mu.Lock()
	r.store.cache[id] = user
	r.store.mu.Unlock()

	return user, nil
}

func (r *UserRepository) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) ([]*userpb.User, int32, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "UserRepository.ListUsers")
	defer span.End()

	span.SetAttributes(
		attribute.Int("page.size", int(req.PageSize)),
		attribute.Int("page.number", int(req.PageNumber)),
	)

	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	// Convert map to slice
	allUsers := make([]*userpb.User, 0, len(r.store.users))
	for _, user := range r.store.users {
		allUsers = append(allUsers, user)
	}

	// Simple pagination
	start := int((req.PageNumber - 1) * req.PageSize)
	end := int(req.PageNumber * req.PageSize)

	if start >= len(allUsers) {
		span.SetAttributes(attribute.Int("result.count", 0))
		return []*userpb.User{}, int32(len(allUsers)), nil
	}

	if end > len(allUsers) {
		end = len(allUsers)
	}

	users := allUsers[start:end]
	span.SetAttributes(attribute.Int("result.count", len(users)))

	return users, int32(len(allUsers)), nil
}

func (r *UserRepository) GetUsersByIds(ctx context.Context, ids []string) ([]*userpb.User, error) {
	ctx, span := tracing.GetTracer().Start(ctx, "UserRepository.GetUsersByIds")
	defer span.End()

	span.SetAttributes(
		attribute.StringSlice("user.ids", ids),
		attribute.Int("ids.count", len(ids)),
	)

	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	var users []*userpb.User
	for _, id := range ids {
		if user, exists := r.store.users[id]; exists {
			users = append(users, user)
		}
	}

	span.SetAttributes(attribute.Int("result.count", len(users)))

	return users, nil
}

func generateUserID() string {
	return fmt.Sprintf("user-%d", time.Now().UnixNano())
}