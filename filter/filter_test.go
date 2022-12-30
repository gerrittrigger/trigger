package filter

import (
	"context"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"

	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/events"
)

func initFilter() filter {
	f := filter{
		cfg: DefaultConfig(),
	}

	f.cfg.Config = config.Config{}

	f.cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Name:  "filter",
		Level: hclog.LevelFromString("INFO"),
	})

	return f
}

// nolint: funlen
func TestFilterEvents(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	var cfg []config.Event
	event := events.Event{}

	b := f.filterEvents(ctx, cfg, &event)
	assert.Equal(t, false, b)

	cfg = []config.Event{
		{
			CommitMessage:         "",
			ExcludeDrafts:         false,
			ExcludeNoCodeChange:   false,
			ExcludePrivateChanges: false,
			ExcludeTrivialRebase:  false,
			ExcludeWIPChanges:     false,
			Name:                  "",
			UploaderName:          "",
		},
	}

	event = events.Event{
		Change: events.Change{
			CommitMessage: "",
			Private:       false,
			Status:        "",
			WIP:           false,
		},
		PatchSet: events.PatchSet{
			IsDraft: false,
			Kind:    "",
			Uploader: events.Account{
				Name: "",
			},
		},
		Type: "",
		Uploader: events.Account{
			Name: "",
		},
	}

	b = f.filterEvents(ctx, cfg, &event)
	assert.Equal(t, false, b)

	cfg = []config.Event{
		{
			CommitMessage:         "",
			ExcludeDrafts:         false,
			ExcludeNoCodeChange:   false,
			ExcludePrivateChanges: false,
			ExcludeTrivialRebase:  false,
			ExcludeWIPChanges:     false,
			Name:                  events.EVENTS_PATCHSET_CREATED,
			UploaderName:          "",
		},
	}

	event = events.Event{
		Change: events.Change{
			CommitMessage: "",
			Private:       false,
			Status:        "",
			WIP:           false,
		},
		PatchSet: events.PatchSet{
			IsDraft: false,
			Kind:    "",
			Uploader: events.Account{
				Name: "",
			},
		},
		Type: events.EVENTS_PATCHSET_CREATED,
		Uploader: events.Account{
			Name: "",
		},
	}

	b = f.filterEvents(ctx, cfg, &event)
	assert.Equal(t, true, b)

	cfg = []config.Event{
		{
			CommitMessage:         "Init*",
			ExcludeDrafts:         false,
			ExcludeNoCodeChange:   false,
			ExcludePrivateChanges: false,
			ExcludeTrivialRebase:  false,
			ExcludeWIPChanges:     false,
			Name:                  events.EVENTS_PATCHSET_CREATED,
			UploaderName:          "ad*",
		},
	}

	event = events.Event{
		Change: events.Change{
			CommitMessage: "Initial commit",
			Private:       false,
			Status:        "",
			WIP:           false,
		},
		PatchSet: events.PatchSet{
			IsDraft: false,
			Kind:    "",
			Uploader: events.Account{
				Name: "admin",
			},
		},
		Type: events.EVENTS_PATCHSET_CREATED,
		Uploader: events.Account{
			Name: "admin",
		},
	}

	b = f.filterEvents(ctx, cfg, &event)
	assert.Equal(t, true, b)

	cfg = []config.Event{
		{
			CommitMessage:         "Init*",
			ExcludeDrafts:         true,
			ExcludeNoCodeChange:   true,
			ExcludePrivateChanges: true,
			ExcludeTrivialRebase:  true,
			ExcludeWIPChanges:     true,
			Name:                  events.EVENTS_PATCHSET_CREATED,
			UploaderName:          "ad*",
		},
	}

	event = events.Event{
		Change: events.Change{
			CommitMessage: "Initial commit",
			Private:       true,
			Status:        "DRAFT",
			WIP:           true,
		},
		PatchSet: events.PatchSet{
			IsDraft: true,
			Kind:    "NO_CODE_CHANGE",
			Uploader: events.Account{
				Name: "admin",
			},
		},
		Type: events.EVENTS_PATCHSET_CREATED,
		Uploader: events.Account{
			Name: "admin",
		},
	}

	b = f.filterEvents(ctx, cfg, &event)
	assert.Equal(t, false, b)
}

// nolint: funlen
func TestFilterProjects(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	var cfg []config.Project
	event := events.Event{}

	b := f.filterProjects(ctx, cfg, &event)
	assert.Equal(t, false, b)

	cfg = []config.Project{
		{
			Branches: []config.Match{
				{
					Pattern: "invalid",
					Type:    "",
				},
			},
			Repo: config.Match{
				Pattern: "",
				Type:    "invalid",
			},
		},
	}

	b = f.filterProjects(ctx, cfg, &event)
	assert.Equal(t, false, b)

	cfg = []config.Project{
		{
			Branches: []config.Match{
				{
					Pattern: "master",
					Type:    matchPlain,
				},
			},
			Repo: config.Match{
				Pattern: "test",
				Type:    matchPlain,
			},
		},
	}

	event = events.Event{
		Change: events.Change{
			Branch: "master",
		},
		Project: "invalid",
	}

	b = f.filterProjects(ctx, cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		Change: events.Change{
			Branch: "master",
		},
		Project: "test",
	}

	b = f.filterProjects(ctx, cfg, &event)
	assert.Equal(t, true, b)

	cfg = []config.Project{
		{
			Branches: []config.Match{
				{
					Pattern: "master",
					Type:    matchPlain,
				},
			},
			FilePaths: []config.Match{
				{
					Pattern: "**/test.txt",
					Type:    matchPath,
				},
			},
			ForbiddenFilePaths: []config.Match{
				{
					Pattern: "**/invalid.txt",
					Type:    matchPath,
				},
			},
			Repo: config.Match{
				Pattern: "test",
				Type:    matchPlain,
			},
			Topics: []config.Match{
				{
					Pattern: "t*",
					Type:    matchRegExp,
				},
			},
		},
	}

	event = events.Event{
		Change: events.Change{
			Branch: "master",
			Topic:  "test",
		},
		PatchSet: events.PatchSet{
			Files: []events.File{
				{
					File: "test.txt",
				},
			},
		},
		Project: "test",
	}

	b = f.filterProjects(ctx, cfg, &event)
	assert.Equal(t, true, b)
}

func TestEventCommitMessage(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Event{}
	event := events.Event{}

	b := f.eventCommitMessage(ctx, &cfg, &event)
	assert.Equal(t, true, b)

	cfg = config.Event{
		CommitMessage: "Init*",
	}

	event = events.Event{
		Change: events.Change{
			CommitMessage: "invalid",
		},
	}

	b = f.eventCommitMessage(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		Change: events.Change{
			CommitMessage: "Initial commit",
		},
	}

	b = f.eventCommitMessage(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestEventExcludeDrafts(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Event{}
	event := events.Event{}

	b := f.eventExcludeDrafts(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	cfg = config.Event{
		ExcludeDrafts: true,
	}

	event = events.Event{
		Change: events.Change{
			Status: "invalid",
		},
		PatchSet: events.PatchSet{
			IsDraft: false,
		},
	}

	b = f.eventExcludeDrafts(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		Change: events.Change{
			Status: "DRAFT",
		},
		PatchSet: events.PatchSet{
			IsDraft: false,
		},
	}

	b = f.eventExcludeDrafts(ctx, &cfg, &event)
	assert.Equal(t, true, b)

	event = events.Event{
		Change: events.Change{
			Status: "invalid",
		},
		PatchSet: events.PatchSet{
			IsDraft: true,
		},
	}

	b = f.eventExcludeDrafts(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestEventExcludeTrivialRebase(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Event{}
	event := events.Event{}

	b := f.eventExcludeTrivialRebase(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	cfg = config.Event{
		ExcludeTrivialRebase: true,
	}

	event = events.Event{
		PatchSet: events.PatchSet{
			Kind: "invalid",
		},
	}

	b = f.eventExcludeTrivialRebase(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		PatchSet: events.PatchSet{
			Kind: "TRIVIAL_REBASE",
		},
	}

	b = f.eventExcludeTrivialRebase(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestEventExcludeNoCodeChange(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Event{}
	event := events.Event{}

	b := f.eventExcludeNoCodeChange(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	cfg = config.Event{
		ExcludeNoCodeChange: true,
	}

	event = events.Event{
		PatchSet: events.PatchSet{
			Kind: "invalid",
		},
	}

	b = f.eventExcludeNoCodeChange(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		PatchSet: events.PatchSet{
			Kind: "NO_CODE_CHANGE",
		},
	}

	b = f.eventExcludeNoCodeChange(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestEventExcludePrivateChanges(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Event{}
	event := events.Event{}

	b := f.eventExcludePrivateChanges(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	cfg = config.Event{
		ExcludePrivateChanges: true,
	}

	event = events.Event{
		Change: events.Change{
			Private: false,
		},
	}

	b = f.eventExcludePrivateChanges(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		Change: events.Change{
			Private: true,
		},
	}

	b = f.eventExcludePrivateChanges(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestEventExcludeWIPChanges(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Event{}
	event := events.Event{}

	b := f.eventExcludeWIPChanges(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	cfg = config.Event{
		ExcludeWIPChanges: true,
	}

	event = events.Event{
		Change: events.Change{
			WIP: false,
		},
	}

	b = f.eventExcludeWIPChanges(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		Change: events.Change{
			WIP: true,
		},
	}

	b = f.eventExcludeWIPChanges(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestEventName(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Event{}
	event := events.Event{}

	b := f.eventName(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	cfg = config.Event{
		Name: events.EVENTS_PATCHSET_CREATED,
	}

	event = events.Event{
		Type: "invalid",
	}

	b = f.eventName(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	cfg = config.Event{
		Name: events.EVENTS_PATCHSET_CREATED,
	}

	event = events.Event{
		Type: events.EVENTS_PATCHSET_CREATED,
	}

	b = f.eventName(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestEventUploaderName(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Event{}
	event := events.Event{}

	b := f.eventUploaderName(ctx, &cfg, &event)
	assert.Equal(t, true, b)

	cfg = config.Event{
		UploaderName: "adm*",
	}

	event = events.Event{
		Uploader: events.Account{
			Name: "invalid",
		},
	}

	b = f.eventUploaderName(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		Uploader: events.Account{
			Name: "admin",
		},
	}

	b = f.eventUploaderName(ctx, &cfg, &event)
	assert.Equal(t, true, b)

	event = events.Event{
		PatchSet: events.PatchSet{
			Uploader: events.Account{
				Name: "admin",
			},
		},
	}

	b = f.eventUploaderName(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestProjectBranches(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Project{}
	event := events.Event{}

	b := f.projectBranches(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	cfg = config.Project{
		Branches: []config.Match{
			{
				Pattern: "master",
				Type:    matchPlain,
			},
		},
	}

	event = events.Event{
		Change: events.Change{
			Branch: "invalid",
		},
	}

	b = f.projectBranches(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		Change: events.Change{
			Branch: "master",
		},
	}

	b = f.projectBranches(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestProjectFilePaths(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Project{}
	event := events.Event{}

	b := f.projectFilePaths(ctx, &cfg, &event)
	assert.Equal(t, true, b)

	cfg = config.Project{
		FilePaths: []config.Match{
			{
				Pattern: "**/test.txt",
				Type:    matchPath,
			},
		},
	}

	event = events.Event{
		PatchSet: events.PatchSet{
			Files: []events.File{
				{
					File: "invalid",
				},
			},
		},
	}

	b = f.projectFilePaths(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		PatchSet: events.PatchSet{
			Files: []events.File{
				{
					File: "path/to/test.txt",
				},
			},
		},
	}

	b = f.projectFilePaths(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestProjectForbiddenFilePaths(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Project{}
	event := events.Event{}

	b := f.projectForbiddenFilePaths(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	cfg = config.Project{
		ForbiddenFilePaths: []config.Match{
			{
				Pattern: "**/test.txt",
				Type:    matchPath,
			},
		},
	}

	event = events.Event{
		PatchSet: events.PatchSet{
			Files: []events.File{
				{
					File: "invalid",
				},
			},
		},
	}

	b = f.projectForbiddenFilePaths(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		PatchSet: events.PatchSet{
			Files: []events.File{
				{
					File: "path/to/test.txt",
				},
			},
		},
	}

	b = f.projectForbiddenFilePaths(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestProjectRepo(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Project{
		Repo: config.Match{
			Pattern: "test",
			Type:    matchPlain,
		},
	}

	event := events.Event{
		Project: "invalid",
	}

	b := f.projectRepo(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		Project: "test",
	}

	b = f.projectRepo(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestProjectTopics(t *testing.T) {
	f := initFilter()
	ctx := context.Background()

	cfg := config.Project{}
	event := events.Event{}

	b := f.projectTopics(ctx, &cfg, &event)
	assert.Equal(t, true, b)

	cfg = config.Project{
		Topics: []config.Match{
			{
				Pattern: "test",
				Type:    matchPlain,
			},
			{
				Pattern: "invalid",
				Type:    matchPlain,
			},
		},
	}

	event = events.Event{
		Change: events.Change{
			Topic: "",
		},
	}

	b = f.projectTopics(ctx, &cfg, &event)
	assert.Equal(t, false, b)

	event = events.Event{
		Change: events.Change{
			Topic: "test",
		},
	}

	b = f.projectTopics(ctx, &cfg, &event)
	assert.Equal(t, true, b)
}

func TestProjectMatch(t *testing.T) {
	f := initFilter()

	m := config.Match{
		Pattern: "",
		Type:    "",
	}

	b := f.projectMatch(m, "")
	assert.Equal(t, false, b)

	m = config.Match{
		Pattern: "",
		Type:    matchPath,
	}

	b = f.projectMatch(m, "")
	assert.Equal(t, false, b)

	m = config.Match{
		Pattern: "test",
		Type:    "",
	}

	b = f.projectMatch(m, "")
	assert.Equal(t, false, b)

	m = config.Match{
		Pattern: "**/test.txt",
		Type:    matchPath,
	}

	b = f.projectMatch(m, "invalid.txt")
	assert.Equal(t, false, b)

	m = config.Match{
		Pattern: "**/test.txt",
		Type:    matchPath,
	}

	b = f.projectMatch(m, "test.txt")
	assert.Equal(t, true, b)

	m = config.Match{
		Pattern: "**/test.txt",
		Type:    matchPath,
	}

	b = f.projectMatch(m, "path/to/test.txt")
	assert.Equal(t, true, b)

	m = config.Match{
		Pattern: "test",
		Type:    matchPlain,
	}

	b = f.projectMatch(m, "test.txt")
	assert.Equal(t, false, b)

	m = config.Match{
		Pattern: "test.txt",
		Type:    matchPlain,
	}

	b = f.projectMatch(m, "test.txt")
	assert.Equal(t, true, b)

	m = config.Match{
		Pattern: "test*",
		Type:    matchRegExp,
	}

	b = f.projectMatch(m, "test")
	assert.Equal(t, true, b)

	m = config.Match{
		Pattern: "test*",
		Type:    matchRegExp,
	}

	b = f.projectMatch(m, "test.txt")
	assert.Equal(t, true, b)

	m = config.Match{
		Pattern: "test.txt",
		Type:    "invalid",
	}

	b = f.projectMatch(m, "test.txt")
	assert.Equal(t, false, b)
}
