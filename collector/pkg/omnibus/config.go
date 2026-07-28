package omnibus

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/robfig/cron/v3"
	"go.yaml.in/yaml/v3"
)

// CollectorSpec describes one collector managed by the omnibus process.
type CollectorSpec struct {
	Name             string
	BinaryName       string
	EnabledEnv       string
	ScheduleEnv      string
	RunStartupEnv    string
	StartupSleepEnv  string
	ConfigEnv        string
	SuppressSchedule []string
}

// CollectorSpecs is the stable execution and startup order for bundled collectors.
var CollectorSpecs = []*CollectorSpec{
	{
		Name:             "metrics",
		BinaryName:       "scrutiny-collector-metrics",
		EnabledEnv:       "COLLECTOR_METRICS_ENABLED",
		ScheduleEnv:      "COLLECTOR_CRON_SCHEDULE",
		RunStartupEnv:    "COLLECTOR_RUN_STARTUP",
		StartupSleepEnv:  "COLLECTOR_RUN_STARTUP_SLEEP",
		ConfigEnv:        "COLLECTOR_METRICS_CONFIG",
		SuppressSchedule: []string{"COLLECTOR_CRON_SCHEDULE", "COLLECTOR_RUN_STARTUP", "COLLECTOR_RUN_STARTUP_SLEEP"},
	},
	{
		Name:             "zfs",
		BinaryName:       "scrutiny-collector-zfs",
		EnabledEnv:       "COLLECTOR_ZFS_ENABLED",
		ScheduleEnv:      "COLLECTOR_ZFS_CRON_SCHEDULE",
		RunStartupEnv:    "COLLECTOR_ZFS_RUN_STARTUP",
		StartupSleepEnv:  "COLLECTOR_ZFS_RUN_STARTUP_SLEEP",
		ConfigEnv:        "COLLECTOR_ZFS_CONFIG",
		SuppressSchedule: []string{"COLLECTOR_ZFS_CRON_SCHEDULE", "COLLECTOR_ZFS_RUN_STARTUP", "COLLECTOR_ZFS_RUN_STARTUP_SLEEP"},
	},
	{
		Name:             "mdadm",
		BinaryName:       "scrutiny-collector-mdadm",
		EnabledEnv:       "COLLECTOR_MDADM_ENABLED",
		ScheduleEnv:      "COLLECTOR_MDADM_CRON_SCHEDULE",
		RunStartupEnv:    "COLLECTOR_MDADM_RUN_STARTUP",
		StartupSleepEnv:  "COLLECTOR_MDADM_RUN_STARTUP_SLEEP",
		ConfigEnv:        "COLLECTOR_MDADM_CONFIG",
		SuppressSchedule: []string{"COLLECTOR_MDADM_CRON_SCHEDULE", "COLLECTOR_MDADM_RUN_STARTUP", "COLLECTOR_MDADM_RUN_STARTUP_SLEEP"},
	},
	{
		Name:             "btrfs",
		BinaryName:       "scrutiny-collector-btrfs",
		EnabledEnv:       "COLLECTOR_BTRFS_ENABLED",
		ScheduleEnv:      "COLLECTOR_BTRFS_CRON_SCHEDULE",
		RunStartupEnv:    "COLLECTOR_BTRFS_RUN_STARTUP",
		StartupSleepEnv:  "COLLECTOR_BTRFS_RUN_STARTUP_SLEEP",
		ConfigEnv:        "COLLECTOR_BTRFS_CONFIG",
		SuppressSchedule: []string{"COLLECTOR_BTRFS_CRON_SCHEDULE", "COLLECTOR_BTRFS_RUN_STARTUP", "COLLECTOR_BTRFS_RUN_STARTUP_SLEEP"},
	},
	{
		Name:             "filesystem",
		BinaryName:       "scrutiny-collector-filesystem",
		EnabledEnv:       "COLLECTOR_FILESYSTEM_ENABLED",
		ScheduleEnv:      "COLLECTOR_FILESYSTEM_CRON_SCHEDULE",
		RunStartupEnv:    "COLLECTOR_FILESYSTEM_RUN_STARTUP",
		StartupSleepEnv:  "COLLECTOR_FILESYSTEM_RUN_STARTUP_SLEEP",
		ConfigEnv:        "COLLECTOR_FILESYSTEM_CONFIG",
		SuppressSchedule: []string{"COLLECTOR_FILESYSTEM_CRON_SCHEDULE", "COLLECTOR_FILESYSTEM_RUN_STARTUP", "COLLECTOR_FILESYSTEM_RUN_STARTUP_SLEEP"},
	},
	{
		Name:             "performance",
		BinaryName:       "scrutiny-collector-performance",
		EnabledEnv:       "COLLECTOR_PERF_ENABLED",
		ScheduleEnv:      "COLLECTOR_PERF_CRON_SCHEDULE",
		RunStartupEnv:    "COLLECTOR_PERF_RUN_STARTUP",
		StartupSleepEnv:  "COLLECTOR_PERF_RUN_STARTUP_SLEEP",
		ConfigEnv:        "COLLECTOR_PERF_CONFIG",
		SuppressSchedule: []string{"COLLECTOR_PERF_CRON_SCHEDULE", "COLLECTOR_PERF_RUN_STARTUP", "COLLECTOR_PERF_RUN_STARTUP_SLEEP"},
	},
}

// CollectorConfig controls one child collector.
type CollectorConfig struct {
	Enabled          bool   `yaml:"enabled"`
	Schedule         string `yaml:"schedule"`
	RunOnStartup     bool   `yaml:"run_on_startup"`
	StartupSleepSecs int    `yaml:"startup_sleep_secs"`
	ConfigPath       string `yaml:"config"`
}

// Config controls the standalone omnibus manager.
type Config struct {
	BinaryDir  string                     `yaml:"binary_dir"`
	Collectors map[string]CollectorConfig `yaml:"collectors"`
}

// LookupEnv allows deterministic environment override tests.
type LookupEnv func(string) (string, bool)

// LoadConfig reads, resolves, overrides, and validates an omnibus configuration.
func LoadConfig(path, executablePath string, lookupEnv LookupEnv) (Config, error) {
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open omnibus config: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode omnibus config: %w", err)
	}
	if cfg.Collectors == nil {
		cfg.Collectors = map[string]CollectorConfig{}
	}

	configDir, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return Config{}, fmt.Errorf("resolve omnibus config directory: %w", err)
	}
	if cfg.BinaryDir == "" {
		cfg.BinaryDir = filepath.Dir(executablePath)
	}
	if value, ok := lookupEnv("COLLECTOR_OMNIBUS_BINARY_DIR"); ok {
		cfg.BinaryDir = value
	}
	cfg.BinaryDir = resolvePath(configDir, cfg.BinaryDir)

	known := make(map[string]struct{}, len(CollectorSpecs))
	for _, spec := range CollectorSpecs {
		known[spec.Name] = struct{}{}
		collectorCfg := cfg.Collectors[spec.Name]
		if err := applyEnvironmentOverrides(&collectorCfg, spec, lookupEnv); err != nil {
			return Config{}, err
		}
		if collectorCfg.ConfigPath != "" {
			collectorCfg.ConfigPath = resolvePath(configDir, collectorCfg.ConfigPath)
		}
		cfg.Collectors[spec.Name] = collectorCfg
	}
	for name := range cfg.Collectors {
		if _, ok := known[name]; !ok {
			return Config{}, fmt.Errorf("unknown collector %q", name)
		}
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate checks manager configuration and required local files.
func (cfg Config) Validate() error {
	enabled := 0
	for _, spec := range CollectorSpecs {
		collectorCfg := cfg.Collectors[spec.Name]
		if !collectorCfg.Enabled {
			continue
		}
		enabled++
		if collectorCfg.Schedule == "" && !collectorCfg.RunOnStartup {
			return fmt.Errorf("collector %q is enabled but has no schedule or startup run", spec.Name)
		}
		if collectorCfg.Schedule != "" {
			if _, err := cron.ParseStandard(collectorCfg.Schedule); err != nil {
				return fmt.Errorf("collector %q has invalid schedule %q: %w", spec.Name, collectorCfg.Schedule, err)
			}
		}
		if collectorCfg.StartupSleepSecs < 0 {
			return fmt.Errorf("collector %q startup_sleep_secs must not be negative", spec.Name)
		}
		if collectorCfg.ConfigPath == "" {
			return fmt.Errorf("collector %q requires a config path", spec.Name)
		}
		if err := requireRegularFile(collectorCfg.ConfigPath); err != nil {
			return fmt.Errorf("collector %q config: %w", spec.Name, err)
		}
		binaryPath := filepath.Join(cfg.BinaryDir, executableName(spec.BinaryName))
		if err := requireRegularFile(binaryPath); err != nil {
			return fmt.Errorf("collector %q binary: %w", spec.Name, err)
		}
	}
	if enabled == 0 {
		return fmt.Errorf("at least one collector must be enabled")
	}
	return nil
}

func applyEnvironmentOverrides(cfg *CollectorConfig, spec *CollectorSpec, lookupEnv LookupEnv) error {
	enabledExplicitlySet := false
	if value, ok := lookupEnv(spec.EnabledEnv); ok {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("%s must be a boolean: %w", spec.EnabledEnv, err)
		}
		cfg.Enabled = parsed
		enabledExplicitlySet = true
	}
	if value, ok := lookupEnv(spec.ScheduleEnv); ok {
		cfg.Schedule = strings.TrimSpace(value)
		if cfg.Schedule != "" && !enabledExplicitlySet {
			cfg.Enabled = true
		}
	}
	if value, ok := lookupEnv(spec.RunStartupEnv); ok {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("%s must be a boolean: %w", spec.RunStartupEnv, err)
		}
		cfg.RunOnStartup = parsed
		if parsed && !enabledExplicitlySet {
			cfg.Enabled = true
		}
	}
	if value, ok := lookupEnv(spec.StartupSleepEnv); ok {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("%s must be an integer: %w", spec.StartupSleepEnv, err)
		}
		cfg.StartupSleepSecs = parsed
	}
	if value, ok := lookupEnv(spec.ConfigEnv); ok {
		cfg.ConfigPath = value
	}
	return nil
}

func resolvePath(base, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(base, path))
}

func requireRegularFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", path)
	}
	return nil
}
