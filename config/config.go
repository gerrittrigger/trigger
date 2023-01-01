package config

type Config struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	MetaData   MetaData `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type MetaData struct {
	Name string `yaml:"name"`
}

type Spec struct {
	Connect  Connect  `yaml:"connect"`
	Queue    Queue    `yaml:"queue"`
	Playback Playback `yaml:"playback"`
	Report   Report   `yaml:"report"`
	Trigger  Trigger  `yaml:"trigger"`
	Watchdog Watchdog `yaml:"watchdog"`
}

type Connect struct {
	FrontendUrl string `yaml:"frontendUrl"`
	Hostname    string `yaml:"hostname"`
	Name        string `yaml:"name"`
	Http        Http   `yaml:"http"`
	Ssh         Ssh    `yaml:"ssh"`
}

type Http struct {
	Password string `yaml:"password"`
	Username string `yaml:"username"`
}

type Ssh struct {
	Keyfile         string `yaml:"keyfile"`
	KeyfilePassword string `yaml:"keyfilePassword"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
}

type Queue struct {
}

type Playback struct {
	EventsApi string `yaml:"eventsApi"`
}

type Report struct {
}

type Trigger struct {
	Events   []Event   `yaml:"events"`
	Projects []Project `yaml:"projects"`
}

type Event struct {
	CommitMessage         string `yaml:"commitMessage"`
	ExcludeDrafts         bool   `yaml:"excludeDrafts"`
	ExcludeTrivialRebase  bool   `yaml:"excludeTrivialRebase"`
	ExcludeNoCodeChange   bool   `yaml:"excludeNoCodeChange"`
	ExcludePrivateChanges bool   `yaml:"excludePrivateChanges"`
	ExcludeWIPChanges     bool   `yaml:"excludeWIPChanges"`
	Name                  string `yaml:"name"`
	UploaderName          string `yaml:"uploaderName"`
}

type Project struct {
	Branches           []Match `yaml:"branches"`
	FilePaths          []Match `yaml:"filePaths"`
	ForbiddenFilePaths []Match `yaml:"forbiddenFilePaths"`
	Repo               Match   `yaml:"repo"`
	Topics             []Match `yaml:"topics"`
}

type Match struct {
	Pattern string `yaml:"pattern"`
	Type    string `yaml:"type"`
}

type Watchdog struct {
	PeriodSeconds  int `yaml:"periodSeconds"`
	TimeoutSeconds int `yaml:"timeoutSeconds"`
}

var (
	Build   string
	Version string
)

func New() *Config {
	return &Config{}
}
