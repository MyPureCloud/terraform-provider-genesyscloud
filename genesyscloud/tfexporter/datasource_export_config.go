package tfexporter

import (
	"sync"
)

var (
	DataSourceExports []string
	dsMutex           sync.Mutex
)

func SetDataSourceExports() []string {
	if len(DataSourceExports) < 1 {
		DataSourceExports = append(DataSourceExports, "")
	}
	return DataSourceExports
}

func AddToDataSource(resource string) {
	dsMutex.Lock()
	defer dsMutex.Unlock()
	DataSourceExports = append(DataSourceExports, resource)
}
