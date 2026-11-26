package homedivision

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tokenConfig(basePath, token string) *platformclientv2.Configuration {
	return &platformclientv2.Configuration{BasePath: basePath, AccessToken: token}
}

func clientConfig(basePath, clientID, token string) *platformclientv2.Configuration {
	return &platformclientv2.Configuration{BasePath: basePath, ClientID: clientID, AccessToken: token}
}

func keyOf(c *platformclientv2.Configuration) string {
	key, _ := cacheKey(c)
	return key
}

// countingFetcher returns a fetcher that maps each config's cache key to an ID and counts calls.
func countingFetcher(calls *int32, ids map[string]string) Fetcher {
	return func(c *platformclientv2.Configuration) (string, error) {
		atomic.AddInt32(calls, 1)
		id, ok := ids[keyOf(c)]
		if !ok {
			return "", errors.New("unknown org")
		}
		return id, nil
	}
}

// fakeClock replaces the expiry clock for one test.
func fakeClock(t *testing.T, start time.Time) *time.Time {
	current := start
	previous := now
	now = func() time.Time { return current }
	t.Cleanup(func() { now = previous })
	return &current
}

func TestUnitHomeDivisionIDPerOrg(t *testing.T) {
	source := tokenConfig("https://api.mypurecloud.com", "source-token")
	target := tokenConfig("https://api.mypurecloud.ie", "target-token")

	var calls int32
	restore := SetFetcherForTesting(countingFetcher(&calls, map[string]string{
		keyOf(source): "source-home",
		keyOf(target): "target-home",
	}))
	defer restore()

	id, err := ID(source)
	require.Nil(t, err)
	assert.Equal(t, "source-home", id)

	id, err = ID(target)
	require.Nil(t, err)
	assert.Equal(t, "target-home", id, "target org must not get the source org's cached ID")

	// Both are now cached.
	id, _ = ID(source)
	assert.Equal(t, "source-home", id)
	id, _ = ID(target)
	assert.Equal(t, "target-home", id)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
}

func TestUnitHomeDivisionIDSameTokenSharesEntry(t *testing.T) {
	// Separate config instances carrying the same token (e.g. per-call configs that reuse a
	// cached OAuth token) must share one cache entry.
	var calls int32
	restore := SetFetcherForTesting(countingFetcher(&calls, map[string]string{
		keyOf(tokenConfig("https://api.mypurecloud.com", "token-a")): "home-a",
	}))
	defer restore()

	for i := 0; i < 5; i++ {
		id, err := ID(tokenConfig("https://api.mypurecloud.com", "token-a"))
		require.Nil(t, err)
		assert.Equal(t, "home-a", id)
	}
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

func TestUnitHomeDivisionCacheKeyPrefersClientID(t *testing.T) {
	// Client-credential configs refresh their token in place; the key must not change.
	cfg := clientConfig("https://api.mypurecloud.com", "client-a", "token-1")
	key, expires := cacheKey(cfg)
	assert.False(t, expires, "client-keyed entries must not expire")
	cfg.AccessToken = "token-2"
	assert.Equal(t, key, keyOf(cfg))

	// Same client ID in a different region is a different key.
	assert.NotEqual(t, key, keyOf(clientConfig("https://api.mypurecloud.ie", "client-a", "token-1")))

	_, expires = cacheKey(tokenConfig("https://api.mypurecloud.com", "token-1"))
	assert.True(t, expires, "token-keyed entries must expire")
}

func TestUnitHomeDivisionIDPooledClientsShareEntry(t *testing.T) {
	// A client pool authorises each config separately, so each has its own token, but they all
	// belong to one client and must share one lookup.
	var calls int32
	restore := SetFetcherForTesting(countingFetcher(&calls, map[string]string{
		keyOf(clientConfig("https://api.mypurecloud.com", "client-a", "")): "home-a",
	}))
	defer restore()

	for i := 0; i < 10; i++ {
		id, err := ID(clientConfig("https://api.mypurecloud.com", "client-a", fmt.Sprintf("token-%d", i)))
		require.Nil(t, err)
		assert.Equal(t, "home-a", id)
	}
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

func TestUnitHomeDivisionClientKeyedEntriesDoNotExpire(t *testing.T) {
	now := fakeClock(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := clientConfig("https://api.mypurecloud.com", "client-a", "token-a")

	var calls int32
	restore := SetFetcherForTesting(countingFetcher(&calls, map[string]string{keyOf(cfg): "home-a"}))
	defer restore()

	_, _ = ID(cfg)
	*now = now.Add(30 * 24 * time.Hour)
	_, _ = ID(cfg)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

func TestUnitHomeDivisionCacheKey(t *testing.T) {
	key := keyOf(tokenConfig("https://api.mypurecloud.com", "token-a"))
	assert.Equal(t, key, keyOf(tokenConfig("https://api.mypurecloud.com", "token-a")))
	assert.NotEqual(t, key, keyOf(tokenConfig("https://api.mypurecloud.com", "token-b")))
	assert.NotEqual(t, key, keyOf(tokenConfig("https://api.mypurecloud.ie", "token-a")))
	assert.NotContains(t, key, "token-a", "the raw token must not appear in the cache key")
}

func TestUnitHomeDivisionIDDoesNotCacheFailures(t *testing.T) {
	cfg := tokenConfig("https://api.mypurecloud.com", "token-a")

	var calls int32
	fail := true
	restore := SetFetcherForTesting(func(c *platformclientv2.Configuration) (string, error) {
		atomic.AddInt32(&calls, 1)
		if fail {
			return "", errors.New("401 - No authentication bearer token specified")
		}
		return "home-a", nil
	})
	defer restore()

	_, err := ID(cfg)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "401")

	fail = false
	id, err := ID(cfg)
	require.Nil(t, err, "a failed lookup must not be cached")
	assert.Equal(t, "home-a", id)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
}

func TestUnitHomeDivisionIDConcurrentColdCache(t *testing.T) {
	cfg := tokenConfig("https://api.mypurecloud.com", "token-a")

	var calls int32
	release := make(chan struct{})
	restore := SetFetcherForTesting(func(c *platformclientv2.Configuration) (string, error) {
		atomic.AddInt32(&calls, 1)
		<-release
		return "home-a", nil
	})
	defer restore()

	const workers = 500
	var wg sync.WaitGroup
	results := make([]string, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id, err := ID(cfg)
			if err == nil {
				results[i] = id
			}
		}(i)
	}
	// Give the goroutines time to pile up behind the in-flight lookup.
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "concurrent callers for one org must share one API call")
	for _, id := range results {
		assert.Equal(t, "home-a", id)
	}
}

func TestUnitHomeDivisionIDConcurrentFailureShared(t *testing.T) {
	cfg := tokenConfig("https://api.mypurecloud.com", "token-a")

	var calls int32
	release := make(chan struct{})
	restore := SetFetcherForTesting(func(c *platformclientv2.Configuration) (string, error) {
		atomic.AddInt32(&calls, 1)
		<-release
		return "", errors.New("503")
	})
	defer restore()

	const workers = 50
	var wg, started sync.WaitGroup
	var failures int32
	for i := 0; i < workers; i++ {
		wg.Add(1)
		started.Add(1)
		go func() {
			defer wg.Done()
			started.Done()
			if _, err := ID(cfg); err != nil {
				atomic.AddInt32(&failures, 1)
			}
		}()
	}
	started.Wait()
	// Every goroutine has started; give them time to join the in-flight lookup.
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	// Every caller sees the failure, and waiters that joined the in-flight call share its error
	// instead of each making their own call. Failures are not cached, so a goroutine the scheduler
	// only runs after the flight has finished starts a new one: the call count is bounded, not
	// exactly one. (Exactly-one sharing is asserted deterministically by the cold-cache test, where
	// the successful result is cached.)
	assert.Equal(t, int32(workers), atomic.LoadInt32(&failures))
	got := atomic.LoadInt32(&calls)
	assert.GreaterOrEqual(t, got, int32(1))
	assert.Less(t, got, int32(workers), "concurrent failing lookups must share in-flight calls")
}

func TestUnitHomeDivisionEntriesExpire(t *testing.T) {
	now := fakeClock(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := tokenConfig("https://api.mypurecloud.com", "token-a")

	var calls int32
	restore := SetFetcherForTesting(countingFetcher(&calls, map[string]string{keyOf(cfg): "home-a"}))
	defer restore()

	_, err := ID(cfg)
	require.Nil(t, err)

	*now = now.Add(ttl - time.Second)
	_, _ = ID(cfg)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "entry is still valid just before the TTL")

	*now = now.Add(time.Second)
	id, err := ID(cfg)
	require.Nil(t, err)
	assert.Equal(t, "home-a", id)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls), "entry is refetched once the TTL has passed")
}

func TestUnitHomeDivisionRotatedTokensAreSwept(t *testing.T) {
	// Simulates a caller that rotates its access token every 30 minutes for a day.
	now := fakeClock(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	restore := SetFetcherForTesting(func(c *platformclientv2.Configuration) (string, error) {
		return "home-a", nil
	})
	defer restore()

	// A client-keyed entry must survive every sweep.
	client := clientConfig("https://api.mypurecloud.com", "client-a", "")
	_, _ = ID(client)

	for i := range 48 {
		_, err := ID(tokenConfig("https://api.mypurecloud.com", fmt.Sprintf("token-%d", i)))
		require.Nil(t, err)
		*now = now.Add(30 * time.Minute)
	}

	mu.Lock()
	size := len(ids)
	_, clientEntryKept := ids[keyOf(client)]
	mu.Unlock()

	// With a 1h TTL and 30 minute rotation, at most 2 token entries are live at once.
	assert.LessOrEqual(t, size, 3, "expired token-keyed entries must be swept")
	assert.True(t, clientEntryKept)
}
