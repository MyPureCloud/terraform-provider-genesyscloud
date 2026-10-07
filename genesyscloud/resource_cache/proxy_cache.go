package resource_cache

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
)

// CacheForProxy returns a dedicated cache for MRMO standalone mode (export then apply in
// one process with distinct org credentials). Legacy Terraform and unit tests keep the
// shared package-level cache singleton.
func CacheForProxy[T any](shared CacheInterface[T]) CacheInterface[T] {
	if mrmo.IsActive() {
		return NewResourceCache[T]()
	}
	return shared
}
