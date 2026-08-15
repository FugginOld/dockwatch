package flags

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// DockerAPIMinVersion is the minimum version of the docker api required to
// use dockwatch
const DockerAPIMinVersion string = "1.25"

var defaultInterval = int((time.Hour * 24).Seconds())

// RegisterDockerFlags that are used directly by the docker api client
func RegisterDockerFlags(rootCmd *cobra.Command) {
	flags := rootCmd.PersistentFlags()
	flags.StringP("host", "H", envStringOr("DOCKER_HOST", "unix:///var/run/docker.sock"), "daemon socket to connect to")
	flags.BoolP("tlsverify", "v", envBool("DOCKER_TLS_VERIFY"), "use TLS and verify the remote")
	flags.StringP("api-version", "a", envString("DOCKER_API_VERSION"), "docker api version to use; empty negotiates the highest the daemon supports")
	flags.BoolP(
		"registry-tls-skip-verify",
		"",
		envBool("DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY"),
		"skip TLS certificate verification for image registry HEAD requests (testing only)")
}

// RegisterSystemFlags that are used by dockwatch to modify the program flow
func RegisterSystemFlags(rootCmd *cobra.Command) {
	flags := rootCmd.PersistentFlags()
	flags.IntP(
		"interval",
		"i",
		envInt("DOCKWATCH_POLL_INTERVAL", defaultInterval),
		"Poll interval (in seconds)")

	flags.StringP(
		"schedule",
		"s",
		envString("DOCKWATCH_SCHEDULE"),
		"The cron expression which defines when to update")

	flags.StringP(
		"cron",
		"",
		"",
		"Alias for --schedule")

	flags.DurationP(
		"stop-timeout",
		"t",
		envDuration("DOCKWATCH_TIMEOUT", 10*time.Second),
		"Timeout before a container is forcefully stopped")

	flags.BoolP(
		"no-pull",
		"",
		envBool("DOCKWATCH_NO_PULL"),
		"Do not pull any new images")

	flags.BoolP(
		"no-restart",
		"",
		envBool("DOCKWATCH_NO_RESTART"),
		"Do not restart any containers")

	flags.BoolP(
		"no-startup-message",
		"",
		envBool("DOCKWATCH_NO_STARTUP_MESSAGE"),
		"Prevents dockwatch from sending a startup message")

	flags.BoolP(
		"cleanup",
		"c",
		envBool("DOCKWATCH_CLEANUP"),
		"Remove previously used images after updating")

	flags.BoolP(
		"remove-volumes",
		"",
		envBool("DOCKWATCH_REMOVE_VOLUMES"),
		"Remove attached volumes before updating")

	flags.BoolP(
		"label-enable",
		"e",
		envBool("DOCKWATCH_LABEL_ENABLE"),
		"Watch containers where the com.centurylinklabs.dockwatch.enable label is true")

	flags.StringSliceP(
		"disable-containers",
		"x",
		// Due to issue spf13/viper#380, can't use viper.GetStringSlice:
		regexp.MustCompile("[, ]+").Split(envString("DOCKWATCH_DISABLE_CONTAINERS"), -1),
		"Comma-separated list of containers to explicitly exclude from watching.")

	flags.StringP(
		"log-format",
		"l",
		envStringOr("DOCKWATCH_LOG_FORMAT", "auto"),
		"Sets what logging format to use for console output. Possible values: Auto, LogFmt, Pretty, JSON")

	flags.BoolP(
		"debug",
		"d",
		envBool("DOCKWATCH_DEBUG"),
		"Enable debug mode with verbose logging")

	flags.BoolP(
		"trace",
		"",
		envBool("DOCKWATCH_TRACE"),
		"Enable trace mode with very verbose logging - caution, exposes credentials")

	flags.BoolP(
		"monitor-only",
		"m",
		envBool("DOCKWATCH_MONITOR_ONLY"),
		"Will only monitor for new images, not update the containers")

	flags.BoolP(
		"run-once",
		"R",
		envBool("DOCKWATCH_RUN_ONCE"),
		"Run once now and exit")

	flags.BoolP(
		"force-update",
		"",
		false,
		"Alias for --run-once; immediately check for updates and exit")

	flags.BoolP(
		"include-restarting",
		"",
		envBool("DOCKWATCH_INCLUDE_RESTARTING"),
		"Will also include restarting containers")

	flags.BoolP(
		"include-stopped",
		"S",
		envBool("DOCKWATCH_INCLUDE_STOPPED"),
		"Will also include created and exited containers")

	flags.BoolP(
		"revive-stopped",
		"",
		envBool("DOCKWATCH_REVIVE_STOPPED"),
		"Will also start stopped containers that were updated, if include-stopped is active")

	flags.BoolP(
		"enable-lifecycle-hooks",
		"",
		envBool("DOCKWATCH_LIFECYCLE_HOOKS"),
		"Enable the execution of commands triggered by pre- and post-update lifecycle hooks")

	flags.BoolP(
		"rolling-restart",
		"",
		envBool("DOCKWATCH_ROLLING_RESTART"),
		"Restart containers one at a time")

	flags.BoolP(
		"http-api-update",
		"",
		envBool("DOCKWATCH_HTTP_API_UPDATE"),
		"Runs Dockwatch in HTTP API mode, so that image updates must to be triggered by a request")
	flags.BoolP(
		"http-api-metrics",
		"",
		envBool("DOCKWATCH_HTTP_API_METRICS"),
		"Runs Dockwatch with the Prometheus metrics API enabled")

	flags.StringP(
		"http-api-token",
		"",
		envString("DOCKWATCH_HTTP_API_TOKEN"),
		"Sets an authentication token to HTTP API requests.")

	flags.BoolP(
		"http-api-periodic-polls",
		"",
		envBool("DOCKWATCH_HTTP_API_PERIODIC_POLLS"),
		"Also run periodic updates (specified with --interval and --schedule) if HTTP API is enabled")

	// https://no-color.org/
	flags.BoolP(
		"no-color",
		"",
		envIsSet("NO_COLOR"),
		"Disable ANSI color escape codes in log output")

	flags.StringP(
		"scope",
		"",
		envString("DOCKWATCH_SCOPE"),
		"Defines a monitoring scope for the Dockwatch instance.")

	flags.StringP(
		"porcelain",
		"P",
		envString("DOCKWATCH_PORCELAIN"),
		`Write session results to stdout using a stable versioned format. Supported values: "v1"`)

	flags.String(
		"log-level",
		envStringOr("DOCKWATCH_LOG_LEVEL", "info"),
		"The maximum log level that will be written to STDERR. Possible values: panic, fatal, error, warn, info, debug or trace")

	flags.BoolP(
		"health-check",
		"",
		false,
		"Do health check and exit")

	flags.BoolP(
		"label-take-precedence",
		"",
		envBool("DOCKWATCH_LABEL_TAKE_PRECEDENCE"),
		"Label applied to containers take precedence over arguments")
}

func envString(key string) string {
	return os.Getenv(key)
}

func envStringOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envIsSet(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envBool(key string) bool {
	v, _ := strconv.ParseBool(os.Getenv(key))
	return v
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// EnvConfig translates the command-line options into environment variables
// that will initialize the api client
func EnvConfig(cmd *cobra.Command) error {
	var err error
	var host string
	var tls bool
	var version string
	var registryTLSSkipVerify bool

	flags := cmd.PersistentFlags()

	if host, err = flags.GetString("host"); err != nil {
		return err
	}
	if tls, err = flags.GetBool("tlsverify"); err != nil {
		return err
	}
	if version, err = flags.GetString("api-version"); err != nil {
		return err
	}
	if registryTLSSkipVerify, err = flags.GetBool("registry-tls-skip-verify"); err != nil {
		return err
	}
	if err = setEnvOptStr("DOCKER_HOST", host); err != nil {
		return err
	}
	if err = setEnvOptBool("DOCKER_TLS_VERIFY", tls, flags.Changed("tlsverify")); err != nil {
		return err
	}
	if err = setEnvOptStr("DOCKER_API_VERSION", version); err != nil {
		return err
	}
	if err = setEnvOptBool("DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY", registryTLSSkipVerify, flags.Changed("registry-tls-skip-verify")); err != nil {
		return err
	}
	return nil
}

// RunFlags holds the common runtime flags read by ReadFlags
type RunFlags struct {
	Cleanup     bool
	NoRestart   bool
	MonitorOnly bool
	Timeout     time.Duration
}

// ReadFlags reads common flags used in the main program flow of dockwatch
func ReadFlags(cmd *cobra.Command) RunFlags {
	flags := cmd.PersistentFlags()

	var err error
	var rf RunFlags

	if rf.Cleanup, err = flags.GetBool("cleanup"); err != nil {
		log.Fatal(err)
	}
	if rf.NoRestart, err = flags.GetBool("no-restart"); err != nil {
		log.Fatal(err)
	}
	if rf.MonitorOnly, err = flags.GetBool("monitor-only"); err != nil {
		log.Fatal(err)
	}
	if rf.Timeout, err = flags.GetDuration("stop-timeout"); err != nil {
		log.Fatal(err)
	}

	return rf
}

func setEnvOptStr(env string, opt string) error {
	if opt == "" || opt == os.Getenv(env) {
		return nil
	}
	err := os.Setenv(env, opt)
	if err != nil {
		return err
	}
	return nil
}

// setEnvOptBool writes a boolean flag through to its environment variable, in both
// directions. Only setting it meant an explicit false could never override a value
// inherited from the environment, so --registry-tls-skip-verify=false left
// InsecureSkipVerify on.
//
// A false value clears the variable only when the operator actually passed the flag.
// The flag default comes from envBool, which uses strconv.ParseBool, while the docker
// client asks only whether DOCKER_TLS_VERIFY is non-empty -- so "yes", "on", "0" and
// "false" all yield a false default while the daemon connection is verifying today.
// Clearing on a default-valued flag would silently turn that verification off.
//
// Clearing rather than writing "0" for the same reason: to the docker client "0" is
// non-empty and therefore means verify, the opposite of what was asked.
func setEnvOptBool(env string, opt bool, flagChanged bool) error {
	if opt {
		return setEnvOptStr(env, "1")
	}
	if !flagChanged {
		return nil
	}
	return os.Unsetenv(env)
}

// GetSecretsFromFiles checks if passwords/tokens/webhooks have been passed as a file instead of plaintext.
// If so, the value of the flag will be replaced with the contents of the file.
func GetSecretsFromFiles(rootCmd *cobra.Command) {
	flags := rootCmd.PersistentFlags()

	// Keep these string-valued. The fatal below interpolates the error, which is safe
	// only because a failed flags.Set on a string flag cannot contain the value; a
	// Duration or Int secret would return a strconv error carrying the file contents.
	secrets := []string{
		"http-api-token",
	}
	for _, secret := range secrets {
		if err := getSecretFromFile(flags, secret); err != nil {
			log.Fatalf("failed to get secret from flag %v: %s", secret, err)
		}
	}
}

// getSecretFromFile will check if the flag contains a reference to a file; if it does, replaces the value of the flag with the contents of the file.
func getSecretFromFile(flags *pflag.FlagSet, secret string) error {
	flag := flags.Lookup(secret)
	if sliceValue, ok := flag.Value.(pflag.SliceValue); ok {
		oldValues := sliceValue.GetSlice()
		values := make([]string, 0, len(oldValues))
		for _, value := range oldValues {
			if value != "" && isFile(value) {
				file, err := os.Open(value)
				if err != nil {
					return err
				}
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					if line == "" {
						continue
					}
					values = append(values, line)
				}
				// A read that fails part-way -- a directory, or an unreadable file --
				// otherwise leaves an empty slice behind and the secret silently
				// becomes "no value" instead of a fatal startup error.
				if err := scanner.Err(); err != nil {
					_ = file.Close()
					return err
				}
				if err := file.Close(); err != nil {
					return err
				}
			} else {
				values = append(values, value)
			}
		}
		return sliceValue.Replace(values)
	}

	value := flag.Value.String()
	if value != "" && isFile(value) {
		content, err := os.ReadFile(value)
		if err != nil {
			return err
		}
		return flags.Set(secret, strings.TrimSpace(string(content)))
	}

	return nil
}

func isFile(s string) bool {
	firstColon := strings.IndexRune(s, ':')
	if firstColon != 1 && firstColon != -1 {
		// If the string contains a ':', but it's not the second character, it's probably not a file
		// and will cause a fatal error on windows if stat'ed
		// This still allows for paths that start with 'c:\' etc.
		return false
	}
	return isFileStat(os.Stat(s))
}

// isFileStat decides whether a stat result means "this value is a path".
//
// A value too long or malformed for the filesystem -- a strong random token, for
// instance -- must not be read as a filename, or the resulting error carries the
// secret into the logs. But a path we merely cannot stat, typically because the
// file or a parent directory is not readable by this user, is still a path: saying
// otherwise would leave the flag holding the path string and use that as the
// secret, so an unreadable token file would silently become a guessable one. The
// same applies to a directory -- what a missing bind-mount source or a subPath-less
// secret volume leaves behind: it must stay a path so the read fails loudly rather
// than the flag keeping the path string. The read error names only the path.
func isFileStat(info os.FileInfo, err error) bool {
	if err == nil {
		return info != nil
	}
	return errors.Is(err, os.ErrPermission)
}

// ProcessFlagAliases updates the value of flags that are being set by helper flags
func ProcessFlagAliases(flags *pflag.FlagSet) {

	porcelain, err := flags.GetString(`porcelain`)
	if err != nil {
		log.Fatalf(`Failed to get flag: %v`, err)
	}
	if porcelain != "" {
		if porcelain != "v1" {
			log.Fatalf(`Unknown porcelain version %q. Supported values: "v1"`, porcelain)
		}
	}

	cronValue, err := flags.GetString(`cron`)
	if err != nil {
		log.Fatalf(`Failed to get flag: %v`, err)
	}

	scheduleChanged := flags.Changed(`schedule`)
	intervalChanged := flags.Changed(`interval`)
	cronChanged := flags.Changed(`cron`)
	if cronChanged {
		if scheduleChanged {
			if schedule, _ := flags.GetString(`schedule`); schedule != cronValue {
				log.Fatal(`Only schedule or cron can be defined, not both.`)
			}
		} else {
			_ = flags.Set(`schedule`, cronValue)
			scheduleChanged = true
		}
	}

	// FIXME: snakeswap
	// due to how viper is integrated by swapping the defaults for the flags, we need this hack:
	if val, _ := flags.GetString(`schedule`); val != `` {
		scheduleChanged = true
	}
	if val, _ := flags.GetInt(`interval`); val != defaultInterval {
		intervalChanged = true
	}

	if intervalChanged && scheduleChanged {
		log.Fatal(`Only schedule or interval can be defined, not both.`)
	}

	// update schedule flag to match interval if it's set, or to the default if none of them are
	if intervalChanged || !scheduleChanged {
		interval, _ := flags.GetInt(`interval`)
		_ = flags.Set(`schedule`, fmt.Sprintf(`@every %ds`, interval))
	}

	if flagIsEnabled(flags, `debug`) {
		_ = flags.Set(`log-level`, `debug`)
	}

	if flagIsEnabled(flags, `trace`) {
		_ = flags.Set(`log-level`, `trace`)
	}

	if flagIsEnabled(flags, `force-update`) {
		_ = flags.Set(`run-once`, `true`)
	}

}

// SetupLogging reads only the flags that is needed to set up logging and applies them to the global logger
func SetupLogging(f *pflag.FlagSet) error {
	logFormat, _ := f.GetString(`log-format`)
	noColor, _ := f.GetBool("no-color")

	switch strings.ToLower(logFormat) {
	case "auto":
		// This will either use the "pretty" or "logfmt" format, based on whether the standard out is connected to a TTY
		log.SetFormatter(&log.TextFormatter{
			DisableColors: noColor,
			// enable logrus built-in support for https://bixense.com/clicolors/
			EnvironmentOverrideColors: true,
		})
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	case "logfmt":
		log.SetFormatter(&log.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
	case "pretty":
		log.SetFormatter(&log.TextFormatter{
			// "Pretty" format combined with `--no-color` will only change the timestamp to the time since start
			ForceColors:   !noColor,
			FullTimestamp: false,
		})
	default:
		return fmt.Errorf("invalid log format: %s", logFormat)
	}

	rawLogLevel, _ := f.GetString(`log-level`)
	if logLevel, err := log.ParseLevel(rawLogLevel); err != nil {
		return fmt.Errorf("invalid log level: %e", err)
	} else {
		log.SetLevel(logLevel)
	}

	return nil
}

func flagIsEnabled(flags *pflag.FlagSet, name string) bool {
	value, err := flags.GetBool(name)
	if err != nil {
		log.Fatalf(`The flag %q is not defined`, name)
	}
	return value
}
