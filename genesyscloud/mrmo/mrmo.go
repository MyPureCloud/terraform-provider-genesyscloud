package mrmo

import (
	"errors"
	"os"
	"sync"

	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

/*
	When MRMO is active (i.e. MRMO is using code defined in this codebase), specialised code will be used to
	perform the CRUD operations and to export resources. These specialised code paths are defined to serve MRMO
	the functionality it requires without all the bells and whistles (client pool, unnecessary export features, etc.)
*/

var (
	clientConfig *platformclientv2.Configuration
	mutex        sync.RWMutex
)

var (
	ErrConfigNil                      = errors.New("client configuration is nil")
	ErrNotActive                      = errors.New("MRMO is not active")
	MRMO_CXASCODE_INTEGRATION_ENABLED = "MRMO_CXASCODE_INTEGRATION_ENABLED"
)

// GetClientConfig returns the current client configuration
func GetClientConfig() (*platformclientv2.Configuration, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	if !IsActive() {
		return nil, ErrNotActive
	}

	if clientConfig == nil {
		return nil, ErrConfigNil
	}

	return clientConfig, nil
}

// IsActive returns whether MRMO is currently active
func IsActive() bool {
	return os.Getenv(MRMO_CXASCODE_INTEGRATION_ENABLED) != ""
}

// Activate activates MRMO with the provided client configuration
func Activate(config *platformclientv2.Configuration) error {
	if config == nil {
		return ErrConfigNil
	}

	mutex.Lock()
	defer mutex.Unlock()

	clientConfig = config
	return nil
}

// Reset completely resets the MRMO state (useful for testing)
func Reset() {
	mutex.Lock()
	defer mutex.Unlock()

	clientConfig = nil
}
