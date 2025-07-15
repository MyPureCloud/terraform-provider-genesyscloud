package export_server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// JobManager interface defines the contract for job management
type JobManager interface {
	CreateJob(params ExportParams) (string, error)
	GetJobStatus(jobID string) (*JobStatus, error)
	CancelJob(jobID string) error
	GetJobResults(jobID string) (*JobResults, error)
	GetJobFiles(jobID string) ([]byte, error)
	UpdateJobProgress(jobID string, progress int) error
	CompleteJob(jobID string, success bool, errorMessage string) error
	CleanupOldJobs() error
}

// InMemoryJobManager implements JobManager with in-memory storage
type InMemoryJobManager struct {
	jobs       map[string]*JobStatus
	config     *ServerConfig
	mutex      sync.RWMutex
	jobWorkers map[string]context.CancelFunc
}

// NewInMemoryJobManager creates a new in-memory job manager
func NewInMemoryJobManager(config *ServerConfig) *InMemoryJobManager {
	return &InMemoryJobManager{
		jobs:       make(map[string]*JobStatus),
		config:     config,
		jobWorkers: make(map[string]context.CancelFunc),
	}
}

// CreateJob creates a new export job
func (m *InMemoryJobManager) CreateJob(params ExportParams) (string, error) {
	jobID := generateJobID()
	now := time.Now()

	// Validate export parameters
	if err := validateExportParams(params); err != nil {
		return "", fmt.Errorf("invalid export parameters: %w", err)
	}

	// Create job directory
	jobDir := filepath.Join(m.config.ExportBaseDir, jobID)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create job directory: %w", err)
	}

	job := &JobStatus{
		JobID:     jobID,
		Status:    JobStatusPending,
		Progress:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m.mutex.Lock()
	m.jobs[jobID] = job
	m.mutex.Unlock()

	return jobID, nil
}

// GetJobStatus retrieves the status of a job
func (m *InMemoryJobManager) GetJobStatus(jobID string) (*JobStatus, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// CancelJob cancels a running job
func (m *InMemoryJobManager) CancelJob(jobID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status == JobStatusCompleted || job.Status == JobStatusFailed {
		return nil // Job already finished
	}

	// Cancel the worker if it's running
	if cancelFunc, exists := m.jobWorkers[jobID]; exists {
		cancelFunc()
		delete(m.jobWorkers, jobID)
	}

	now := time.Now()
	job.Status = JobStatusCancelled
	job.UpdatedAt = now
	job.CancelledAt = &now

	return nil
}

// GetJobResults retrieves the results of a completed job
func (m *InMemoryJobManager) GetJobResults(jobID string) (*JobResults, error) {
	job, err := m.GetJobStatus(jobID)
	if err != nil {
		return nil, err
	}

	if job.Status != JobStatusCompleted {
		return nil, fmt.Errorf("job not completed: %s", job.Status)
	}

	jobDir := filepath.Join(m.config.ExportBaseDir, jobID)
	files, err := getExportFiles(jobDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get export files: %w", err)
	}

	return &JobResults{
		JobID:           jobID,
		Status:          job.Status,
		Files:           files,
		ExportDirectory: jobDir,
	}, nil
}

// GetJobFiles returns the ZIP file of a completed job
func (m *InMemoryJobManager) GetJobFiles(jobID string) ([]byte, error) {
	job, err := m.GetJobStatus(jobID)
	if err != nil {
		return nil, err
	}

	if job.Status != JobStatusCompleted {
		return nil, fmt.Errorf("job not completed: %s", job.Status)
	}

	jobDir := filepath.Join(m.config.ExportBaseDir, jobID)
	zipFile := filepath.Join(jobDir, "export.zip")

	if _, err := os.Stat(zipFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("export files not found")
	}

	return os.ReadFile(zipFile)
}

// UpdateJobProgress updates the progress of a job
func (m *InMemoryJobManager) UpdateJobProgress(jobID string, progress int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Progress = progress
	job.UpdatedAt = time.Now()

	return nil
}

// CompleteJob marks a job as completed or failed
func (m *InMemoryJobManager) CompleteJob(jobID string, success bool, errorMessage string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	now := time.Now()
	job.UpdatedAt = now

	if success {
		job.Status = JobStatusCompleted
		job.Progress = 100
		job.CompletedAt = &now
	} else {
		job.Status = JobStatusFailed
		job.ErrorMessage = errorMessage
	}

	// Remove worker reference
	delete(m.jobWorkers, jobID)

	return nil
}

// CleanupOldJobs removes jobs older than the configured maximum age
func (m *InMemoryJobManager) CleanupOldJobs() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cutoff := time.Now().Add(-m.config.MaxJobAge)
	var jobsToDelete []string

	for jobID, job := range m.jobs {
		if job.CreatedAt.Before(cutoff) {
			jobsToDelete = append(jobsToDelete, jobID)
		}
	}

	for _, jobID := range jobsToDelete {
		// Remove job directory
		jobDir := filepath.Join(m.config.ExportBaseDir, jobID)
		os.RemoveAll(jobDir)

		// Remove from memory
		delete(m.jobs, jobID)
	}

	return nil
}

// SetJobWorker stores the cancel function for a job worker
func (m *InMemoryJobManager) SetJobWorker(jobID string, cancelFunc context.CancelFunc) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.jobWorkers[jobID] = cancelFunc
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("job-%s", uuid.New().String()[:8])
}

// validateExportParams validates the export parameters
func validateExportParams(params ExportParams) error {
	if params.Directory == "" {
		return fmt.Errorf("directory is required")
	}

	if params.ExportFormat != "" {
		validFormats := []string{ExportFormatHCL, ExportFormatJSON, ExportFormatJSONHCL, ExportFormatHCLJSON}
		valid := false
		for _, format := range validFormats {
			if params.ExportFormat == format {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid export format: %s", params.ExportFormat)
		}
	}

	return nil
}

// getExportFiles returns a list of files in the export directory
func getExportFiles(jobDir string) ([]ExportFile, error) {
	var files []ExportFile

	err := filepath.Walk(jobDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(jobDir, path)
			if err != nil {
				return err
			}

			fileType := "unknown"
			switch filepath.Ext(path) {
			case ".tf":
				fileType = "hcl"
			case ".tf.json":
				fileType = "json"
			case ".tfstate":
				fileType = "state"
			case ".zip":
				fileType = "archive"
			}

			files = append(files, ExportFile{
				Name: relPath,
				Size: info.Size(),
				Type: fileType,
			})
		}

		return nil
	})

	return files, err
}
