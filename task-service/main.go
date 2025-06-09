package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	taskpb "github.com/bonyuta0204/otel-lab/proto/taskpb"
	"github.com/bonyuta0204/otel-lab/task-service/server"
	"github.com/bonyuta0204/otel-lab/task-service/storage"
	"github.com/bonyuta0204/otel-lab/task-service/tracing"
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

	// Initialize database
	db, err := storage.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repository
	repo := storage.NewTaskRepository(db)

	// Initialize server
	taskServer := server.NewTaskServer(repo)

	// Setup gRPC server
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	taskpb.RegisterTaskServiceServer(s, taskServer)

	// Start server
	go func() {
		log.Println("Task Service starting on :8081")
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