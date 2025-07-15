package export_server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Server represents the export HTTP server
type Server struct {
	config        *ServerConfig
	jobManager    JobManager
	exportWorker  *ExportWorker
	handler       *Handler
	server        *http.Server
	cleanupTicker *time.Ticker
	stopChan      chan os.Signal
	wg            sync.WaitGroup
}

// NewServer creates a new export server
func NewServer(config *ServerConfig) *Server {
	jobManager := NewInMemoryJobManager(config)
	exportWorker := NewExportWorker(jobManager, config)
	handler := NewHandler(jobManager, exportWorker, config)

	return &Server{
		config:       config,
		jobManager:   jobManager,
		exportWorker: exportWorker,
		handler:      handler,
		stopChan:     make(chan os.Signal, 1),
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create router
	mux := http.NewServeMux()
	s.handler.RegisterRoutes(mux)

	// Create HTTP server
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start cleanup goroutine
	s.startCleanup()

	// Start server in goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		log.Printf("Export server starting on port %d", s.config.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	signal.Notify(s.stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-s.stopChan

	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	log.Println("Shutting down export server...")

	// Stop cleanup ticker
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}

	// Create shutdown context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	// Wait for all goroutines to finish
	s.wg.Wait()

	log.Println("Export server stopped")
	return nil
}

// startCleanup starts the cleanup goroutine
func (s *Server) startCleanup() {
	s.cleanupTicker = time.NewTicker(s.config.CleanupInterval)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.cleanupTicker.C:
				if err := s.jobManager.CleanupOldJobs(); err != nil {
					log.Printf("Error during cleanup: %v", err)
				}
			case <-s.stopChan:
				return
			}
		}
	}()
}

// GetJobManager returns the job manager for testing
func (s *Server) GetJobManager() JobManager {
	return s.jobManager
}

// GetExportWorker returns the export worker for testing
func (s *Server) GetExportWorker() *ExportWorker {
	return s.exportWorker
}
