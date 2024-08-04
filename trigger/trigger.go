package trigger

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/connect"
	"github.com/gerrittrigger/trigger/events"
	"github.com/gerrittrigger/trigger/filter"
	"github.com/gerrittrigger/trigger/playback"
	"github.com/gerrittrigger/trigger/query"
	"github.com/gerrittrigger/trigger/queue"
	"github.com/gerrittrigger/trigger/report"
	"github.com/gerrittrigger/trigger/watchdog"
)

const (
	num = -1
)

type Trigger interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context, []config.Event, []config.Project, chan map[string]string) error
}

type Config struct {
	Config   config.Config
	Filter   filter.Filter
	Logger   hclog.Logger
	Playback playback.Playback
	Query    query.Query
	Queue    queue.Queue
	Report   report.Report
	Ssh      connect.Ssh
	Watchdog watchdog.Watchdog
}

type trigger struct {
	cfg *Config
	pb  bool
}

func New(_ context.Context, cfg *Config) Trigger {
	return &trigger{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (t *trigger) Init(ctx context.Context) error {
	t.cfg.Logger.Debug("trigger: Init")

	if err := t.cfg.Filter.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init filter")
	}

	if err := t.cfg.Playback.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init playback")
	}

	if strings.TrimSpace(t.cfg.Config.Spec.Playback.EventsApi) != "" {
		t.pb = true
	} else {
		t.pb = false
	}

	if err := t.cfg.Query.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init query")
	}

	if err := t.cfg.Queue.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init queue")
	}

	if err := t.cfg.Report.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init report")
	}

	if err := t.cfg.Ssh.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init ssh")
	}

	return nil
}

func (t *trigger) Deinit(ctx context.Context) error {
	t.cfg.Logger.Debug("trigger: Deinit")

	_ = t.cfg.Ssh.Deinit(ctx)
	_ = t.cfg.Report.Deinit(ctx)
	_ = t.cfg.Queue.Deinit(ctx)
	_ = t.cfg.Query.Deinit(ctx)
	_ = t.cfg.Playback.Deinit(ctx)
	_ = t.cfg.Filter.Deinit(ctx)

	return nil
}

func (t *trigger) Run(ctx context.Context, _events []config.Event, projects []config.Project, param chan map[string]string) error {
	t.cfg.Logger.Debug("trigger: Run")

	if t.pb {
		if err := t.playbackEvent(ctx); err != nil {
			return errors.Wrap(err, "failed to playback event")
		}
	}

	t.fetchEvent(ctx)

	if _events == nil || len(_events) == 0 {
		_events = t.cfg.Config.Spec.Trigger.Events
	}

	if projects == nil || len(projects) == 0 {
		projects = t.cfg.Config.Spec.Trigger.Projects
	}

	if err := t.postReport(ctx, _events, projects, param); err != nil {
		return errors.Wrap(err, "failed to post report")
	}

	return nil
}

func (t *trigger) playbackEvent(ctx context.Context) error {
	t.cfg.Logger.Debug("trigger: playbackEvent")

	var err error

	b, err := t.cfg.Playback.Load(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load event")
	}

	go func(c context.Context, d []string) {
		for i := range d {
			err = t.cfg.Queue.Put(c, d[i])
			if err != nil {
				return
			}
		}
	}(ctx, b)

	return err
}

func (t *trigger) fetchEvent(ctx context.Context) {
	t.cfg.Logger.Debug("trigger: fetchEvent")

	_ = t.cfg.Ssh.Start(ctx, "stream-events", t.cfg.Queue)
}

func (t *trigger) postReport(ctx context.Context, _events []config.Event, projects []config.Project, param chan map[string]string) error {
	t.cfg.Logger.Debug("trigger: postReport")

	helper := func(data string) error {
		e := events.Event{}
		if err := json.Unmarshal([]byte(data), &e); err != nil {
			return errors.Wrap(err, "failed to unmarshal json")
		}
		if err := t.cfg.Query.Run(ctx, _events, projects, &e, t.cfg.Ssh); err != nil {
			return errors.Wrap(err, "failed to run query")
		}
		m, err := t.cfg.Filter.Run(ctx, _events, projects, &e)
		if err != nil {
			return errors.Wrap(err, "failed to run filter")
		}
		if m {
			b, err := t.cfg.Report.Run(ctx, &e)
			if err != nil {
				return errors.Wrap(err, "failed to run report")
			}
			param <- b
		}
		if t.pb {
			if err := t.cfg.Playback.Store(ctx, data); err != nil {
				return errors.Wrap(err, "failed to run store")
			}
		}
		return nil
	}

	defer func() {
		_ = t.cfg.Queue.Close(ctx)
		close(param)
	}()

	r, err := t.cfg.Queue.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get queue")
	}

	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(num)

	g.Go(func() error {
		for {
			select {
			case buf := <-r:
				if err := helper(buf); err != nil {
					return err
				}
			case <-ctx.Done():
				return nil
			}
		}
	})

	return nil
}
