package queue

import (
	"context"

	"github.com/hashicorp/go-hclog"

	"github.com/gerrittrigger/trigger/config"
)

type Queue interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Get(context.Context) (chan string, error)
	Put(context.Context, string) error
	Close(context.Context) error
}

type Config struct {
	Config config.Config
	Logger hclog.Logger
}

type queue struct {
	cfg    *Config
	events chan string
}

func New(_ context.Context, cfg *Config) Queue {
	return &queue{
		cfg:    cfg,
		events: make(chan string),
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (q *queue) Init(_ context.Context) error {
	q.cfg.Logger.Debug("queue: Init")

	return nil
}

func (q *queue) Deinit(_ context.Context) error {
	q.cfg.Logger.Debug("queue: Deinit")

	return nil
}

func (q *queue) Get(_ context.Context) (chan string, error) {
	return q.events, nil
}

func (q *queue) Put(_ context.Context, data string) error {
	q.events <- data
	return nil
}

func (q *queue) Close(_ context.Context) error {
	close(q.events)
	return nil
}
