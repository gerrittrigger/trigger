package playback

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"

	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/events"
)

var (
	eventCreatedOn = 1672567200
	eventDecode    = `{"data": "eventBase64"}`
	eventEncode    = "eyJkYXRhIjogImV2ZW50QmFzZTY0In0="
)

func initPlayback() playback {
	p := playback{
		cfg: DefaultConfig(),
	}

	p.cfg.Config = config.Config{}
	p.cfg.Config.Spec.Playback.EventsApi = "http://localhost:8081/events"

	p.cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Name:  "playback",
		Level: hclog.LevelFromString("INFO"),
	})

	return p
}

func TestStore(t *testing.T) {
	p := initPlayback()
	ctx := context.Background()

	_ = p.Init(ctx)

	err := p.Store(ctx, eventDecode)
	assert.Equal(t, nil, err)

	b, err := os.ReadFile(fileName)
	assert.Equal(t, nil, err)
	assert.Equal(t, eventEncode, string(b))

	err = p.Store(ctx, eventDecode)
	assert.Equal(t, nil, err)

	b, err = os.ReadFile(fileName)
	assert.Equal(t, nil, err)
	assert.Equal(t, eventEncode, string(b))

	_ = os.Remove(fileName)
}

func TestLoadCache(t *testing.T) {
	p := initPlayback()
	ctx := context.Background()

	_ = p.Init(ctx)

	_, err := p.loadCache(ctx, "invalid")
	assert.NotEqual(t, nil, err)

	_ = os.WriteFile(fileName, []byte(eventEncode), fileMode)

	_, err = p.loadCache(ctx, fileName)
	assert.Equal(t, nil, err)

	_ = os.Remove(fileName)
}

func TestFormatQuery(t *testing.T) {
	p := initPlayback()
	ctx := context.Background()

	_ = p.Init(ctx)

	e := events.Event{
		EventCreatedOn: int64(eventCreatedOn),
	}

	_, err := p.formatQuery(ctx, &e)
	assert.Equal(t, nil, err)
}

func TestBuildEvent(t *testing.T) {
	var b []httpResult

	p := initPlayback()
	ctx := context.Background()

	_ = p.Init(ctx)

	r, err := p.buildEvent(ctx, b)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(r))

	b = []httpResult{
		{
			EventBase64:    eventEncode,
			EventCreatedOn: int64(eventCreatedOn),
		},
	}

	r, err = p.buildEvent(ctx, b)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(r))
}
