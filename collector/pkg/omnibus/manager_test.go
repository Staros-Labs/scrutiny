package omnibus

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type runnerCall struct {
	name string
}

type fakeRunner struct {
	calls   chan runnerCall
	release chan struct{}
	errs    map[string]error
}

func (r *fakeRunner) Run(ctx context.Context, spec CollectorSpec, _ CollectorConfig, _ string) error {
	r.calls <- runnerCall{name: spec.Name}
	if r.release != nil {
		select {
		case <-r.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return r.errs[spec.Name]
}

func TestManagerRunsStartupCollectorsInStableOrder(t *testing.T) {
	runner := &fakeRunner{
		calls: make(chan runnerCall, 2),
		errs:  map[string]error{"metrics": errors.New("metrics failed")},
	}
	cfg := Config{Collectors: map[string]CollectorConfig{
		"metrics": {Enabled: true, RunOnStartup: true},
		"zfs":     {Enabled: true, RunOnStartup: true},
	}}
	manager := NewManager(cfg, runner, testLogger())

	require.NoError(t, manager.Run(context.Background()))
	assert.Equal(t, "metrics", (<-runner.calls).name)
	assert.Equal(t, "zfs", (<-runner.calls).name)
}

func TestManagerSkipsOverlappingCollectorRun(t *testing.T) {
	runner := &fakeRunner{
		calls:   make(chan runnerCall, 2),
		release: make(chan struct{}),
		errs:    map[string]error{},
	}
	cfg := Config{Collectors: map[string]CollectorConfig{
		"metrics": {Enabled: true},
	}}
	manager := NewManager(cfg, runner, testLogger())
	spec := CollectorSpecs[0]

	manager.launch(context.Background(), spec)
	require.Equal(t, "metrics", (<-runner.calls).name)
	manager.execute(context.Background(), spec)
	close(runner.release)
	manager.wg.Wait()
	assert.Empty(t, runner.calls)
}

func TestManagerStopsScheduledModeOnContextCancellation(t *testing.T) {
	runner := &fakeRunner{
		calls: make(chan runnerCall, 1),
		errs:  map[string]error{},
	}
	cfg := Config{Collectors: map[string]CollectorConfig{
		"metrics": {Enabled: true, Schedule: "0 0 * * *"},
	}}
	manager := NewManager(cfg, runner, testLogger())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- manager.Run(ctx)
	}()
	cancel()

	require.NoError(t, <-done)
}

func TestManagerWaitsForStartupCollectorCancellation(t *testing.T) {
	runner := &fakeRunner{
		calls:   make(chan runnerCall, 1),
		release: make(chan struct{}),
		errs:    map[string]error{},
	}
	cfg := Config{Collectors: map[string]CollectorConfig{
		"metrics": {Enabled: true, RunOnStartup: true},
	}}
	manager := NewManager(cfg, runner, testLogger())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- manager.Run(ctx)
	}()
	require.Equal(t, "metrics", (<-runner.calls).name)
	cancel()

	require.NoError(t, <-done)
}

func TestChildEnvironmentSuppressesManagerSchedulingVariables(t *testing.T) {
	t.Setenv("COLLECTOR_CRON_SCHEDULE", "0 0 * * *")
	t.Setenv("COLLECTOR_RUN_STARTUP", "true")
	t.Setenv("COLLECTOR_ZFS_CRON_SCHEDULE", "*/15 * * * *")
	t.Setenv("PRESERVED_VALUE", "yes")

	env := childEnvironment()
	joined := make(map[string]string, len(env))
	for _, item := range env {
		key, value, found := strings.Cut(item, "=")
		if found {
			joined[key] = value
		}
	}

	assert.NotContains(t, joined, "COLLECTOR_CRON_SCHEDULE")
	assert.NotContains(t, joined, "COLLECTOR_RUN_STARTUP")
	assert.NotContains(t, joined, "COLLECTOR_ZFS_CRON_SCHEDULE")
	assert.Equal(t, "yes", joined["PRESERVED_VALUE"])
	assert.Equal(t, "true", joined["SCRUTINY_SUPPRESS_STARTUP_BANNER"])
}

func testLogger() *logrus.Entry {
	logger := logrus.New()
	logger.SetOutput(ioDiscard{})
	return logrus.NewEntry(logger)
}

type ioDiscard struct{}

func (ioDiscard) Write(data []byte) (int, error) {
	return len(data), nil
}
