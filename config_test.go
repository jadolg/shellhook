package main

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfigurationLoadsCorrectly(t *testing.T) {
	c, err := getConfig("config.yaml")
	require.NoError(t, err)
	scriptUUID, err := uuid.Parse("5e5adb92-0d04-11ee-97cf-4b6c30e50f6a")
	require.NoError(t, err)
	assert.Equal(t, "KXjk9waX9fqRLQ4t8sQf5IK94e2u1CXr8X4MscDc", c.DefaultToken)
	assert.Len(t, c.Scripts, 3)
	assert.Equal(t, scriptUUID, c.Scripts[0].ID)
	assert.Equal(t, "./scripts/success.sh", c.Scripts[0].Path)
	assert.False(t, c.Scripts[0].Concurrent)
	assert.True(t, c.Scripts[1].Concurrent)
	assert.Equal(t, "YT9U08gqQ8yxa0Sk3PnDk6jpWu31bCyqa5SRQVFV8", c.Scripts[1].Token)
	assert.Equal(t, "echo \"Hello, world!\"\n", c.Scripts[2].Inline)
}

func TestConfigurationFailsOnInvalidScript(t *testing.T) {
	_, err := getConfig("bad_config.yaml")
	require.Error(t, err)
}
