package playback

import (
	"context"

	"github.com/hashicorp/go-hclog"

	"github.com/gerrittrigger/trigger/config"
)

type Playback interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context) error
}

type Config struct {
	Config config.Config
	Logger hclog.Logger
}

type playback struct {
	cfg *Config
}

func New(_ context.Context, cfg *Config) Playback {
	return &playback{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (p *playback) Init(_ context.Context) error {
	p.cfg.Logger.Debug("playback: Init")

	return nil
}

func (p *playback) Deinit(_ context.Context) error {
	p.cfg.Logger.Debug("playback: Deinit")

	return nil
}

func (p *playback) Run(_ context.Context) error {
	p.cfg.Logger.Debug("playback: Run")

	return nil
}
