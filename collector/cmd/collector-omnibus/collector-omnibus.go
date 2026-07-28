package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/analogj/scrutiny/collector/pkg/omnibus"
	"github.com/analogj/scrutiny/webapp/backend/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var goos string
var goarch string

func main() {
	app := &cli.App{
		Name:     "scrutiny-collector-omnibus",
		Usage:    "schedule and supervise Scrutiny collector binaries",
		Version:  version.VERSION,
		Compiled: time.Now(),
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "Run configured collectors on their schedules",
				Flags:  configFlags(),
				Action: runManager,
			},
			{
				Name:   "validate",
				Usage:  "Validate manager configuration and bundled files",
				Flags:  configFlags(),
				Action: validateConfig,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func configFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "config",
			Usage:    "Path to collector-omnibus.yaml",
			EnvVars:  []string{"COLLECTOR_OMNIBUS_CONFIG"},
			Required: true,
		},
	}
}

func runManager(cliCtx *cli.Context) error {
	cfg, err := loadConfig(cliCtx.String("config"))
	if err != nil {
		return err
	}

	logger := logrus.WithField("type", "omnibus")
	runner := omnibus.ExecRunner{Stdout: os.Stdout, Stderr: os.Stderr}
	manager := omnibus.NewManager(cfg, runner, logger)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return manager.Run(ctx)
}

func validateConfig(cliCtx *cli.Context) error {
	if _, err := loadConfig(cliCtx.String("config")); err != nil {
		return err
	}
	fmt.Fprintln(cliCtx.App.Writer, "omnibus configuration is valid")
	return nil
}

func loadConfig(path string) (omnibus.Config, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return omnibus.Config{}, fmt.Errorf("resolve omnibus executable: %w", err)
	}
	return omnibus.LoadConfig(path, executablePath, os.LookupEnv)
}
