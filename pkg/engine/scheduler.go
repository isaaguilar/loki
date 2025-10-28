package engine

import (
	"sync"

	"github.com/go-kit/log"
	"github.com/grafana/dskit/services"

	"github.com/grafana/loki/v3/pkg/engine/internal/scheduler"
)

// Scheduler is a service that can schedule tasks to connected [Worker]
// instances.
//
// Use [Scheduler.Service] to manage the lifecycle of the Scheduler.
type Scheduler struct {
	Logger log.Logger // Logger for optional log messages.

	inner *scheduler.Scheduler

	initOnce  sync.Once
	initError error
}

// Service returns the service used to manage the lifecycle of the Scheduler.
func (s *Scheduler) Service() (services.Service, error) {
	if err := s.tryInit(); err != nil {
		return nil, err
	}
	return s.inner.Service(), nil
}

// tryInit initializes the Scheduler if necessary. tryInit will only perform
// initialization once; if initialization fails, it returns with the same error
// each time.
func (s *Scheduler) tryInit() error {
	s.initOnce.Do(func() { s.initError = s.init() })
	return s.initError
}

// init runs the initialization of the Scheduler. It must be called through
// [Scheduler.tryInit] to avoid re-initializing multiple times.
func (s *Scheduler) init() error {
	inner, err := scheduler.New(scheduler.Config{
		Logger: s.Logger,
	})
	if err != nil {
		return err
	}

	s.inner = inner
	return nil
}
