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
	Start(context.Context, chan bool, chan bool) error
	Stop(context.Context) error
}

type Config struct {
	Config config.Config
	Logger hclog.Logger
	Ssh    connect.Ssh
}

type watchdog struct {
	cfg  *Config
	done chan bool
}

func New(_ context.Context, cfg *Config) Watchdog {
	return &watchdog{
		cfg:  cfg,
		done: make(chan bool, 1),
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (w *watchdog) Init(ctx context.Context) error {
	w.cfg.Logger.Debug("watchdog: Init")

	if err := w.cfg.Ssh.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init ssh")
	}

	return nil
}

func (w *watchdog) Deinit(ctx context.Context) error {
	w.cfg.Logger.Debug("watchdog: Deinit")

	_ = w.cfg.Ssh.Deinit(ctx)

	return nil
}

func (w *watchdog) Start(ctx context.Context, conn, done chan bool) error {
	p := time.Duration(w.cfg.Config.Spec.Watchdog.PeriodSeconds)

	if p == 0 {
		done <- true
		return nil
	}

	ticker := time.NewTicker(p * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := w.check(ctx); err == nil {
					conn <- true
				} else {
					conn <- false
				}
			case <-w.done:
				done <- true
				return
			}
		}
	}()

	return nil
}

func (w *watchdog) Stop(_ context.Context) error {
	w.done <- true
	return nil
}

func (w *watchdog) check(ctx context.Context) error {
	b, err := w.cfg.Ssh.Run(ctx, "version")
	if err != nil {
		return errors.Wrap(err, "failed to run ssh")
	}

	if !strings.HasPrefix(b, prefix) {
		return errors.New("invalid version")
	}

	return nil
}
