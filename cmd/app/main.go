package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vadym-H/Student-Complaint-Portal/internal/config"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/handlers"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/lib/logger"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/middleware"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/services"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/services/cosmos"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.ENV)
	log.Info("application starting", slog.String("env", cfg.ENV), slog.String("port", cfg.HTTPPort))

	// Initialize Cosmos DB service
	cosmosService, err := cosmos.NewCosmosService(
		cfg.CosmosDB.Endpoint,
		cfg.CosmosDB.Key,
		cfg.CosmosDB.Database,
		log,
	)
	if err != nil {
		log.Error("failed to initialize cosmos DB service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize Service Bus service
	serviceBusService, err := services.NewServiceBusService(cfg.ServiceBusConnection, log)
	if err != nil {
		log.Error("failed to initialize service bus service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cosmosService, cfg.JWTSecret, log)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Public routes
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/logout", authHandler.Logout)

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(cfg.JWTSecret, log))

		// Health check endpoint (private)
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})

		// TODO: Add complaint routes here
		// r.Get("/api/complaints", complaintHandler.GetComplaints)
		// r.Post("/api/complaints", complaintHandler.CreateComplaint)

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin(log))

			// TODO: Add admin routes here
			// r.Patch("/api/complaints/{id}/status", complaintHandler.UpdateComplaintStatus)
		})
	})

	// Setup HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("HTTP server starting", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", slog.String("error", err.Error()))
	}

	log.Info("server stopped gracefully")

	// Suppress unused variable warnings
	_ = serviceBusService
}
