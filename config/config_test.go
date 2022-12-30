package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg := New()
	assert.NotEqual(t, nil, cfg)
}
