package trigger

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

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
	waitCount = 2
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

	if err := t.cfg.Watchdog.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init watchdog")
	}

	return nil
}

func (t *trigger) Deinit(ctx context.Context) error {
	t.cfg.Logger.Debug("trigger: Deinit")

	_ = t.cfg.Watchdog.Deinit(ctx)
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

	var err error
	var wg sync.WaitGroup

	if t.pb {
		if err = t.playbackEvent(ctx); err != nil {
			return errors.Wrap(err, "failed to playback event")
		}
	}

	buf := make(chan string)

	go func(c context.Context, b chan string) {
		t.fetchEvent(c, b)
	}(ctx, buf)

	wg.Add(waitCount)

	go func(c context.Context, b chan string) {
		defer wg.Done()
		for item := range b {
			err = t.cfg.Queue.Put(c, item)
			if err != nil {
				return
			}
		}
	}(ctx, buf)

	if err != nil {
		return errors.Wrap(err, "failed to put queue")
	}

	if _events == nil || len(_events) == 0 {
		_events = t.cfg.Config.Spec.Trigger.Events
	}

	if projects == nil || len(projects) == 0 {
		projects = t.cfg.Config.Spec.Trigger.Projects
	}

	go func(ctx context.Context, _events []config.Event, projects []config.Project, param chan map[string]string) {
		defer wg.Done()
		err = t.postReport(ctx, _events, projects, param)
		if err != nil {
			return
		}
	}(ctx, _events, projects, param)

	wg.Wait()

	return err
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

func (t *trigger) fetchEvent(ctx context.Context, event chan string) {
	t.cfg.Logger.Debug("trigger: fetchEvent")

	reconn := make(chan bool, 1)
	start := make(chan bool, 1)

	_ = t.cfg.Ssh.Start(ctx, "stream-events", event)

	go func(ctx context.Context, reconn, start chan bool) {
		_ = t.cfg.Watchdog.Run(ctx, t.cfg.Ssh, reconn, start)
	}(ctx, reconn, start)

	for {
		select {
		case <-reconn:
			if err := t.cfg.Ssh.Reconnect(ctx); err == nil {
				start <- true
			}
		case <-start:
			_ = t.cfg.Ssh.Start(ctx, "stream-events", event)
		}
	}
}

func (t *trigger) postReport(ctx context.Context, _events []config.Event, projects []config.Project, param chan map[string]string) error {
	t.cfg.Logger.Debug("trigger: postReport")

	var b map[string]string
	var err error
	var m bool
	var r chan string

	r, err = t.cfg.Queue.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get queue")
	}

	for item := range r {
		e := events.Event{}
		if err = json.Unmarshal([]byte(item), &e); err != nil {
			break
		}
		if err = t.cfg.Query.Run(ctx, _events, projects, &e, t.cfg.Ssh); err != nil {
			break
		}
		m, err = t.cfg.Filter.Run(ctx, _events, projects, &e)
		if err != nil {
			break
		}
		if m {
			b, err = t.cfg.Report.Run(ctx, &e)
			if err != nil {
				break
			}
			param <- b
		}
		if t.pb {
			if err = t.cfg.Playback.Store(ctx, item); err != nil {
				break
			}
		}
	}

	return err
}
