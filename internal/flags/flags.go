package flags

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// DockerAPIMinVersion is the minimum version of the docker api required to
// use dockwatch
const DockerAPIMinVersion string = "1.25"

var defaultInterval = int((time.Hour * 24).Seconds())

// RegisterDockerFlags that are used directly by the docker api client
func RegisterDockerFlags(rootCmd *cobra.Command) {
	flags := rootCmd.PersistentFlags()
	flags.StringP("host", "H", envString("DOCKER_HOST"), "daemon socket to connect to")
	flags.BoolP("tlsverify", "v", envBool("DOCKER_TLS_VERIFY"), "use TLS and verify the remote")
	flags.StringP("api-version", "a", envString("DOCKER_API_VERSION"), "api version to use by docker client")
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
		envInt("DOCKWATCH_POLL_INTERVAL"),
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
		envDuration("DOCKWATCH_TIMEOUT"),
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
		viper.GetString("DOCKWATCH_LOG_FORMAT"),
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
		viper.IsSet("NO_COLOR"),
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
		envString("DOCKWATCH_LOG_LEVEL"),
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
	viper.MustBindEnv(key)
	return viper.GetString(key)
}

func envInt(key string) int {
	viper.MustBindEnv(key)
	return viper.GetInt(key)
}

func envBool(key string) bool {
	viper.MustBindEnv(key)
	return viper.GetBool(key)
}

func envDuration(key string) time.Duration {
	viper.MustBindEnv(key)
	return viper.GetDuration(key)
}

// SetDefaults provides default values for environment variables
func SetDefaults() {
	viper.AutomaticEnv()
	viper.SetDefault("DOCKER_HOST", "unix:///var/run/docker.sock")
	viper.SetDefault("DOCKER_API_VERSION", DockerAPIMinVersion)
	viper.SetDefault("DOCKWATCH_POLL_INTERVAL", defaultInterval)
	viper.SetDefault("DOCKWATCH_TIMEOUT", time.Second*10)
	viper.SetDefault("DOCKWATCH_LOG_LEVEL", "info")
	viper.SetDefault("DOCKWATCH_LOG_FORMAT", "auto")
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
	if err = setEnvOptBool("DOCKER_TLS_VERIFY", tls); err != nil {
		return err
	}
	if err = setEnvOptStr("DOCKER_API_VERSION", version); err != nil {
		return err
	}
	if err = setEnvOptBool("DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY", registryTLSSkipVerify); err != nil {
		return err
	}
	return nil
}

// ReadFlags reads common flags used in the main program flow of dockwatch
func ReadFlags(cmd *cobra.Command) (bool, bool, bool, time.Duration) {
	flags := cmd.PersistentFlags()

	var err error
	var cleanup bool
	var noRestart bool
	var monitorOnly bool
	var timeout time.Duration

	if cleanup, err = flags.GetBool("cleanup"); err != nil {
		log.Fatal(err)
	}
	if noRestart, err = flags.GetBool("no-restart"); err != nil {
		log.Fatal(err)
	}
	if monitorOnly, err = flags.GetBool("monitor-only"); err != nil {
		log.Fatal(err)
	}
	if timeout, err = flags.GetDuration("stop-timeout"); err != nil {
		log.Fatal(err)
	}

	return cleanup, noRestart, monitorOnly, timeout
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

func setEnvOptBool(env string, opt bool) error {
	if opt {
		return setEnvOptStr(env, "1")
	}
	return nil
}

// GetSecretsFromFiles checks if passwords/tokens/webhooks have been passed as a file instead of plaintext.
// If so, the value of the flag will be replaced with the contents of the file.
func GetSecretsFromFiles(rootCmd *cobra.Command) {
	flags := rootCmd.PersistentFlags()

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
	_, err := os.Stat(s)
	return !errors.Is(err, os.ErrNotExist)
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
