package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/bonyuta0204/otel-lab/api-gateway/handlers"
	"github.com/bonyuta0204/otel-lab/api-gateway/middleware"
	"github.com/bonyuta0204/otel-lab/api-gateway/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

	// Initialize handlers
	taskHandler, err := handlers.NewTaskHandler()
	if err != nil {
		log.Fatalf("Failed to initialize task handler: %v", err)
	}
	defer taskHandler.Close()

	userHandler, err := handlers.NewUserHandler()
	if err != nil {
		log.Fatalf("Failed to initialize user handler: %v", err)
	}
	defer userHandler.Close()

	// Setup routes
	r := mux.NewRouter()
	
	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS)

	// Task routes
	r.HandleFunc("/api/tasks", taskHandler.CreateTask).Methods("POST")
	r.HandleFunc("/api/tasks", taskHandler.ListTasks).Methods("GET")
	r.HandleFunc("/api/tasks/{id}", taskHandler.GetTask).Methods("GET")
	r.HandleFunc("/api/tasks/{id}", taskHandler.UpdateTask).Methods("PUT")
	r.HandleFunc("/api/tasks/{id}", taskHandler.DeleteTask).Methods("DELETE")
	r.HandleFunc("/api/tasks/{id}/trace", taskHandler.GetTaskTrace).Methods("GET")

	// User routes
	r.HandleFunc("/api/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/api/users", userHandler.ListUsers).Methods("GET")
	r.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods("GET")

	// Health check
	r.HandleFunc("/health", handlers.Health).Methods("GET")

	// Wrap the router with OpenTelemetry instrumentation
	handler := otelhttp.NewHandler(r, "api-gateway")

	// Setup server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Println("API Gateway starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}