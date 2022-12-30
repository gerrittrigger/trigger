package query

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/connect"
	"github.com/gerrittrigger/trigger/events"
)

const (
	queryCount = 2
)

type Query interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context, []config.Event, []config.Project, *events.Event, connect.Ssh) error
}

type Config struct {
	Config config.Config
	Logger hclog.Logger
}

type query struct {
	cfg *Config
}

func New(_ context.Context, cfg *Config) Query {
	return &query{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (q *query) Init(ctx context.Context) error {
	q.cfg.Logger.Debug("query: Init")

	return nil
}

func (q *query) Deinit(ctx context.Context) error {
	q.cfg.Logger.Debug("query: Deinit")

	return nil
}

func (q *query) Run(ctx context.Context, _ []config.Event, projects []config.Project, event *events.Event, ssh connect.Ssh) error {
	q.cfg.Logger.Debug("query: Run")

	if err := q.filePaths(ctx, projects, event, ssh); err != nil {
		return errors.Wrap(err, "failed to query filePaths")
	}

	return nil
}

func (q *query) filePaths(ctx context.Context, projects []config.Project, event *events.Event, ssh connect.Ssh) error {
	q.cfg.Logger.Debug("query: filePaths")

	m := false

	for i := range projects {
		if (projects[i].FilePaths != nil && len(projects[i].FilePaths) != 0) ||
			(projects[i].ForbiddenFilePaths != nil && len(projects[i].ForbiddenFilePaths) != 0) {
			m = true
			break
		}
	}

	if !m {
		return nil
	}

	b, err := q.query(ctx, event, ssh)
	if err != nil {
		return errors.Wrap(err, "failed to query")
	}

	if b == "" {
		return nil
	}

	p, err := q.parse(ctx, b)
	if err != nil {
		return errors.Wrap(err, "failed to parse")
	}

	event.PatchSet = p

	return nil
}

func (q *query) query(ctx context.Context, event *events.Event, ssh connect.Ssh) (string, error) {
	q.cfg.Logger.Debug("query: query")

	var param string

	if event.PatchSet.Revision != "" {
		param = fmt.Sprintf("project:%s commit:%s", event.Project, event.PatchSet.Revision)
	} else if event.Change.Number > 0 {
		param = fmt.Sprintf("project:%s change:%d", event.Project, event.Change.Number)
	} else {
		param = ""
	}

	if param == "" {
		return "", nil
	}

	buf, err := ssh.Run(ctx, fmt.Sprintf("query --current-patch-set --files --format=JSON limit:1 %s", param))
	if err != nil {
		return "", errors.Wrap(err, "failed to run")
	}

	return buf, nil
}

func (q *query) parse(_ context.Context, data string) (events.PatchSet, error) {
	q.cfg.Logger.Debug("query: parse")

	d := strings.Split(strings.Trim(data, "\n"), "\n")
	if len(d) != queryCount {
		return events.PatchSet{}, errors.New("invalid count")
	}

	b := map[string]any{}

	if err := json.Unmarshal([]byte(d[0]), &b); err != nil {
		return events.PatchSet{}, errors.Wrap(err, "failed to unmarshal")
	}

	_, ok := b["currentPatchSet"]
	if !ok {
		return events.PatchSet{}, errors.New("invalid patchset")
	}

	p, err := json.Marshal(b["currentPatchSet"])
	if err != nil {
		return events.PatchSet{}, errors.New("failed to marshal")
	}

	r := events.PatchSet{}

	if err := json.Unmarshal(p, &r); err != nil {
		return events.PatchSet{}, errors.Wrap(err, "failed to unmarshal")
	}

	return r, nil
}
