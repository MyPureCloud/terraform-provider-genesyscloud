package mrmo

import (
	"errors"
	"sync"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
	Defines the client configuration to be used by the MRMO service.
	MRMO will perform the authenication itself, populate the clientConfig var, and set isActive to true.
	When MRMO is active (i.e. MRMO is using code defined in this codebase), specialised code will be used to
	perform the CRUD operations and to export resources. These specialised code paths are defined to serve MRMO
	the functionality it requires without all the bells and whistles (client pool, export features, etc.)
*/

var (
	clientConfig *platformclientv2.Configuration
	isActive     bool
	mutex        sync.RWMutex
)

var (
	ErrConfigNil = errors.New("client configuration is nil")
	ErrNotActive = errors.New("MRMO is not active")
)

// GetClientConfig returns the current client configuration
func GetClientConfig() (*platformclientv2.Configuration, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	if !isActive {
		return nil, ErrNotActive
	}

	if clientConfig == nil {
		return nil, ErrConfigNil
	}

	return clientConfig, nil
}

// IsActive returns whether MRMO is currently active
func IsActive() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return isActive
}

// Activate activates MRMO with the provided client configuration
func Activate(config *platformclientv2.Configuration) error {
	if config == nil {
		return ErrConfigNil
	}

	mutex.Lock()
	defer mutex.Unlock()

	clientConfig = config
	isActive = true
	return nil
}

// Reset completely resets the MRMO state (useful for testing)
func Reset() {
	mutex.Lock()
	defer mutex.Unlock()

	clientConfig = nil
	isActive = false
}
