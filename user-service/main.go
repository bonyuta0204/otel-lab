package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	userpb "github.com/bonyuta0204/otel-lab/proto/userpb"
	"github.com/bonyuta0204/otel-lab/user-service/server"
	"github.com/bonyuta0204/otel-lab/user-service/storage"
	"github.com/bonyuta0204/otel-lab/user-service/tracing"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	// Initialize tracing
	tp, err := tracing.InitTracer(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Initialize in-memory storage
	userStore := storage.NewInMemoryUserStore()

	// Initialize repository
	repo := storage.NewUserRepository(userStore)

	// Seed some initial data
	repo.SeedInitialData(ctx)

	// Initialize server
	userServer := server.NewUserServer(repo)

	// Setup gRPC server
	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	userpb.RegisterUserServiceServer(s, userServer)

	// Start server
	go func() {
		log.Println("User Service starting on :8082")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	s.GracefulStop()
	log.Println("Server exited")
}