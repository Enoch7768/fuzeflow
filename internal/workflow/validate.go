package workflow

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,62}$`)

func Validate(definition Definition) error {
	if definition.Version != CurrentVersion {
		return fmt.Errorf("unsupported workflow version %d", definition.Version)
	}
	if !namePattern.MatchString(definition.Name) {
		return fmt.Errorf("invalid workflow name %q", definition.Name)
	}
	if len(definition.Jobs) == 0 {
		return fmt.Errorf("workflow must define at least one job")
	}

	for name, job := range definition.Jobs {
		if !namePattern.MatchString(name) {
			return fmt.Errorf("invalid job name %q", name)
		}
		if job.RunsOn != "" && !namePattern.MatchString(job.RunsOn) {
			return fmt.Errorf("invalid runs_on value %q for job %q", job.RunsOn, name)
		}
		if job.Timeout != "" {
			if _, err := time.ParseDuration(job.Timeout); err != nil || strings.HasPrefix(job.Timeout, "-") {
				return fmt.Errorf("invalid timeout for job %q", name)
			}
		}
		if len(job.Steps) == 0 {
			return fmt.Errorf("job %q must define at least one step", name)
		}
		for i, step := range job.Steps {
			if strings.TrimSpace(step.Run) == "" && strings.TrimSpace(step.Uses) == "" {
				return fmt.Errorf("job %q step %d must define run or uses", name, i+1)
			}
			if strings.TrimSpace(step.Run) != "" && strings.TrimSpace(step.Uses) != "" {
				return fmt.Errorf("job %q step %d cannot define both run and uses", name, i+1)
			}
			if step.Timeout != "" {
				if _, err := time.ParseDuration(step.Timeout); err != nil || strings.HasPrefix(step.Timeout, "-") {
					return fmt.Errorf("invalid timeout for job %q step %d", name, i+1)
				}
			}
		}
		for _, dependency := range job.Needs {
			if _, ok := definition.Jobs[dependency]; !ok {
				return fmt.Errorf("job %q depends on unknown job %q", name, dependency)
			}
			if dependency == name {
				return fmt.Errorf("job %q cannot depend on itself", name)
			}
		}
	}

	return nil
}
