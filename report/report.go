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

	data[params.ParamsGerritBranch] = event.Change.Branch
	data[params.ParamsGerritChangeCommitMessage] = base64.StdEncoding.EncodeToString([]byte(event.Change.CommitMessage))
	data[params.ParamsGerritChangeId] = event.Change.ID
	data[params.ParamsGerritChangeNumber] = strconv.Itoa(event.Change.Number)
	data[params.ParamsGerritChangeOwner] = fmt.Sprintf(`%q <%s>`, event.Change.Owner.Name, event.Change.Owner.Email)
	data[params.ParamsGerritChangeOwnerEmail] = event.Change.Owner.Email
	data[params.ParamsGerritChangeOwnerName] = event.Change.Owner.Name
	data[params.ParamsGerritChangePrivateState] = strconv.FormatBool(event.Change.Private)
	data[params.ParamsGerritChangeSubject] = event.Change.Subject
	data[params.ParamsGerritChangeUrl] = event.Change.URL
	data[params.ParamsGerritChangeWipState] = strconv.FormatBool(event.Change.WIP)
	data[params.ParamsGerritPatchsetNumber] = strconv.Itoa(event.PatchSet.Number)
	data[params.ParamsGerritPatchsetRevision] = event.PatchSet.Revision
	data[params.ParamsGerritPatchsetUploader] = fmt.Sprintf(`%q <%s>`, event.PatchSet.Uploader.Name, event.PatchSet.Uploader.Email)
	data[params.ParamsGerritPatchsetUploaderEmail] = event.PatchSet.Uploader.Email
	data[params.ParamsGerritPatchsetUploaderName] = event.PatchSet.Uploader.Name
	data[params.ParamsGerritProject] = event.Project
	data[params.ParamsGerritRefspec] = event.PatchSet.Ref
	data[params.ParamsGerritTopic] = event.Change.Topic

	return data, nil
}

func (r *report) fetchEvent(_ context.Context, event *events.Event, data map[string]string) (map[string]string, error) {
	r.cfg.Logger.Debug("report: fetchEvent")

	data[params.ParamsGerritEventType] = event.Type

	return data, nil
}

func (r *report) fetchGeneral(_ context.Context, data map[string]string) (map[string]string, error) {
	r.cfg.Logger.Debug("report: fetchGeneral")

	data[params.ParamsGerritHost] = r.cfg.Config.Spec.Connect.Hostname
	data[params.ParamsGerritName] = r.cfg.Config.Spec.Connect.Name
	data[params.ParamsGerritPort] = port
	data[params.ParamsGerritScheme] = scheme

	return data, nil
}
