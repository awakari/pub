package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	os.Setenv("LOG_LEVEL", "4")
	os.Setenv("API_PORT", "56789")
	os.Setenv("API_WRITER_INTERNAL_VALUE", "-12345")
	cfg, err := NewConfigFromEnv()
	assert.Nil(t, err)
	assert.Equal(t, uint16(8080), cfg.Api.Http.Port)
	assert.Equal(t, 4, cfg.Log.Level)
}
