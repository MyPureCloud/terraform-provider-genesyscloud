package util

import (
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v204/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitGetHomeDivisionIDWithConfigNilIsError(t *testing.T) {
	calls := 0
	restore := SetHomeDivisionFetcherForTesting(func(c *platformclientv2.Configuration) (string, error) {
		calls++
		return "home-default", nil
	})
	defer restore()

	id, diagErr := GetHomeDivisionIDWithConfig(nil)
	require.NotNil(t, diagErr, "a nil config must not fall back to the default config's org")
	assert.Contains(t, diagErr[0].Summary, "Failed to query home division")
	assert.Empty(t, id)
	assert.Equal(t, 0, calls)
}

func TestUnitGetHomeDivisionIDUsesDefaultConfig(t *testing.T) {
	var got *platformclientv2.Configuration
	restore := SetHomeDivisionFetcherForTesting(func(c *platformclientv2.Configuration) (string, error) {
		got = c
		return "home-default", nil
	})
	defer restore()

	id, diagErr := GetHomeDivisionID()
	require.Nil(t, diagErr)
	assert.Equal(t, "home-default", id)
	assert.Same(t, platformclientv2.GetDefaultConfiguration(), got)
}
