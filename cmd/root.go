package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fugginold/dockwatch/internal/daemon"
	"github.com/fugginold/dockwatch/internal/flags"
	"github.com/fugginold/dockwatch/pkg/container"
	"github.com/fugginold/dockwatch/pkg/filters"
	"github.com/fugginold/dockwatch/pkg/registry"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = NewRootCommand()

// NewRootCommand creates the root command for dockwatch
func NewRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "dockwatch",
		Short: "Automatically updates running Docker containers",
		Long: `
	Dockwatch automatically updates running Docker containers whenever a new image is released.
	More information available at https://github.com/fugginold/dockwatch/.
	`,
		Run:    Run,
		PreRun: PreRun,
		Args:   cobra.ArbitraryArgs,
	}
}

func init() {
	flags.RegisterDockerFlags(rootCmd)
	flags.RegisterSystemFlags(rootCmd)
}

// Execute the root func and exit in case of errors
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// PreRun sets up logging, secrets and environment before the run. It is the
// thin cobra-coupled shell; all configuration assembly lives in buildConfig.
func PreRun(cmd *cobra.Command, _ []string) {
	f := cmd.PersistentFlags()
	flags.ProcessFlagAliases(f)
	if err := flags.SetupLogging(f); err != nil {
		log.Fatalf("Failed to initialize logging: %s", err.Error())
	}
	flags.GetSecretsFromFiles(cmd)
	if err := flags.EnvConfig(cmd); err != nil {
		log.Fatal(err)
	}
	if registry.EnvCredentialsAreUnscoped() {
		log.Warn("REPO_USER/REPO_PASS are set without REPO_HOST, so they will be offered to " +
			"every registry hosting a watched image. Set REPO_HOST to restrict them to one registry.")
	}
}

// Run is the thin entry point: it resolves configuration and options from the
// flags and hands control to the daemon. os.Exit and flag reading stay here;
// the orchestration lives behind daemon.Run.
func Run(c *cobra.Command, names []string) {
	f := c.PersistentFlags()

	if healthCheck, _ := f.GetBool("health-check"); healthCheck {
		// The health check should never run as pid 1.
		if os.Getpid() == 1 {
			time.Sleep(1 * time.Second)
			log.Fatal("The health check flag should never be passed to the main dockwatch container process")
		}
		os.Exit(0)
	}

	cfg, err := buildConfig(c)
	if err != nil {
		log.Fatal(err)
	}

	filter, filterDesc := filters.BuildFilter(names, cfg.DisableContainers, cfg.EnableLabel, cfg.Scope)

	runOnce, _ := f.GetBool("run-once")
	forceUpdate, _ := f.GetBool("force-update")
	enableUpdateAPI, _ := f.GetBool("http-api-update")
	enableMetricsAPI, _ := f.GetBool("http-api-metrics")
	unblockHTTPAPI, _ := f.GetBool("http-api-periodic-polls")
	apiToken, _ := f.GetString("http-api-token")
	noStartupMessage, _ := f.GetBool("no-startup-message")

	opts := daemon.Options{
		Filter:           filter,
		FilterDesc:       filterDesc,
		RunOnce:          runOnce || forceUpdate,
		EnableUpdateAPI:  enableUpdateAPI,
		EnableMetricsAPI: enableMetricsAPI,
		UnblockHTTPAPI:   unblockHTTPAPI,
		Interactive:      isInteractiveInput(os.Stdin, os.Stdout),
		NoStartupMessage: noStartupMessage,
		APIToken:         apiToken,
		In:               os.Stdin,
		Out:              os.Stdout,
	}

	if err := daemon.New(cfg).Run(opts); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	os.Exit(0)
}

// buildConfig assembles the resolved daemon.Config from the command's flags and
// validates it. It replaces the former package-global config and keeps the
// legality rules (via Config.Validate) in one testable place.
func buildConfig(c *cobra.Command) (daemon.Config, error) {
	f := c.PersistentFlags()
	rf := flags.ReadFlags(c)

	scheduleSpec, _ := f.GetString("schedule")
	enableLabel, _ := f.GetBool("label-enable")
	disableContainers, _ := f.GetStringSlice("disable-containers")
	lifecycleHooks, _ := f.GetBool("enable-lifecycle-hooks")
	rollingRestart, _ := f.GetBool("rolling-restart")
	scope, _ := f.GetString("scope")
	labelPrecedence, _ := f.GetBool("label-take-precedence")
	noPull, _ := f.GetBool("no-pull")

	cfg := daemon.Config{
		ScheduleSpec:      scheduleSpec,
		Cleanup:           rf.Cleanup,
		NoRestart:         rf.NoRestart,
		MonitorOnly:       rf.MonitorOnly,
		Timeout:           rf.Timeout,
		EnableLabel:       enableLabel,
		DisableContainers: disableContainers,
		LifecycleHooks:    lifecycleHooks,
		RollingRestart:    rollingRestart,
		Scope:             scope,
		LabelPrecedence:   labelPrecedence,
		NoPull:            noPull,
	}

	if err := cfg.Validate(); err != nil {
		return daemon.Config{}, err
	}

	if cfg.MonitorOnly && cfg.NoPull {
		log.Warn("Using `DOCKWATCH_NO_PULL` and `DOCKWATCH_MONITOR_ONLY` simultaneously might lead to no action being taken at all. If this is intentional, you may safely ignore this message.")
	}
	if cfg.Scope != "" {
		log.Debugf(`Using scope %q`, cfg.Scope)
	}

	includeStopped, _ := f.GetBool("include-stopped")
	includeRestarting, _ := f.GetBool("include-restarting")
	reviveStopped, _ := f.GetBool("revive-stopped")
	removeVolumes, _ := f.GetBool("remove-volumes")
	warnOnHeadPullFailed, _ := f.GetString("warn-on-head-failure")

	client, err := container.NewClient(container.ClientOptions{
		IncludeStopped:    includeStopped,
		ReviveStopped:     reviveStopped,
		RemoveVolumes:     removeVolumes,
		IncludeRestarting: includeRestarting,
		WarnOnHeadFailed:  container.WarningStrategy(warnOnHeadPullFailed),
	})
	if err != nil {
		return daemon.Config{}, fmt.Errorf("error instantiating Docker client: %w", err)
	}
	cfg.Client = client

	return cfg, nil
}

// isInteractiveInput reports whether both stdin and stdout are a TTY.
func isInteractiveInput(stdin *os.File, stdout *os.File) bool {
	if stdin == nil || stdout == nil {
		return false
	}
	stdinStat, err := stdin.Stat()
	if err != nil {
		return false
	}
	stdoutStat, err := stdout.Stat()
	if err != nil {
		return false
	}
	return (stdinStat.Mode()&os.ModeCharDevice) != 0 && (stdoutStat.Mode()&os.ModeCharDevice) != 0
}
