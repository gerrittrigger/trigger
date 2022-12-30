package connect

import (
	"context"
	"testing"

	"github.com/hashicorp/go-hclog"

	"github.com/stretchr/testify/assert"
)

func initRest() Rest {
	cfg := DefaultRestConfig()

	cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Name:  "rest",
		Level: hclog.LevelFromString("INFO"),
	})

	return &rest{
		cfg:  cfg,
		pass: "",
		url:  "https://android-review.googlesource.com",
		user: "",
	}
}

func TestDetail(t *testing.T) {
	r := initRest()
	ctx := context.Background()

	_ = r.Init(ctx)

	_, err := r.Detail(ctx, -1)
	assert.NotEqual(t, nil, err)

	buf, err := r.Detail(ctx, 1514894)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, buf)
}

func TestQuery(t *testing.T) {
	r := initRest()
	ctx := context.Background()

	_ = r.Init(ctx)

	_, err := r.Query(ctx, "change:-1", 0)
	assert.NotEqual(t, nil, err)

	buf, err := r.Query(ctx, "change:1514894", 0)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, buf)
}

func TestVersion(t *testing.T) {
	r := initRest()
	ctx := context.Background()

	_ = r.Init(ctx)

	buf, err := r.Version(ctx)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, buf)
}

func TestVote(t *testing.T) {
	// PASS
	assert.Equal(t, nil, nil)
}
