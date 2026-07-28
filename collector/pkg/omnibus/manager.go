package omnibus

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// Manager schedules and supervises bundled collectors.
type Manager struct {
	runner  Runner
	logger  *logrus.Entry
	running map[string]*atomic.Bool
	config  Config
	wg      sync.WaitGroup
}

// NewManager creates an omnibus manager from validated configuration.
func NewManager(cfg Config, runner Runner, logger *logrus.Entry) *Manager {
	running := make(map[string]*atomic.Bool, len(CollectorSpecs))
	for _, spec := range CollectorSpecs {
		running[spec.Name] = &atomic.Bool{}
	}
	return &Manager{
		config:  cfg,
		runner:  runner,
		logger:  logger,
		running: running,
	}
}

// Run starts configured schedules and waits for shutdown.
func (m *Manager) Run(ctx context.Context) error {
	scheduler := cron.New()
	hasSchedule := false
	for _, spec := range CollectorSpecs {
		collectorCfg := m.config.Collectors[spec.Name]
		if !collectorCfg.Enabled || collectorCfg.Schedule == "" {
			continue
		}
		hasSchedule = true
		currentSpec := spec
		if _, err := scheduler.AddFunc(collectorCfg.Schedule, func() {
			m.launch(ctx, currentSpec)
		}); err != nil {
			return fmt.Errorf("schedule collector %q: %w", spec.Name, err)
		}
		m.logger.WithFields(logrus.Fields{
			"collector": spec.Name,
			"schedule":  collectorCfg.Schedule,
		}).Info("collector scheduled")
	}

	startupDone := make(chan struct{})
	go func() {
		defer close(startupDone)
		m.runStartupCollectors(ctx)
	}()

	if !hasSchedule {
		select {
		case <-startupDone:
			m.wg.Wait()
			return nil
		case <-ctx.Done():
			<-startupDone
			m.wg.Wait()
			return nil
		}
	}

	scheduler.Start()
	select {
	case <-ctx.Done():
	case <-startupDone:
		// Scheduled collectors keep the manager alive after startup runs finish.
		<-ctx.Done()
	}

	m.logger.Info("stopping collector scheduler")
	stopCtx := scheduler.Stop()
	<-stopCtx.Done()
	<-startupDone
	m.wg.Wait()
	return nil
}

func (m *Manager) runStartupCollectors(ctx context.Context) {
	for _, spec := range CollectorSpecs {
		cfg := m.config.Collectors[spec.Name]
		if !cfg.Enabled || !cfg.RunOnStartup {
			continue
		}
		if cfg.StartupSleepSecs > 0 {
			timer := time.NewTimer(time.Duration(cfg.StartupSleepSecs) * time.Second)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
		if ctx.Err() != nil {
			return
		}
		m.execute(ctx, spec)
	}
}

func (m *Manager) launch(ctx context.Context, spec *CollectorSpec) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.execute(ctx, spec)
	}()
}

func (m *Manager) execute(ctx context.Context, spec *CollectorSpec) {
	state := m.running[spec.Name]
	if !state.CompareAndSwap(false, true) {
		m.logger.WithField("collector", spec.Name).Warn("collector run skipped because previous run is still active")
		return
	}
	defer state.Store(false)

	m.logger.WithField("collector", spec.Name).Info("collector run started")
	err := m.runner.Run(ctx, spec, m.config.Collectors[spec.Name], m.config.BinaryDir)
	if err != nil && ctx.Err() == nil {
		m.logger.WithError(err).WithField("collector", spec.Name).Error("collector run failed")
		return
	}
	m.logger.WithField("collector", spec.Name).Info("collector run finished")
}
