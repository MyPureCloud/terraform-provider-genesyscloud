package delay

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const (
	// Default delay range when environment variable is present but no value is set
	DefaultMaxDelaySeconds = 7
)

// ConfigurableDelay applies a configurable random delay based on an environment variable
// to prevent snapshotting issues caused by high-volume asynchronous operations.
//
// Parameters:
// - envVarName: The name of the environment variable to check for delay configuration
//
// Environment Variable Configuration:
// - If envVarName is not set: no delay is applied
// - If envVarName is set without a value: default range (0-7 seconds) is used
// - If envVarName is set with a numeric value: defines upper bound (0-N seconds)
//
// The delay is logged at debug level for observability.
func ConfigurableDelay(envVarName string) {
	// Check if the environment variable is present
	envValue, exists := os.LookupEnv(envVarName)
	if !exists {
		// No environment variable set, no delay applied
		return
	}

	var maxDelaySeconds int

	if envValue == "" {
		// Environment variable present but no value, use default
		maxDelaySeconds = DefaultMaxDelaySeconds
		log.Printf("[DEBUG] Configurable delay enabled with default range: 0-%d seconds (env: %s)", maxDelaySeconds, envVarName)
	} else {
		// Environment variable has a value, parse it
		parsedValue, err := strconv.Atoi(envValue)
		if err != nil {
			// Invalid value, log warning and use default
			log.Printf("[WARN] Invalid value for %s: %s. Using default delay range: 0-%d seconds",
				envVarName, envValue, DefaultMaxDelaySeconds)
			maxDelaySeconds = DefaultMaxDelaySeconds
		} else if parsedValue < 0 {
			// Negative value, log warning and use default
			log.Printf("[WARN] Negative value for %s: %s. Using default delay range: 0-%d seconds",
				envVarName, envValue, DefaultMaxDelaySeconds)
			maxDelaySeconds = DefaultMaxDelaySeconds
		} else {
			maxDelaySeconds = parsedValue
			log.Printf("[DEBUG] Configurable delay enabled with custom range: 0-%d seconds (env: %s)", maxDelaySeconds, envVarName)
		}
	}

	// Ensure maxDelaySeconds is non-negative (should always be true after validation above)
	if maxDelaySeconds < 0 {
		maxDelaySeconds = DefaultMaxDelaySeconds
		log.Printf("[WARN] Corrected negative maxDelaySeconds to default: 0-%d seconds", maxDelaySeconds)
	}

	// Generate random delay between 0 and maxDelaySeconds
	delaySeconds := rand.Intn(maxDelaySeconds + 1)

	if delaySeconds > 0 {
		delayDuration := time.Duration(delaySeconds) * time.Second
		log.Printf("[DEBUG] Applying configurable delay: %v (env: %s)", delayDuration, envVarName)
		time.Sleep(delayDuration)
		log.Printf("[DEBUG] Configurable delay completed (env: %s)", envVarName)
	} else {
		log.Printf("[DEBUG] Configurable delay: 0 seconds (no delay applied, env: %s)", envVarName)
	}
}

// init initializes the random seed for consistent behavior
func init() {
	rand.Seed(time.Now().UnixNano())
}
