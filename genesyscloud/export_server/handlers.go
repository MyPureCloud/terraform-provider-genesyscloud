package export_server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// Handler holds the HTTP handlers for the export server
type Handler struct {
	jobManager   JobManager
	exportWorker *ExportWorker
	config       *ServerConfig
}

// NewHandler creates a new HTTP handler
func NewHandler(jobManager JobManager, exportWorker *ExportWorker, config *ServerConfig) *Handler {
	return &Handler{
		jobManager:   jobManager,
		exportWorker: exportWorker,
		config:       config,
	}
}

// RegisterRoutes registers all the HTTP routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/exports", h.handleExports)
	mux.HandleFunc("/api/v1/exports/", h.handleExportWithID)
}

// handleExports handles the /api/v1/exports endpoint
func (h *Handler) handleExports(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.CreateExportJob(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleExportWithID handles the /api/v1/exports/{job_id} endpoints
func (h *Handler) handleExportWithID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	jobID := extractJobID(path)

	if jobID == "" {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		if strings.HasSuffix(path, "/results") {
			h.GetJobResults(w, r)
		} else if strings.HasSuffix(path, "/download") {
			h.DownloadJobFiles(w, r)
		} else {
			h.GetJobStatus(w, r)
		}
	case "DELETE":
		h.CancelJob(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// CreateExportJob handles POST /api/v1/exports
func (h *Handler) CreateExportJob(w http.ResponseWriter, r *http.Request) {
	// Authenticate request
	sdkConfig, err := AuthenticateRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Directory == "" {
		req.Directory = filepath.Join(h.config.ExportBaseDir, "default")
	}

	// Create job
	jobID, err := h.jobManager.CreateJob(req.ExportParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Start export in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), h.config.JobTimeout)
		defer cancel()

		// Store cancel function for potential cancellation
		if jobManager, ok := h.jobManager.(*InMemoryJobManager); ok {
			jobManager.SetJobWorker(jobID, cancel)
		}

		h.exportWorker.ProcessJob(ctx, jobID, req.ExportParams, sdkConfig)
	}()

	// Return 202 Accepted
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	response := CreateJobResponse{
		JobID:     jobID,
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// GetJobStatus handles GET /api/v1/exports/{job_id}
func (h *Handler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	jobID := extractJobID(r.URL.Path)

	job, err := h.jobManager.GetJobStatus(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if job.Status == JobStatusCompleted {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(job)
}

// CancelJob handles DELETE /api/v1/exports/{job_id}
func (h *Handler) CancelJob(w http.ResponseWriter, r *http.Request) {
	jobID := extractJobID(r.URL.Path)

	job, err := h.jobManager.GetJobStatus(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if job.Status == JobStatusCompleted || job.Status == JobStatusFailed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		response := map[string]interface{}{
			"job_id":  jobID,
			"status":  job.Status,
			"message": "Job already completed",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	err = h.jobManager.CancelJob(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get updated job status
	job, _ = h.jobManager.GetJobStatus(jobID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(job)
}

// GetJobResults handles GET /api/v1/exports/{job_id}/results
func (h *Handler) GetJobResults(w http.ResponseWriter, r *http.Request) {
	jobID := extractJobID(r.URL.Path)

	job, err := h.jobManager.GetJobStatus(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if job.Status != JobStatusCompleted {
		w.WriteHeader(http.StatusNoContent)
		response := map[string]interface{}{
			"job_id":  jobID,
			"status":  job.Status,
			"message": "Export still in progress",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	results, err := h.jobManager.GetJobResults(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

// DownloadJobFiles handles GET /api/v1/exports/{job_id}/download
func (h *Handler) DownloadJobFiles(w http.ResponseWriter, r *http.Request) {
	jobID := extractJobID(r.URL.Path)

	job, err := h.jobManager.GetJobStatus(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if job.Status != JobStatusCompleted {
		http.Error(w, "Job not completed", http.StatusNotFound)
		return
	}

	zipData, err := h.jobManager.GetJobFiles(jobID)
	if err != nil {
		http.Error(w, "Export files not found or job not completed", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"export-%s.zip\"", jobID))
	w.WriteHeader(http.StatusOK)
	w.Write(zipData)
}

// writeErrorResponse writes an error response
func (h *Handler) writeErrorResponse(w http.ResponseWriter, statusCode int, error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   error,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// extractJobID extracts the job ID from the URL path
func extractJobID(path string) string {
	// Remove the prefix and extract the job ID
	// Path format: /api/v1/exports/{job_id} or /api/v1/exports/{job_id}/results or /api/v1/exports/{job_id}/download
	parts := strings.Split(path, "/")
	if len(parts) >= 5 {
		return parts[4] // job_id is the 5th part
	}
	return ""
}

// sanitizePath ensures the path is safe and doesn't contain directory traversal
func (h *Handler) sanitizePath(path string) (string, error) {
	// Clean the path to remove any directory traversal attempts
	cleanPath := filepath.Clean(path)

	// Ensure the path doesn't start with .. or contain any directory traversal
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("path contains directory traversal")
	}

	return cleanPath, nil
}
