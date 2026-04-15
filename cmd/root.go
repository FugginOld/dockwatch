package cmd

import (
	"errors"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fugginold/dockwatch/internal/actions"
	"github.com/fugginold/dockwatch/internal/flags"
	"github.com/fugginold/dockwatch/internal/meta"
	"github.com/fugginold/dockwatch/pkg/api"
	apiMetrics "github.com/fugginold/dockwatch/pkg/api/metrics"
	apiSchedule "github.com/fugginold/dockwatch/pkg/api/schedule"
	"github.com/fugginold/dockwatch/pkg/api/update"
	"github.com/fugginold/dockwatch/pkg/container"
	"github.com/fugginold/dockwatch/pkg/filters"
	"github.com/fugginold/dockwatch/pkg/metrics"
	t "github.com/fugginold/dockwatch/pkg/types"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// runConfig holds all runtime configuration populated during PreRun.
// Consolidating these into a struct avoids package-level mutable state.
type runConfig struct {
	client            container.Client
	scheduleSpec      string
	cleanup           bool
	noRestart         bool
	noPull            bool
	monitorOnly       bool
	enableLabel       bool
	disableContainers []string
	timeout           time.Duration
	lifecycleHooks    bool
	rollingRestart    bool
	scope             string
	labelPrecedence   bool
}

var cfg runConfig

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
	flags.SetDefaults()
	flags.RegisterDockerFlags(rootCmd)
	flags.RegisterSystemFlags(rootCmd)
}

// Execute the root func and exit in case of errors
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// PreRun is a lifecycle hook that runs before the command is executed.
func PreRun(cmd *cobra.Command, _ []string) {
	f := cmd.PersistentFlags()
	flags.ProcessFlagAliases(f)
	if err := flags.SetupLogging(f); err != nil {
		log.Fatalf("Failed to initialize logging: %s", err.Error())
	}

	cfg.scheduleSpec, _ = f.GetString("schedule")

	flags.GetSecretsFromFiles(cmd)
	rf := flags.ReadFlags(cmd)
	cfg.cleanup = rf.Cleanup
	cfg.noRestart = rf.NoRestart
	cfg.monitorOnly = rf.MonitorOnly
	cfg.timeout = rf.Timeout

	if cfg.timeout < 0 {
		log.Fatal("Please specify a positive value for timeout value.")
	}

	cfg.enableLabel, _ = f.GetBool("label-enable")
	cfg.disableContainers, _ = f.GetStringSlice("disable-containers")
	cfg.lifecycleHooks, _ = f.GetBool("enable-lifecycle-hooks")
	cfg.rollingRestart, _ = f.GetBool("rolling-restart")
	cfg.scope, _ = f.GetString("scope")
	cfg.labelPrecedence, _ = f.GetBool("label-take-precedence")

	if cfg.scope != "" {
		log.Debugf(`Using scope %q`, cfg.scope)
	}

	// configure environment vars for client
	err := flags.EnvConfig(cmd)
	if err != nil {
		log.Fatal(err)
	}

	cfg.noPull, _ = f.GetBool("no-pull")
	includeStopped, _ := f.GetBool("include-stopped")
	includeRestarting, _ := f.GetBool("include-restarting")
	reviveStopped, _ := f.GetBool("revive-stopped")
	removeVolumes, _ := f.GetBool("remove-volumes")
	warnOnHeadPullFailed, _ := f.GetString("warn-on-head-failure")

	if cfg.monitorOnly && cfg.noPull {
		log.Warn("Using `DOCKWATCH_NO_PULL` and `DOCKWATCH_MONITOR_ONLY` simultaneously might lead to no action being taken at all. If this is intentional, you may safely ignore this message.")
	}

	cfg.client, err = container.NewClient(container.ClientOptions{
		IncludeStopped:    includeStopped,
		ReviveStopped:     reviveStopped,
		RemoveVolumes:     removeVolumes,
		IncludeRestarting: includeRestarting,
		WarnOnHeadFailed:  container.WarningStrategy(warnOnHeadPullFailed),
	})
	if err != nil {
		log.Fatalf("Error instantiating Docker client: %s", err)
	}
}

// Run is the main execution flow of the command
func Run(c *cobra.Command, names []string) {
	filter, filterDesc := filters.BuildFilter(names, cfg.disableContainers, cfg.enableLabel, cfg.scope)
	runOnce, _ := c.PersistentFlags().GetBool("run-once")
	forceUpdate, _ := c.PersistentFlags().GetBool("force-update")
	enableUpdateAPI, _ := c.PersistentFlags().GetBool("http-api-update")
	enableMetricsAPI, _ := c.PersistentFlags().GetBool("http-api-metrics")
	unblockHTTPAPI, _ := c.PersistentFlags().GetBool("http-api-periodic-polls")
	apiToken, _ := c.PersistentFlags().GetString("http-api-token")
	healthCheck, _ := c.PersistentFlags().GetBool("health-check")

	if forceUpdate {
		runOnce = true
	}

	if healthCheck {
		// health check should not have pid 1
		if os.Getpid() == 1 {
			time.Sleep(1 * time.Second)
			log.Fatal("The health check flag should never be passed to the main dockwatch container process")
		}
		os.Exit(0)
	}

	if cfg.rollingRestart && cfg.monitorOnly {
		log.Fatal("Rolling restarts is not compatible with the global monitor only flag")
	}

	awaitDockerClient(cfg.client)

	if err := actions.CheckForSanity(cfg.client, filter, cfg.rollingRestart); err != nil {
		logNotifyExit(err)
	}

	if runOnce {
		writeStartupMessage(c, time.Time{}, filterDesc)
		runUpdates(filter)
		os.Exit(0)
		return
	}

	if err := actions.CheckForMultipleDockwatchInstances(cfg.client, cfg.cleanup, cfg.scope); err != nil {
		logNotifyExit(err)
	}

	// The lock is shared between the scheduler and the HTTP API. It only allows one update to run at a time.
	updateLock := make(chan bool, 1)
	updateLock <- true
	periodicEnabled := !enableUpdateAPI || unblockHTTPAPI

	isInteractive := isInteractiveInput(os.Stdin, os.Stdout)

	var scheduleCtrl *scheduleController
	if periodicEnabled || isInteractive {
		scheduleCtrl = newScheduleController(updateLock, filter)
		if periodicEnabled {
			nextRun, err := scheduleCtrl.Set(cfg.scheduleSpec)
			if err != nil {
				log.Error(err)
				os.Exit(1)
				return
			}
			writeStartupMessage(c, nextRun, filterDesc)
		} else {
			writeStartupMessage(c, time.Time{}, filterDesc)
		}
	} else if !unblockHTTPAPI {
		writeStartupMessage(c, time.Time{}, filterDesc)
	}

	httpAPI := api.New(apiToken)

	if enableUpdateAPI {
		updateHandler := update.New(func(images []string) {
			metric := runUpdates(filters.FilterByImage(images, filter))
			metrics.RegisterScan(metric)
		}, updateLock)
		httpAPI.RegisterFunc(updateHandler.Path, updateHandler.Handle)

		if scheduleCtrl != nil {
			scheduleHandler := apiSchedule.New(
				func() string { return scheduleCtrl.Current() },
				func() time.Time { return scheduleCtrl.NextRun() },
				func(spec string) (time.Time, error) { return scheduleCtrl.Set(spec) },
			)
			httpAPI.RegisterFunc(scheduleHandler.Path, scheduleHandler.Handle)
		}
	}

	if enableMetricsAPI {
		metricsHandler := apiMetrics.New()
		httpAPI.RegisterHandler(metricsHandler.Path, metricsHandler.Handle)
	}

	apiShouldBlock := enableUpdateAPI && !unblockHTTPAPI && !isInteractive
	if err := httpAPI.Start(apiShouldBlock); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to start API", err)
	}

	if isInteractive {
		go func() {
			runShell(scheduleCtrl, updateLock, filter)
			// Trigger a clean shutdown when the shell exits
			p, _ := os.FindProcess(os.Getpid())
			_ = p.Signal(os.Interrupt)
		}()
	}

	if periodicEnabled || isInteractive {
		waitForInterrupt()
		if scheduleCtrl != nil {
			scheduleCtrl.Stop()
		}
		log.Info("Waiting for running update to be finished...")
		<-updateLock
	}

	os.Exit(0) // clean shutdown
}

func logNotifyExit(err error) {
	log.Error(err)
	os.Exit(1)
}

// awaitDockerClient retries Ping until the Docker API is reachable, failing
// fatally after 5 attempts. This replaces a fixed 1-second sleep that could
// succeed before Docker was ready or fail silently when it wasn't.
func awaitDockerClient(client container.Client) {
	for attempt := 1; attempt <= 5; attempt++ {
		if err := client.Ping(); err == nil {
			return
		}
		log.Debugf("Docker API not yet available (attempt %d/5), retrying in 1s...", attempt)
		time.Sleep(1 * time.Second)
	}
	log.Fatal("Docker API unreachable after 5 attempts")
}

func formatDuration(d time.Duration) string {
	sb := strings.Builder{}

	hours := int64(d.Hours())
	minutes := int64(math.Mod(d.Minutes(), 60))
	seconds := int64(math.Mod(d.Seconds(), 60))

	if hours == 1 {
		sb.WriteString("1 hour")
	} else if hours != 0 {
		sb.WriteString(strconv.FormatInt(hours, 10))
		sb.WriteString(" hours")
	}

	if hours != 0 && (seconds != 0 || minutes != 0) {
		sb.WriteString(", ")
	}

	if minutes == 1 {
		sb.WriteString("1 minute")
	} else if minutes != 0 {
		sb.WriteString(strconv.FormatInt(minutes, 10))
		sb.WriteString(" minutes")
	}

	if minutes != 0 && (seconds != 0) {
		sb.WriteString(", ")
	}

	if seconds == 1 {
		sb.WriteString("1 second")
	} else if seconds != 0 || (hours == 0 && minutes == 0) {
		sb.WriteString(strconv.FormatInt(seconds, 10))
		sb.WriteString(" seconds")
	}

	return sb.String()
}

func writeStartupMessage(c *cobra.Command, sched time.Time, filtering string) {
	noStartupMessage, _ := c.PersistentFlags().GetBool("no-startup-message")
	enableUpdateAPI, _ := c.PersistentFlags().GetBool("http-api-update")

	if !noStartupMessage {
		startupLog := log.NewEntry(log.StandardLogger())
		startupLog.Info("Dockwatch ", meta.Version)
		startupLog.Info(filtering)

		if !sched.IsZero() {
			until := formatDuration(time.Until(sched))
			startupLog.Info("Scheduling first run: " + sched.Format("2006-01-02 15:04:05 -0700 MST"))
			startupLog.Info("Note that the first check will be performed in " + until)
		} else if runOnce, _ := c.PersistentFlags().GetBool("run-once"); runOnce {
			startupLog.Info("Running a one time update.")
		} else {
			startupLog.Info("Periodic runs are not enabled.")
		}

		if enableUpdateAPI {
			// TODO: make listen port configurable
			startupLog.Info("The HTTP API is enabled at :8080.")
		}
	}

	if log.IsLevelEnabled(log.TraceLevel) {
		log.Warn("Trace level enabled: log will include sensitive information as credentials and tokens")
	}
}

func waitForInterrupt() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(interrupt, syscall.SIGTERM)

	<-interrupt
}

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

type scheduleController struct {
	mu        sync.Mutex
	scheduler *cron.Cron
	lock      chan bool
	filter    t.Filter
	spec      string
}

func newScheduleController(lock chan bool, filter t.Filter) *scheduleController {
	return &scheduleController{
		lock:   lock,
		filter: filter,
	}
}

func (c *scheduleController) Set(spec string) (time.Time, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	scheduler := cron.New()
	err := scheduler.AddFunc(spec, func() {
		select {
		case v := <-c.lock:
			defer func() { c.lock <- v }()
			metric := runUpdates(c.filter)
			metrics.RegisterScan(metric)
		default:
			metrics.RegisterScan(nil)
			log.Debug("Skipped another update already running.")
		}

		nextRuns := scheduler.Entries()
		if len(nextRuns) > 0 {
			log.Debug("Scheduled next run: " + nextRuns[0].Next.String())
		}
	})
	if err != nil {
		return time.Time{}, err
	}

	nextRun := scheduler.Entries()[0].Schedule.Next(time.Now())
	scheduler.Start()

	if c.scheduler != nil {
		c.scheduler.Stop()
	}

	c.scheduler = scheduler
	c.spec = spec

	return nextRun, nil
}

func (c *scheduleController) Current() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.spec
}

func (c *scheduleController) NextRun() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.scheduler == nil {
		return time.Time{}
	}
	entries := c.scheduler.Entries()
	if len(entries) == 0 {
		return time.Time{}
	}
	return entries[0].Next
}

func (c *scheduleController) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.scheduler != nil {
		c.scheduler.Stop()
	}
}

func runUpdates(filter t.Filter) *metrics.Metric {
	updateParams := t.UpdateParams{
		Filter:          filter,
		Cleanup:         cfg.cleanup,
		NoRestart:       cfg.noRestart,
		Timeout:         cfg.timeout,
		MonitorOnly:     cfg.monitorOnly,
		LifecycleHooks:  cfg.lifecycleHooks,
		RollingRestart:  cfg.rollingRestart,
		LabelPrecedence: cfg.labelPrecedence,
		NoPull:          cfg.noPull,
	}
	result, err := actions.Update(cfg.client, updateParams)
	if err != nil {
		log.Error(err)
	}
	metricResults := metrics.NewMetric(result)
	log.WithFields(log.Fields{
		"Scanned": metricResults.Scanned,
		"Updated": metricResults.Updated,
		"Failed":  metricResults.Failed,
	}).Info("Session done")
	return metricResults
}
