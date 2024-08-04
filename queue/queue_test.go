package queue

import (
	"context"
	"strconv"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"

	"github.com/gerrittrigger/trigger/config"
)

func initQueue() queue {
	q := queue{
		cfg:    DefaultConfig(),
		events: make(chan string),
	}

	q.cfg.Config = config.Config{}

	q.cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Name:  "queue",
		Level: hclog.LevelFromString("INFO"),
	})

	return q
}

func TestQueue(t *testing.T) {
	q := initQueue()
	ctx := context.Background()

	defer func(q *queue, ctx context.Context) {
		_ = q.Close(ctx)
	}(&q, ctx)

	done := make(chan bool, 1)

	go func() {
		for i := 0; i < 2; i++ {
			err := q.Put(ctx, strconv.Itoa(i))
			assert.Equal(t, nil, err)
		}
		done <- true
	}()

	_q, err := q.Get(ctx)
	assert.Equal(t, nil, err)

L:
	for {
		select {
		case buf := <-_q:
			assert.NotEqual(t, "", buf)
		case <-done:
			break L
		}
	}
}
