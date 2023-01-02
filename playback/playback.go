package playback

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/events"
)

const (
	fileMode = 0644
	fileName = "events-base64.playback"

	queryLayout   = "2006-01-02 15:04:05"
	queryLocation = "Local"
)

type Playback interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Load(context.Context) ([]string, error)
	Store(context.Context, string) error
}

type Config struct {
	Config config.Config
	Logger hclog.Logger
}

type playback struct {
	cfg *Config
}

type httpResult struct {
	EventBase64    string `json:"eventBase64"`
	EventCreatedOn int64  `json:"eventCreatedOn"`
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

func (p *playback) Load(ctx context.Context) ([]string, error) {
	p.cfg.Logger.Debug("trigger: Load")

	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}

	event, err := p.loadCache(ctx, fileName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load cache")
	}

	query, err := p.formatQuery(ctx, &event)
	if err != nil {
		return nil, errors.Wrap(err, "failed to format query")
	}

	b, err := p.queryEvent(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query event")
	}

	_events, err := p.buildEvent(ctx, b)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build event")
	}

	return _events, nil
}

func (p *playback) Store(_ context.Context, event string) error {
	p.cfg.Logger.Debug("playback: Store")

	b := base64.StdEncoding.EncodeToString([]byte(event))

	return os.WriteFile(fileName, []byte(b), fileMode)
}

func (p *playback) loadCache(_ context.Context, name string) (events.Event, error) {
	p.cfg.Logger.Debug("trigger: loadCache")

	var e events.Event

	b, err := os.ReadFile(name)
	if err != nil {
		return events.Event{}, errors.Wrap(err, "failed to read file")
	}

	b, err = base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return events.Event{}, errors.Wrap(err, "failed to decode event")
	}

	if err = json.Unmarshal(b, &e); err != nil {
		return events.Event{}, errors.Wrap(err, "failed to unmarshal event")
	}
	return e, nil
}

func (p *playback) formatQuery(_ context.Context, event *events.Event) (string, error) {
	p.cfg.Logger.Debug("trigger: formatQuery")

	loc, _ := time.LoadLocation(queryLocation)
	s := time.Unix(event.EventCreatedOn+1, 0)
	u := time.Now()

	return fmt.Sprintf("since:%s until:%s", s.In(loc).Format(queryLayout), u.In(loc).Format(queryLayout)), nil
}

func (p *playback) queryEvent(_ context.Context, query string) ([]httpResult, error) {
	p.cfg.Logger.Debug("trigger: queryEvent")

	req, err := http.NewRequest(http.MethodGet, p.cfg.Config.Spec.Playback.EventsApi, http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set request")
	}

	q := req.URL.Query()
	q.Add("q", query)

	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body")
	}

	var buf []httpResult

	if err := json.Unmarshal(data, &buf); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal event")
	}

	return buf, nil
}

func (p *playback) buildEvent(_ context.Context, data []httpResult) ([]string, error) {
	p.cfg.Logger.Debug("trigger: buildEvent")

	if data == nil || len(data) == 0 {
		return []string{}, nil
	}

	event := make([]string, len(data))

	for i := range data {
		b, _ := base64.StdEncoding.DecodeString(data[i].EventBase64)
		event[i] = string(b)
	}

	return event, nil
}
