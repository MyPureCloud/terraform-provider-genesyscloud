// Package homedivision looks up and caches the home division ID of the org a client config is
// authorised against. It is a leaf package so both util and provider can use the same cache.
package homedivision

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v204/platformclientv2"
	"golang.org/x/sync/singleflight"
)

// Fetcher looks up the home division ID of the org that sdkConfig is authorised against.
type Fetcher func(sdkConfig *platformclientv2.Configuration) (string, error)

// ErrNilConfig is returned when no config is given. Falling back to the SDK default configuration
// would silently return the home division of whichever org that config is authorised against.
var ErrNilConfig = errors.New("home division lookup requires a client config; got nil")

var (
	fetch Fetcher = fetchFromAPI

	// ttl is how long an entry keyed by access token is kept. A home division ID never changes for
	// an org, so the TTL is not about staleness: every token refresh produces a new key, and
	// without a TTL a long-lived process would accumulate one entry per token it has ever used.
	// Entries keyed by client ID are stable and never expire. A variable so tests can change it.
	ttl = time.Hour

	// now is the clock used for entry expiry. A variable so tests can control it.
	now = time.Now

	// Successful lookups are cached; failures are never cached so a transient error does not
	// poison later calls. Concurrent lookups for the same key share one API request.
	mu      sync.Mutex
	ids     = map[string]entry{}
	flights singleflight.Group
)

// entry is a cached home division ID. A zero expiresAt never expires.
type entry struct {
	id        string
	expiresAt time.Time
}

// ID returns the home division ID for the org that sdkConfig is authorised against. Results are
// cached per org credential, so callers working against different orgs in the same process each
// get their own org's home division.
func ID(sdkConfig *platformclientv2.Configuration) (string, error) {
	if sdkConfig == nil {
		return "", ErrNilConfig
	}
	key, expires := cacheKey(sdkConfig)
	if id, ok := cached(key); ok {
		return id, nil
	}

	id, err, _ := flights.Do(key, func() (any, error) {
		// Another caller may have stored the entry between the check above and joining the flight.
		if id, ok := cached(key); ok {
			return id, nil
		}
		mu.Lock()
		f := fetch
		mu.Unlock()

		id, err := f(sdkConfig)
		if err != nil {
			return "", err
		}
		store(key, id, expires)
		return id, nil
	})
	if err != nil {
		return "", err
	}
	return id.(string), nil
}

func cached(key string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()
	e, ok := ids[key]
	if !ok {
		return "", false
	}
	if !e.expiresAt.IsZero() && !now().Before(e.expiresAt) {
		delete(ids, key)
		return "", false
	}
	return e.id, true
}

// store caches id under key and sweeps expired entries, which keeps the map bounded without a
// background goroutine.
func store(key, id string, expires bool) {
	mu.Lock()
	defer mu.Unlock()
	t := now()
	for k, e := range ids {
		if !e.expiresAt.IsZero() && !t.Before(e.expiresAt) {
			delete(ids, k)
		}
	}
	e := entry{id: id}
	if expires {
		e.expiresAt = t.Add(ttl)
	}
	ids[key] = e
}

// cacheKey identifies the org a config is authorised against without an API call.
//
// An OAuth client belongs to exactly one org, so when Configuration.ClientID is set, region + client
// ID is the key: it is stable across token refreshes and shared by every config for that client.
// The SDK's AuthorizeClientCredentials does not set ClientID; provider.AuthorizeClientCredentials
// does, and it must name the client the config's token was issued to.
//
// Configs without a client ID are keyed by a hash of their access token instead (a token also
// belongs to exactly one org), and expires reports that the entry should age out (see ttl).
func cacheKey(sdkConfig *platformclientv2.Configuration) (key string, expires bool) {
	if sdkConfig.ClientID != "" {
		return sdkConfig.BasePath + "|client:" + sdkConfig.ClientID, false
	}
	sum := sha256.Sum256([]byte(sdkConfig.AccessToken))
	return sdkConfig.BasePath + "|token:" + hex.EncodeToString(sum[:]), true
}

func fetchFromAPI(sdkConfig *platformclientv2.Configuration) (string, error) {
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)
	homeDiv, _, err := authAPI.GetAuthorizationDivisionsHome()
	if err != nil {
		return "", err
	}
	if homeDiv == nil || homeDiv.Id == nil {
		return "", errors.New("home division response did not include an id")
	}
	return *homeDiv.Id, nil
}

// SetFetcherForTesting replaces the home division lookup and clears the cache. It returns a
// function that restores the real lookup and clears the cache again. For unit tests only.
func SetFetcherForTesting(fetcher Fetcher) (restore func()) {
	mu.Lock()
	previous := fetch
	fetch = fetcher
	ids = map[string]entry{}
	mu.Unlock()
	return func() {
		mu.Lock()
		fetch = previous
		ids = map[string]entry{}
		mu.Unlock()
	}
}
