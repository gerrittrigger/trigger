package watchdog

import (
	"context"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"

	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/connect"
)

func initWatchdog() watchdog {
	w := watchdog{
		cfg:  DefaultConfig(),
		done: make(chan bool, 1),
	}

	w.cfg.Config = config.Config{}

	w.cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Name:  "watchdog",
		Level: hclog.LevelFromString("INFO"),
	})

	_ssh := connect.DefaultSshConfig()
	w.cfg.Ssh = connect.SshNew(context.Background(), _ssh)

	return w
}

func TestRun(t *testing.T) {
	w := initWatchdog()
	ctx := context.Background()

	conn := make(chan bool, 1)
	done := make(chan bool, 1)

	w.cfg.Config.Spec.Watchdog.PeriodSeconds = 0

	err := w.Start(ctx, conn, done)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, <-done)

	w.cfg.Config.Spec.Watchdog.PeriodSeconds = 2

	err = w.Start(ctx, conn, done)
	assert.Equal(t, nil, err)

	err = w.Stop(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, <-done)
}
