package daemon

import (
	"errors"
	"io"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fugginold/dockwatch/internal/actions"
	"github.com/fugginold/dockwatch/internal/meta"
	"github.com/fugginold/dockwatch/pkg/api"
	apiSchedule "github.com/fugginold/dockwatch/pkg/api/schedule"
	"github.com/fugginold/dockwatch/pkg/api/update"
	"github.com/fugginold/dockwatch/pkg/container"
	"github.com/fugginold/dockwatch/pkg/filters"
	"github.com/fugginold/dockwatch/pkg/metrics"
	t "github.com/fugginold/dockwatch/pkg/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// Options describe a single invocation's run mode. They are the run-time knobs
// that come from the command line, separated from the resolved Config.
type Options struct {
	Filter           t.Filter
	FilterDesc       string
	RunOnce          bool
	EnableUpdateAPI  bool
	EnableMetricsAPI bool
	UnblockHTTPAPI   bool
	Interactive      bool
	NoStartupMessage bool
	APIToken         string
	In               io.Reader
	Out              io.Writer
}

// mode is the resolved run-mode decision derived purely from Options.
type mode struct {
	periodicEnabled bool
	apiShouldBlock  bool
}

// decideMode collapses the run-mode boolean logic into one pure, testable place.
func decideMode(o Options) mode {
	return mode{
		periodicEnabled: !o.EnableUpdateAPI || o.UnblockHTTPAPI,
		apiShouldBlock:  o.EnableUpdateAPI && !o.UnblockHTTPAPI && !o.Interactive,
	}
}

// Daemon owns the update run: the single-flight Runner and the schedule.
type Daemon struct {
	cfg    Config
	runner *Runner
	sched  *Controller
}

// New builds a Daemon from a resolved Config.
func New(cfg Config) *Daemon {
	return &Daemon{cfg: cfg, runner: NewRunner(cfg)}
}

// Run executes the daemon according to opts and returns an error for any fatal
// condition. A nil error means a clean exit. It blocks until interrupted when
// running in periodic or interactive mode.
func (d *Daemon) Run(opts Options) error {
	awaitDockerClient(d.cfg.Client)

	if err := actions.CheckForSanity(d.cfg.Client, opts.Filter, d.cfg.RollingRestart); err != nil {
		return err
	}

	if opts.RunOnce {
		d.writeStartupMessage(opts, time.Time{})
		d.runner.Run(opts.Filter)
		return nil
	}

	if err := actions.CheckForMultipleDockwatchInstances(d.cfg.Client, d.cfg.Cleanup, d.cfg.Scope); err != nil {
		return err
	}

	m := decideMode(opts)

	if m.periodicEnabled || opts.Interactive {
		d.sched = NewController(d.runner, opts.Filter)
		if m.periodicEnabled {
			nextRun, err := d.sched.Set(d.cfg.ScheduleSpec)
			if err != nil {
				return err
			}
			d.writeStartupMessage(opts, nextRun)
		} else {
			d.writeStartupMessage(opts, time.Time{})
		}
	} else if !opts.UnblockHTTPAPI {
		d.writeStartupMessage(opts, time.Time{})
	}

	httpAPI := api.New(opts.APIToken)
	if opts.EnableUpdateAPI {
		updateHandler := update.New(d.scanner(opts.Filter))
		httpAPI.RegisterFunc(updateHandler.Path, updateHandler.Handle)

		if d.sched != nil {
			scheduleHandler := apiSchedule.New(
				func() string { return d.sched.Current() },
				func() time.Time { return d.sched.NextRun() },
				func(spec string) (time.Time, error) { return d.sched.Set(spec) },
			)
			httpAPI.RegisterFunc(scheduleHandler.Path, scheduleHandler.Handle)
		}
	}
	if opts.EnableMetricsAPI {
		metrics.Default()
		httpAPI.RegisterHandler("/v1/metrics", promhttp.Handler())
	}

	if err := httpAPI.Start(m.apiShouldBlock); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to start API", err)
	}

	if opts.Interactive {
		go func() {
			RunShell(d.sched, d.runner, opts.Filter, opts.In, opts.Out)
			// Trigger a clean shutdown when the shell exits.
			p, _ := os.FindProcess(os.Getpid())
			_ = p.Signal(os.Interrupt)
		}()
	}

	if m.periodicEnabled || opts.Interactive {
		waitForInterrupt()
		if d.sched != nil {
			d.sched.Stop()
		}
		log.Info("Waiting for running update to be finished...")
		d.runner.Wait()
	}

	return nil
}

// scanner adapts the Runner to the HTTP update API's Scanner interface,
// mapping requested images onto the base filter.
func (d *Daemon) scanner(base t.Filter) update.Scanner {
	return scannerFunc(func(images []string, wait bool) {
		filter := filters.FilterByImage(images, base)
		if wait {
			d.runner.Run(filter)
		} else {
			d.runner.TryRun(filter)
		}
	})
}

type scannerFunc func(images []string, wait bool)

func (f scannerFunc) Scan(images []string, wait bool) { f(images, wait) }

// awaitDockerClient retries Ping until the Docker API is reachable, failing
// fatally after 5 attempts.
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

func (d *Daemon) writeStartupMessage(opts Options, sched time.Time) {
	if opts.NoStartupMessage {
		return
	}

	startupLog := log.NewEntry(log.StandardLogger())
	startupLog.Info("Dockwatch ", meta.Version)
	startupLog.Info(opts.FilterDesc)

	if !sched.IsZero() {
		until := formatDuration(time.Until(sched))
		startupLog.Info("Scheduling first run: " + sched.Format("2006-01-02 15:04:05 -0700 MST"))
		startupLog.Info("Note that the first check will be performed in " + until)
	} else if opts.RunOnce {
		startupLog.Info("Running a one time update.")
	} else {
		startupLog.Info("Periodic runs are not enabled.")
	}

	if opts.EnableUpdateAPI {
		// TODO: make listen port configurable
		startupLog.Info("The HTTP API is enabled at :8080.")
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
