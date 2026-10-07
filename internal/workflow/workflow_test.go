package workflow

import (
	"errors"
	"testing"
	"time"
)

func TestParseAndPlan(t *testing.T) {
	input := []byte(`version: 1
name: build
on:
  push:
    branches: [main]
jobs:
  test:
    runs_on: linux
    steps:
      - run: go test ./...
  package:
    needs: [test]
    runs_on: linux
    steps:
      - run: go build ./...
`)

	definition, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := PlanExecution(definition)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Order) != 2 || plan.Order[0] != "test" || plan.Order[1] != "package" {
		t.Fatalf("unexpected execution order: %#v", plan.Order)
	}
}

func TestCycleRejected(t *testing.T) {
	definition := Definition{
		Version: 1,
		Name:    "cycle",
		Jobs: map[string]Job{
			"a": {Needs: []string{"b"}, Steps: []Step{{Run: "echo a"}}},
			"b": {Needs: []string{"a"}, Steps: []Step{{Run: "echo b"}}},
		},
	}
	if _, err := PlanExecution(definition); err == nil {
		t.Fatal("expected dependency cycle to fail")
	}
}

func TestInvalidStepRejected(t *testing.T) {
	definition := Definition{
		Version: 1,
		Name:    "invalid",
		Jobs: map[string]Job{
			"test": {Steps: []Step{{Run: "echo ok", Uses: "x/y"}}},
		},
	}
	if err := Validate(definition); err == nil {
		t.Fatal("expected invalid step to fail")
	}
}

func TestExecutionLifecycle(t *testing.T) {
	now := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	execution := NewExecution(now)

	if state, _ := execution.State(); state != StatePending {
		t.Fatalf("unexpected initial state: %s", state)
	}
	if err := execution.Transition(StateRunning, now); !errors.Is(err, ErrInvalidTransition) {
		t.Fatal("expected pending to running to be rejected")
	}
	if err := execution.Transition(StateQueued, now); err != nil {
		t.Fatal(err)
	}
	if err := execution.Transition(StateRunning, now); err != nil {
		t.Fatal(err)
	}
	if err := execution.Timeout(now); err != nil {
		t.Fatal(err)
	}
	if err := execution.Cancel(now); !errors.Is(err, ErrInvalidTransition) {
		t.Fatal("expected timed out execution to remain terminal")
	}
}
