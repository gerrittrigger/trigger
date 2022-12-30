package watchdog

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/connect"
)

const (
	prefix = "gerrit version"
)

type Watchdog interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context, connect.Ssh, chan bool, chan bool) error
}

type Config struct {
	Config config.Config
	Logger hclog.Logger
}

type watchdog struct {
	cfg *Config
}

func New(_ context.Context, cfg *Config) Watchdog {
	return &watchdog{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (w *watchdog) Init(_ context.Context) error {
	w.cfg.Logger.Debug("watchdog: Init")

	return nil
}

func (w *watchdog) Deinit(_ context.Context) error {
	w.cfg.Logger.Debug("watchdog: Deinit")

	return nil
}

func (w *watchdog) Run(ctx context.Context, ssh connect.Ssh, reconn, start chan bool) error {
	w.cfg.Logger.Debug("watchdog: Run")

	p := time.Duration(w.cfg.Config.Spec.Watchdog.PeriodSeconds)
	t := time.Duration(w.cfg.Config.Spec.Watchdog.TimeoutSeconds)

	if p == 0 || t == 0 {
		start <- true
		return nil
	}

	ticker := time.NewTicker(p * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.check(ctx, ssh); err != nil {
				time.Sleep(t)
				reconn <- true
			}
		}
	}
}

func (w *watchdog) check(ctx context.Context, ssh connect.Ssh) error {
	w.cfg.Logger.Debug("watchdog: check")

	b, err := ssh.Run(ctx, "version")
	if err != nil {
		return errors.Wrap(err, "failed to run ssh")
	}

	if !strings.HasPrefix(b, prefix) {
		return errors.New("invalid version")
	}

	return nil
}
