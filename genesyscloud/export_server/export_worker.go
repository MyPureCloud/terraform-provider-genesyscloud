package export_server

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter"
)

// ExportWorker handles the execution of export jobs
type ExportWorker struct {
	jobManager JobManager
	config     *ServerConfig
}

// NewExportWorker creates a new export worker
func NewExportWorker(jobManager JobManager, config *ServerConfig) *ExportWorker {
	return &ExportWorker{
		jobManager: jobManager,
		config:     config,
	}
}

// ProcessJob executes an export job
func (w *ExportWorker) ProcessJob(ctx context.Context, jobID string, params ExportParams, sdkConfig *platformclientv2.Configuration) {
	// Update job status to running
	if err := w.jobManager.UpdateJobProgress(jobID, 10); err != nil {
		log.Printf("Failed to update job progress: %v", err)
		return
	}

	// Create a mock schema.ResourceData for the tfexporter
	resourceData := createMockResourceData(params)

	// Create provider meta
	providerMeta := &provider.ProviderMeta{
		ClientConfig: sdkConfig,
		Version:      "1.0.0",
	}

	// Execute the export
	err := w.executeExport(ctx, jobID, resourceData, providerMeta, params)
	if err != nil {
		log.Printf("Export failed for job %s: %v", jobID, err)
		w.jobManager.CompleteJob(jobID, false, err.Error())
		return
	}

	// Create ZIP file
	if err := w.createZipFile(jobID); err != nil {
		log.Printf("Failed to create ZIP file for job %s: %v", jobID, err)
		w.jobManager.CompleteJob(jobID, false, "Failed to create ZIP file")
		return
	}

	// Mark job as completed
	w.jobManager.CompleteJob(jobID, true, "")
}

// executeExport runs the actual export using the tfexporter
func (w *ExportWorker) executeExport(ctx context.Context, jobID string, d *schema.ResourceData, meta interface{}, params ExportParams) error {
	// Update progress
	w.jobManager.UpdateJobProgress(jobID, 20)

	// Create the exporter
	var exporter *tfexporter.GenesysCloudResourceExporter
	var diagErr diag.Diagnostics

	// Determine filter type based on parameters
	if len(params.IncludeFilterResources) > 0 {
		exporter, diagErr = tfexporter.NewGenesysCloudResourceExporter(ctx, d, meta, tfexporter.IncludeResources)
	} else if len(params.ExcludeFilterResources) > 0 {
		exporter, diagErr = tfexporter.NewGenesysCloudResourceExporter(ctx, d, meta, tfexporter.ExcludeResources)
	} else {
		exporter, diagErr = tfexporter.NewGenesysCloudResourceExporter(ctx, d, meta, tfexporter.LegacyInclude)
	}

	if diagErr.HasError() {
		return fmt.Errorf("failed to create exporter: %v", diagErr)
	}

	// Update progress
	w.jobManager.UpdateJobProgress(jobID, 40)

	// Execute the export
	diagErr = exporter.Export()
	if diagErr.HasError() {
		return fmt.Errorf("export failed: %v", diagErr)
	}

	// Update progress
	w.jobManager.UpdateJobProgress(jobID, 80)

	return nil
}

// createZipFile creates a ZIP file containing all export files
func (w *ExportWorker) createZipFile(jobID string) error {
	jobDir := filepath.Join(w.config.ExportBaseDir, jobID)
	zipFile := filepath.Join(jobDir, "export.zip")

	// Create ZIP file
	zipWriter, err := os.Create(zipFile)
	if err != nil {
		return fmt.Errorf("failed to create ZIP file: %w", err)
	}
	defer zipWriter.Close()

	archive := zip.NewWriter(zipWriter)
	defer archive.Close()

	// Walk through the job directory and add files to ZIP
	err = filepath.Walk(jobDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and the ZIP file itself
		if info.IsDir() || filepath.Base(path) == "export.zip" {
			return nil
		}

		// Create relative path for ZIP
		relPath, err := filepath.Rel(jobDir, path)
		if err != nil {
			return err
		}

		// Create ZIP file entry
		fileWriter, err := archive.Create(relPath)
		if err != nil {
			return err
		}

		// Open and copy file content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(fileWriter, file)
		return err
	})

	return err
}

// createMockResourceData creates a mock schema.ResourceData for the tfexporter
func createMockResourceData(params ExportParams) *schema.ResourceData {
	// Create a mock resource data with the export parameters
	resourceData := &schema.ResourceData{}

	// Set the directory
	resourceData.Set("directory", params.Directory)

	// Set include filter resources
	if len(params.IncludeFilterResources) > 0 {
		resourceData.Set("include_filter_resources", params.IncludeFilterResources)
	}

	// Set exclude filter resources
	if len(params.ExcludeFilterResources) > 0 {
		resourceData.Set("exclude_filter_resources", params.ExcludeFilterResources)
	}

	// Set replace with datasource
	if len(params.ReplaceWithDatasource) > 0 {
		resourceData.Set("replace_with_datasource", params.ReplaceWithDatasource)
	}

	// Set include state file
	resourceData.Set("include_state_file", params.IncludeStateFile)

	// Set export as HCL
	resourceData.Set("export_as_hcl", params.ExportAsHCL)

	// Set export format
	resourceData.Set("export_format", params.ExportFormat)

	// Set split files by resource
	resourceData.Set("split_files_by_resource", params.SplitFilesByResource)

	// Set log permission errors
	resourceData.Set("log_permission_errors", params.LogPermissionErrors)

	// Set exclude attributes
	if len(params.ExcludeAttributes) > 0 {
		resourceData.Set("exclude_attributes", params.ExcludeAttributes)
	}

	// Set enable dependency resolution
	resourceData.Set("enable_dependency_resolution", params.EnableDependencyResolution)

	// Set ignore cyclic deps
	resourceData.Set("ignore_cyclic_deps", params.IgnoreCyclicDeps)

	// Set compress
	resourceData.Set("compress", params.Compress)

	// Set export computed
	resourceData.Set("export_computed", params.ExportComputed)

	// Set use legacy architect flow exporter
	resourceData.Set("use_legacy_architect_flow_exporter", params.UseLegacyArchitectFlowExporter)

	return resourceData
}
