package report

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/events"
	"github.com/gerrittrigger/trigger/params"
)

const (
	port   = "29418"
	scheme = "ssh"
)

type Report interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context, *events.Event) (map[string]string, error)
}

type Config struct {
	Config config.Config
	Logger hclog.Logger
}

type report struct {
	cfg *Config
}

func New(_ context.Context, cfg *Config) Report {
	return &report{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (r *report) Init(_ context.Context) error {
	r.cfg.Logger.Debug("report: Init")

	return nil
}

func (r *report) Deinit(_ context.Context) error {
	r.cfg.Logger.Debug("report: Deinit")

	return nil
}

func (r *report) Run(ctx context.Context, event *events.Event) (map[string]string, error) {
	r.cfg.Logger.Debug("report: Run")

	var err error
	buf := map[string]string{}

	buf, err = r.fetchChange(ctx, event, buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch change")
	}

	buf, err = r.fetchEvent(ctx, event, buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch event")
	}

	buf, err = r.fetchGeneral(ctx, buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch general")
	}

	return buf, nil
}

func (r *report) fetchChange(_ context.Context, event *events.Event, data map[string]string) (map[string]string, error) {
	r.cfg.Logger.Debug("report: fetchChange")

	data[params.PARAMS_GERRIT_BRANCH] = event.Change.Branch
	data[params.PARAMS_GERRIT_CHANGE_COMMIT_MESSAGE] = base64.StdEncoding.EncodeToString([]byte(event.Change.CommitMessage))
	data[params.PARAMS_GERRIT_CHANGE_ID] = event.Change.ID
	data[params.PARAMS_GERRIT_CHANGE_NUMBER] = strconv.Itoa(event.Change.Number)
	data[params.PARAMS_GERRIT_CHANGE_OWNER] = fmt.Sprintf(`%q <%s>`, event.Change.Owner.Name, event.Change.Owner.Email)
	data[params.PARAMS_GERRIT_CHANGE_OWNER_EMAIL] = event.Change.Owner.Email
	data[params.PARAMS_GERRIT_CHANGE_OWNER_NAME] = event.Change.Owner.Name
	data[params.PARAMS_GERRIT_CHANGE_PRIVATE_STATE] = strconv.FormatBool(event.Change.Private)
	data[params.PARAMS_GERRIT_CHANGE_SUBJECT] = event.Change.Subject
	data[params.PARAMS_GERRIT_CHANGE_URL] = event.Change.URL
	data[params.PARAMS_GERRIT_CHANGE_WIP_STATE] = strconv.FormatBool(event.Change.WIP)
	data[params.PARAMS_GERRIT_PATCHSET_NUMBER] = strconv.Itoa(event.PatchSet.Number)
	data[params.PARAMS_GERRIT_PATCHSET_REVISION] = event.PatchSet.Revision
	data[params.PARAMS_GERRIT_PATCHSET_UPLOADER] = fmt.Sprintf(`%q <%s>`, event.PatchSet.Uploader.Name, event.PatchSet.Uploader.Email)
	data[params.PARAMS_GERRIT_PATCHSET_UPLOADER_EMAIL] = event.PatchSet.Uploader.Email
	data[params.PARAMS_GERRIT_PATCHSET_UPLOADER_NAME] = event.PatchSet.Uploader.Name
	data[params.PARAMS_GERRIT_PROJECT] = event.Project
	data[params.PARAMS_GERRIT_REFSPEC] = event.PatchSet.Ref
	data[params.PARAMS_GERRIT_TOPIC] = event.Change.Topic

	return data, nil
}

func (r *report) fetchEvent(_ context.Context, event *events.Event, data map[string]string) (map[string]string, error) {
	r.cfg.Logger.Debug("report: fetchEvent")

	data[params.PARAMS_GERRIT_EVENT_TYPE] = event.Type

	return data, nil
}

func (r *report) fetchGeneral(_ context.Context, data map[string]string) (map[string]string, error) {
	r.cfg.Logger.Debug("report: fetchGeneral")

	data[params.PARAMS_GERRIT_HOST] = r.cfg.Config.Spec.Connect.Hostname
	data[params.PARAMS_GERRIT_NAME] = r.cfg.Config.Spec.Connect.Name
	data[params.PARAMS_GERRIT_PORT] = port
	data[params.PARAMS_GERRIT_SCHEME] = scheme

	return data, nil
}
