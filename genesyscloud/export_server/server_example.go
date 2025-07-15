package export_server

import (
	"log"
)

// StartServer starts the export server with default configuration
func StartServer() {
	// Create default configuration
	config := DefaultConfig()

	// Create and start server
	server := NewServer(config)

	log.Printf("Starting Genesys Cloud Export Server on port %d", config.Port)
	log.Printf("Export base directory: %s", config.ExportBaseDir)

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
