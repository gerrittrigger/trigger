package report

import (
	"context"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"

	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/events"
)

var (
	// nolint: lll
	event = events.Event{
		Change: events.Change{
			Branch:        "master",
			CommitMessage: "SW5pdGlhbCBjb21taXQNCg0KVGhpcyByZXZlcnRzIGNvbW1pdCA1NzM0MTkxMDEyOWM4MDUwMTk5NmQ0ZmZiNTMyOWRhYWE4YjA5ZDQ2Lg0KDQpDaGFuZ2UtSWQ6IElhZGRmNGVhMGUwNmRkZjhjZjY5ZjAzNWFhN2VmNTM5ZWNjYTAwNjgw",
			ID:            "Iaddf4ea0e06ddf8cf69f035aa7ef539ecca00680",
			Number:        22,
			Owner: events.Account{
				Email: "admin@example.com",
				Name:  "admin",
			},
			Private: false,
			Subject: "Initial commit",
			URL:     "http://localhost:8080/c/test/+/22",
			WIP:     false,
			Topic:   "test",
		},
		PatchSet: events.PatchSet{
			Number:   17,
			Revision: "ba29a60664a69fb54ff342fdb2be8259a62dcc97",
			Uploader: events.Account{
				Email: "admin@example.com",
				Name:  "admin",
			},
			Ref: "refs/changes/22/22/17",
		},
		Project: "test",
		Type:    "patchset-created",
	}
)

func initReport() report {
	r := report{
		cfg: DefaultConfig(),
	}

	r.cfg.Config = config.Config{}
	r.cfg.Config.Spec.Connect.FrontendUrl = "https://android-review.googlesource.com"

	r.cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Name:  "report",
		Level: hclog.LevelFromString("INFO"),
	})

	return r
}

func TestFetchChange(t *testing.T) {
	buf := map[string]string{}

	r := initReport()
	ctx := context.Background()

	_ = r.Init(ctx)

	_, err := r.fetchChange(ctx, &event, buf)
	assert.Equal(t, nil, err)
}

func TestFetchEvent(t *testing.T) {
	buf := map[string]string{}

	r := initReport()
	ctx := context.Background()

	_ = r.Init(ctx)

	_, err := r.fetchEvent(ctx, &event, buf)
	assert.Equal(t, nil, err)
}

func TestFetchGeneral(t *testing.T) {
	buf := map[string]string{}

	r := initReport()
	ctx := context.Background()

	_ = r.Init(ctx)

	_, err := r.fetchGeneral(ctx, buf)
	assert.Equal(t, nil, err)
}
