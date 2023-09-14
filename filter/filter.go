package filter

import (
	"context"
	"regexp"
	"strings"

	"github.com/hashicorp/go-hclog"

	"github.com/gerrittrigger/go-antpath/antpath"
	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/events"
)

const (
	matchPath   = "path"
	matchPlain  = "plain"
	matchRegExp = "regexp"
)

type Filter interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context, []config.Event, []config.Project, *events.Event) (bool, error)
}

type Config struct {
	Config config.Config
	Logger hclog.Logger
}

type filter struct {
	cfg *Config
}

func New(_ context.Context, cfg *Config) Filter {
	return &filter{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (f *filter) Init(_ context.Context) error {
	f.cfg.Logger.Debug("filter: Init")

	return nil
}

func (f *filter) Deinit(_ context.Context) error {
	f.cfg.Logger.Debug("filter: Deinit")

	return nil
}

func (f *filter) Run(ctx context.Context, _events []config.Event, projects []config.Project, event *events.Event) (bool, error) {
	f.cfg.Logger.Debug("filter: Run")

	if (_events == nil || len(_events) == 0) || (projects == nil || len(projects) == 0) {
		return false, nil
	}

	if !f.filterEvents(ctx, _events, event) || !f.filterProjects(ctx, projects, event) {
		return false, nil
	}

	return true, nil
}

func (f *filter) filterEvents(ctx context.Context, cfg []config.Event, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: filterEvents")

	m := false

	for i := range cfg {
		if !f.eventName(ctx, &cfg[i], event) {
			continue
		}
		if !f.eventPatchsetExcludeDrafts(ctx, &cfg[i], event) &&
			!f.eventPatchsetExcludeNoCodeChange(ctx, &cfg[i], event) &&
			!f.eventPatchsetExcludePrivateChanges(ctx, &cfg[i], event) &&
			!f.eventPatchsetExcludeTrivialRebase(ctx, &cfg[i], event) &&
			!f.eventPatchsetExcludeWIPChanges(ctx, &cfg[i], event) &&
			f.eventCommitMessage(ctx, &cfg[i], event) &&
			f.eventUploaderName(ctx, &cfg[i], event) {
			m = true
			break
		}
	}

	return m
}

func (f *filter) filterProjects(ctx context.Context, cfg []config.Project, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: filterProjects")

	m := false

	for i := range cfg {
		if !f.projectRepo(ctx, &cfg[i], event) || !f.projectBranches(ctx, &cfg[i], event) {
			continue
		}
		if f.projectFilePaths(ctx, &cfg[i], event) &&
			!f.projectForbiddenFilePaths(ctx, &cfg[i], event) &&
			f.projectTopics(ctx, &cfg[i], event) {
			m = true
			break
		}
	}

	return m
}

func (f *filter) eventCommitMessage(_ context.Context, cfg *config.Event, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: eventCommitMessage")

	if cfg.CommitMessage == "" {
		return true
	}

	m, _ := regexp.MatchString(cfg.CommitMessage, event.Change.CommitMessage)

	return m
}

func (f *filter) eventPatchsetExcludeDrafts(_ context.Context, cfg *config.Event, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: eventPatchsetExcludeDrafts")

	if event.Change.Status != "DRAFT" && !event.PatchSet.IsDraft {
		return false
	}

	return cfg.PatchsetCreated.ExcludeDrafts
}

func (f *filter) eventPatchsetExcludeNoCodeChange(_ context.Context, cfg *config.Event, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: eventPatchsetExcludeNoCodeChange")

	if event.PatchSet.Kind != "NO_CODE_CHANGE" {
		return false
	}

	return cfg.PatchsetCreated.ExcludeNoCodeChange
}

func (f *filter) eventPatchsetExcludePrivateChanges(_ context.Context, cfg *config.Event, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: eventPatchsetExcludePrivateChanges")

	if !event.Change.Private {
		return false
	}

	return cfg.PatchsetCreated.ExcludePrivateChanges
}

func (f *filter) eventPatchsetExcludeTrivialRebase(_ context.Context, cfg *config.Event, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: eventPatchsetExcludeTrivialRebase")

	if event.PatchSet.Kind != "TRIVIAL_REBASE" {
		return false
	}

	return cfg.PatchsetCreated.ExcludeTrivialRebase
}

func (f *filter) eventPatchsetExcludeWIPChanges(_ context.Context, cfg *config.Event, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: eventPatchsetExcludeWIPChanges")

	if !event.Change.WIP {
		return false
	}

	return cfg.PatchsetCreated.ExcludeWIPChanges
}

func (f *filter) eventName(_ context.Context, cfg *config.Event, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: eventName")

	if cfg.Name == "" {
		return false
	}

	return cfg.Name == event.Type
}

func (f *filter) eventUploaderName(_ context.Context, cfg *config.Event, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: eventUploaderName")

	if cfg.UploaderName == "" {
		return true
	}

	if m, _ := regexp.MatchString(cfg.UploaderName, event.Uploader.Name); m {
		return true
	}

	if m, _ := regexp.MatchString(cfg.UploaderName, event.PatchSet.Uploader.Name); m {
		return true
	}

	return false
}

func (f *filter) projectBranches(_ context.Context, cfg *config.Project, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: projectBranches")

	if cfg.Branches == nil || len(cfg.Branches) == 0 {
		return false
	}

	m := false

	for i := range cfg.Branches {
		if f.projectMatch(cfg.Branches[i], event.Change.Branch) {
			m = true
			break
		}
	}

	return m
}

func (f *filter) projectFilePaths(_ context.Context, cfg *config.Project, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: projectFilePaths")

	helper := func(match config.Match, files []events.File) bool {
		m := false
		for _, item := range files {
			if f.projectMatch(match, item.File) {
				m = true
				break
			}
		}
		return m
	}

	if cfg.FilePaths == nil || len(cfg.FilePaths) == 0 {
		return true
	}

	m := false

	for i := range cfg.FilePaths {
		if helper(cfg.FilePaths[i], event.PatchSet.Files) {
			m = true
			break
		}
	}

	return m
}

func (f *filter) projectForbiddenFilePaths(_ context.Context, cfg *config.Project, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: projectForbiddenFilePaths")

	helper := func(match config.Match, files []events.File) bool {
		m := false
		for _, item := range files {
			if f.projectMatch(match, item.File) {
				m = true
				break
			}
		}
		return m
	}

	if cfg.ForbiddenFilePaths == nil || len(cfg.ForbiddenFilePaths) == 0 {
		return false
	}

	m := false

	for i := range cfg.ForbiddenFilePaths {
		if helper(cfg.ForbiddenFilePaths[i], event.PatchSet.Files) {
			m = true
			break
		}
	}

	return m
}

func (f *filter) projectRepo(_ context.Context, cfg *config.Project, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: projectRepo")

	return f.projectMatch(cfg.Repo, event.Project)
}

func (f *filter) projectTopics(_ context.Context, cfg *config.Project, event *events.Event) bool {
	f.cfg.Logger.Debug("filter: projectTopics")

	if cfg.Topics == nil || len(cfg.Topics) == 0 {
		return true
	}

	m := false

	for i := range cfg.Topics {
		if f.projectMatch(cfg.Topics[i], event.Change.Topic) {
			m = true
			break
		}
	}

	return m
}

func (f *filter) projectMatch(match config.Match, data string) bool {
	f.cfg.Logger.Debug("filter: projectMatch")

	if match.Pattern == "" || match.Type == "" {
		return false
	}

	m := false

	if strings.EqualFold(match.Type, matchPath) {
		a := antpath.New()
		m = a.Match(match.Pattern, data)
	} else if strings.EqualFold(match.Type, matchPlain) {
		if match.Pattern == data {
			m = true
		}
	} else if strings.EqualFold(match.Type, matchRegExp) {
		m, _ = regexp.MatchString(match.Pattern, data)
	} else {
		m = false
	}

	return m
}
