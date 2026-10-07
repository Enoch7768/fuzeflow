package workflow

import (
	"fmt"
	"sort"
)

func PlanExecution(definition Definition) (Plan, error) {
	if err := Validate(definition); err != nil {
		return Plan{}, err
	}

	indegree := make(map[string]int, len(definition.Jobs))
	dependents := make(map[string][]string, len(definition.Jobs))
	for name, job := range definition.Jobs {
		indegree[name] = len(job.Needs)
		for _, dependency := range job.Needs {
			dependents[dependency] = append(dependents[dependency], name)
		}
	}

	ready := make([]string, 0)
	for name, degree := range indegree {
		if degree == 0 {
			ready = append(ready, name)
		}
	}
	sort.Strings(ready)

	order := make([]string, 0, len(definition.Jobs))
	for len(ready) > 0 {
		name := ready[0]
		ready = ready[1:]
		order = append(order, name)

		next := append([]string(nil), dependents[name]...)
		sort.Strings(next)
		for _, dependent := range next {
			indegree[dependent]--
			if indegree[dependent] == 0 {
				ready = append(ready, dependent)
			}
		}
		sort.Strings(ready)
	}

	if len(order) != len(definition.Jobs) {
		return Plan{}, fmt.Errorf("workflow contains a dependency cycle")
	}

	jobs := make(map[string]PlannedJob, len(definition.Jobs))
	for name, job := range definition.Jobs {
		dependencies := append([]string(nil), job.Needs...)
		sort.Strings(dependencies)
		jobs[name] = PlannedJob{Name: name, Dependencies: dependencies}
	}

	return Plan{WorkflowName: definition.Name, Order: order, Jobs: jobs}, nil
}
