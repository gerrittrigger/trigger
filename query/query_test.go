package query

import (
	"context"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"

	"github.com/gerrittrigger/trigger/config"
)

// nolint: lll
const (
	queryData = `{"project":"test","branch":"master","number":22,"currentPatchSet":{"files":[{"file":"README.md","type":"ADDED","insertions":1,"deletions":0}]}}
{"type":"stats","rowCount":1,"runTimeMilliseconds":7,"moreChanges":false}`
)

func initQuery() query {
	q := query{
		cfg: DefaultConfig(),
	}

	q.cfg.Config = config.Config{}

	q.cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Name:  "query",
		Level: hclog.LevelFromString("INFO"),
	})

	return q
}

func TestParse(t *testing.T) {
	q := initQuery()
	ctx := context.Background()

	d := "invalid"

	_, err := q.parse(ctx, d)
	assert.NotEqual(t, nil, err)

	d = `invalid
invalid`

	_, err = q.parse(ctx, d)
	assert.NotEqual(t, nil, err)

	d = `{"key": "invalid"}
invalid`

	_, err = q.parse(ctx, d)
	assert.NotEqual(t, nil, err)

	b, err := q.parse(ctx, queryData)
	assert.Equal(t, nil, err)
	assert.Equal(t, "README.md", b.Files[0].File)
}
