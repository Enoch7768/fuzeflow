package workflow

type Definition struct {
	Version int                `yaml:"version"`
	Name    string             `yaml:"name"`
	On      map[string]Trigger `yaml:"on"`
	Jobs    map[string]Job     `yaml:"jobs"`
}

type Trigger struct {
	Branches []string `yaml:"branches,omitempty"`
	Tags     []string `yaml:"tags,omitempty"`
}

type Job struct {
	Needs   []string `yaml:"needs,omitempty"`
	RunsOn  string   `yaml:"runs_on,omitempty"`
	Timeout string   `yaml:"timeout,omitempty"`
	Steps   []Step   `yaml:"steps"`
}

type Step struct {
	Name    string `yaml:"name,omitempty"`
	Run     string `yaml:"run,omitempty"`
	Uses    string `yaml:"uses,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
}

type Plan struct {
	WorkflowName string
	Order        []string
	Jobs         map[string]PlannedJob
}

type PlannedJob struct {
	Name         string
	Dependencies []string
}

type State string

const (
	StatePending   State = "pending"
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateSucceeded State = "succeeded"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
	StateTimedOut  State = "timed_out"
)
