package cmd

import (
	"context"
	"io"
	"os"
	"os/signal"

	"github.com/alecthomas/kingpin/v2"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/connect"
	"github.com/gerrittrigger/trigger/filter"
	"github.com/gerrittrigger/trigger/playback"
	"github.com/gerrittrigger/trigger/query"
	"github.com/gerrittrigger/trigger/queue"
	"github.com/gerrittrigger/trigger/report"
	"github.com/gerrittrigger/trigger/trigger"
	"github.com/gerrittrigger/trigger/watchdog"
)

const (
	level = "INFO"
	name  = "trigger"
	num   = -1
)

var (
	app        = kingpin.New(name, "gerrit trigger").Version(config.Version + "-build-" + config.Build)
	configFile = app.Flag("config-file", "Config file (.yml)").Required().String()
	logLevel   = app.Flag("log-level", "Log level (DEBUG|INFO|WARN|ERROR)").Default(level).String()
)

func Run(ctx context.Context) error {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	logger, err := initLogger(ctx, *logLevel)
	if err != nil {
		return errors.Wrap(err, "failed to init logger")
	}

	cfg, err := initConfig(ctx, logger, *configFile)
	if err != nil {
		return errors.Wrap(err, "failed to init config")
	}

	flt, err := initFilter(ctx, logger, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to init filter")
	}

	pb, err := initPlayback(ctx, logger, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to init playback")
	}

	qy, err := initQuery(ctx, logger, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to init query")
	}

	mq, err := initQueue(ctx, logger, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to init queue")
	}

	rpt, err := initReport(ctx, logger, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to init report")
	}

	wd, err := initWatchdog(ctx, logger, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to init watchdog")
	}

	t, err := initTrigger(ctx, logger, cfg, flt, pb, qy, mq, rpt, wd)
	if err != nil {
		return errors.Wrap(err, "failed to init trigger")
	}

	if err := runTrigger(ctx, logger, t); err != nil {
		return errors.Wrap(err, "failed to run trigger")
	}

	return nil
}

func initLogger(_ context.Context, level string) (hclog.Logger, error) {
	return hclog.New(&hclog.LoggerOptions{
		Name:  name,
		Level: hclog.LevelFromString(level),
	}), nil
}

func initConfig(_ context.Context, logger hclog.Logger, name string) (*config.Config, error) {
	logger.Debug("cmd: initConfig")

	c := config.New()

	fi, err := os.Open(name)
	if err != nil {
		return c, errors.Wrap(err, "failed to open")
	}

	defer func() {
		_ = fi.Close()
	}()

	buf, _ := io.ReadAll(fi)

	if err := yaml.Unmarshal(buf, c); err != nil {
		return c, errors.Wrap(err, "failed to unmarshal")
	}

	return c, nil
}

func initConnect(ctx context.Context, logger hclog.Logger, cfg *config.Config) (connect.Rest, connect.Ssh, error) {
	logger.Debug("cmd: initConnect")

	rc := connect.DefaultRestConfig()
	if rc == nil {
		return nil, nil, errors.New("failed to config rest")
	}

	rc.Config = *cfg
	rc.Logger = logger

	sc := connect.DefaultSshConfig()
	if sc == nil {
		return nil, nil, errors.New("failed to config ssh")
	}

	sc.Config = *cfg
	sc.Logger = logger

	return connect.RestNew(ctx, rc), connect.SshNew(ctx, sc), nil
}

func initFilter(ctx context.Context, logger hclog.Logger, cfg *config.Config) (filter.Filter, error) {
	logger.Debug("cmd: initFilter")

	c := filter.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg
	c.Logger = logger

	return filter.New(ctx, c), nil
}

func initPlayback(ctx context.Context, logger hclog.Logger, cfg *config.Config) (playback.Playback, error) {
	logger.Debug("cmd: initPlayback")

	c := playback.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg
	c.Logger = logger

	return playback.New(ctx, c), nil
}

func initQuery(ctx context.Context, logger hclog.Logger, cfg *config.Config) (query.Query, error) {
	logger.Debug("cmd: initQuery")

	c := query.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg
	c.Logger = logger

	return query.New(ctx, c), nil
}

func initQueue(ctx context.Context, logger hclog.Logger, cfg *config.Config) (queue.Queue, error) {
	logger.Debug("cmd: initQueue")

	c := queue.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg
	c.Logger = logger

	return queue.New(ctx, c), nil
}

func initReport(ctx context.Context, logger hclog.Logger, cfg *config.Config) (report.Report, error) {
	logger.Debug("cmd: initReport")

	c := report.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg
	c.Logger = logger

	return report.New(ctx, c), nil
}

func initWatchdog(ctx context.Context, logger hclog.Logger, cfg *config.Config) (watchdog.Watchdog, error) {
	logger.Debug("cmd: initWatchdog")

	var err error

	c := watchdog.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg
	c.Logger = logger

	_, c.Ssh, err = initConnect(ctx, logger, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init connect")
	}

	return watchdog.New(ctx, c), nil
}

func initTrigger(ctx context.Context, logger hclog.Logger, cfg *config.Config, flt filter.Filter, pb playback.Playback, qy query.Query,
	mq queue.Queue, rpt report.Report, wd watchdog.Watchdog) (trigger.Trigger, error) {
	logger.Debug("cmd: initTrigger")

	var err error

	c := trigger.DefaultConfig()
	if c == nil {
		return nil, errors.New("failed to config")
	}

	c.Config = *cfg
	c.Filter = flt
	c.Logger = logger
	c.Playback = pb
	c.Query = qy
	c.Queue = mq
	c.Report = rpt
	c.Watchdog = wd

	_, c.Ssh, err = initConnect(ctx, logger, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init connect")
	}

	return trigger.New(ctx, c), nil
}

func runTrigger(ctx context.Context, logger hclog.Logger, t trigger.Trigger) error {
	logger.Debug("cmd: runTrigger")

	if err := t.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init")
	}

	param := make(chan map[string]string)

	go func() {
		_ = t.Run(ctx, nil, nil, param)
	}()

	_signal := make(chan os.Signal, 1)
	signal.Notify(_signal, os.Interrupt)

	go func() {
		<-_signal
		defer close(param)
		_ = t.Deinit(ctx)
	}()

	for item := range param {
		logger.Info("cmd: runTrigger", item)
	}

	return nil
}
