package cmd

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/gerrittrigger/trigger/config"
)

func testInitConfig() *config.Config {
	cfg := config.New()

	fi, _ := os.Open("../test/config/config.yml")

	defer func() {
		_ = fi.Close()
	}()

	buf, _ := io.ReadAll(fi)
	_ = yaml.Unmarshal(buf, cfg)

	return cfg
}

func TestInitLogger(t *testing.T) {
	ctx := context.Background()

	logger, err := initLogger(ctx, level)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, logger)
}

func TestInitConfig(t *testing.T) {
	logger, _ := initLogger(context.Background(), level)
	ctx := context.Background()

	_, err := initConfig(ctx, logger, "invalid.yml")
	assert.NotEqual(t, nil, err)

	_, err = initConfig(ctx, logger, "../test/config/invalid.yml")
	assert.NotEqual(t, nil, err)

	_, err = initConfig(ctx, logger, "../test/config/config.yml")
	assert.Equal(t, nil, err)
}

func TestInitConnect(t *testing.T) {
	logger, _ := initLogger(context.Background(), level)
	cfg := testInitConfig()

	_, _, err := initConnect(context.Background(), logger, cfg)
	assert.Equal(t, nil, err)
}

func TestInitFilter(t *testing.T) {
	logger, _ := initLogger(context.Background(), level)
	cfg := testInitConfig()

	_, err := initFilter(context.Background(), logger, cfg)
	assert.Equal(t, nil, err)
}

func TestInitPlayback(t *testing.T) {
	logger, _ := initLogger(context.Background(), level)
	cfg := testInitConfig()

	_, err := initPlayback(context.Background(), logger, cfg)
	assert.Equal(t, nil, err)
}

func TestInitQuery(t *testing.T) {
	logger, _ := initLogger(context.Background(), level)
	cfg := testInitConfig()

	_, err := initQuery(context.Background(), logger, cfg)
	assert.Equal(t, nil, err)
}

func TestInitQueue(t *testing.T) {
	logger, _ := initLogger(context.Background(), level)
	cfg := testInitConfig()

	_, err := initQueue(context.Background(), logger, cfg)
	assert.Equal(t, nil, err)
}

func TestInitReport(t *testing.T) {
	logger, _ := initLogger(context.Background(), level)
	cfg := testInitConfig()

	_, err := initReport(context.Background(), logger, cfg)
	assert.Equal(t, nil, err)
}

func TestInitWatchdog(t *testing.T) {
	logger, _ := initLogger(context.Background(), level)
	cfg := testInitConfig()

	_, err := initWatchdog(context.Background(), logger, cfg)
	assert.Equal(t, nil, err)
}

func TestInitTrigger(t *testing.T) {
	logger, _ := initLogger(context.Background(), level)
	cfg := testInitConfig()

	_, err := initTrigger(context.Background(), logger, cfg, nil, nil, nil, nil, nil, nil)
	assert.Equal(t, nil, err)
}
