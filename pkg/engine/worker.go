package engine

import (
	"flag"
	"sync"

	"github.com/go-kit/log"
	"github.com/grafana/dskit/services"
	"github.com/thanos-io/objstore"

	"github.com/grafana/loki/v3/pkg/engine/internal/worker"
)

// WorkerConfig represents the configuration for the [Worker].
type WorkerConfig struct {
	WorkerThreads int `yaml:"worker_threads" category:"experimental"`
}

func (cfg *WorkerConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.IntVar(&cfg.WorkerThreads, prefix+"worker-threads", 0, "Experimental: Number of worker threads to spawn. Each worker thread runs one task at a time. 0 means to use GOMAXPROCS value.")
}

// Worker requests tasks from a [Scheduler] and executes them. Task results are
// sent to other [Worker] instances or back to the [Scheduler].
//
// Use [Worker.Service] to manage the lifecycle of the Worker.
type Worker struct {
	Logger log.Logger      // Logger for optional log messages.
	Bucket objstore.Bucket // Bucket to read stored data from.

	Config   WorkerConfig   // Configuration for the worker.
	Executor ExecutorConfig // Configuration for task execution.

	// Local scheduler to connect to. If LocalScheduler is nil, the worker can
	// still connect to remote schedulers.
	LocalScheduler *Scheduler

	// Our public API is a lightweight wrapper around the internal API.
	inner *worker.Worker

	initOnce  sync.Once
	initError error
}

// Service returns the service used to manage the lifecycle of the Worker.
func (w *Worker) Service() (services.Service, error) {
	if err := w.tryInit(); err != nil {
		return nil, err
	}
	return w.inner.Service(), nil
}

// tryInit initializes the Engine if necessary. tryInit will only perform
// initialization once; if initialization fails, it returns with the same error
// each time.
func (w *Worker) tryInit() error {
	w.initOnce.Do(func() { w.initError = w.init() })
	return w.initError
}

// init runs the initialization of the Worker. It must be called through
// [Worker.tryInit] to avoid re-initializing multiple times.
func (w *Worker) init() error {
	inner, err := worker.New(worker.Config{
		Logger:         w.Logger,
		Bucket:         w.Bucket,
		LocalScheduler: w.LocalScheduler.inner,

		BatchSize:  int64(w.Executor.BatchSize),
		NumThreads: w.Config.WorkerThreads,
	})
	if err != nil {
		return err
	}

	w.inner = inner
	return nil
}
