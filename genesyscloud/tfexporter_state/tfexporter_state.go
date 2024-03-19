package tfexporter_state

import (
	"log"
	"sync"
)

/*
Export state is used to indicate whether an export is being done.  If the export state is set to true, then this should be
a signal that any resources being exported should be reading their data from each resource's internal cache rather then the API.
*/
var exportState bool
var once sync.Once

// ActivateExporterState will be used to indicate that caching should be used to process requests.
// We are setting this as an environment variable so we can experiment with it, without creating an attribute
// on the resource
func ActivateExporterState() {
	once.Do(func() {
		log.Printf("Exporter State is active")
		exportState = true
	})
}

func IsExporterActive() bool {
	return exportState
}
