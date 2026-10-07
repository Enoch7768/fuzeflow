package workflow

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrInvalidTransition = errors.New("invalid workflow state transition")

type Execution struct {
	mu        sync.RWMutex
	state     State
	updatedAt time.Time
}

func NewExecution(now time.Time) *Execution {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return &Execution{state: StatePending, updatedAt: now}
}

func (e *Execution) State() (State, time.Time) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state, e.updatedAt
}

func (e *Execution) Transition(next State, now time.Time) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !validTransition(e.state, next) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, e.state, next)
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	e.state = next
	e.updatedAt = now
	return nil
}

func (e *Execution) Cancel(now time.Time) error {
	return e.Transition(StateCancelled, now)
}

func (e *Execution) Timeout(now time.Time) error {
	return e.Transition(StateTimedOut, now)
}

func validTransition(current, next State) bool {
	switch current {
	case StatePending:
		return next == StateQueued || next == StateCancelled
	case StateQueued:
		return next == StateRunning || next == StateCancelled || next == StateTimedOut
	case StateRunning:
		return next == StateSucceeded || next == StateFailed || next == StateCancelled || next == StateTimedOut
	default:
		return false
	}
}
