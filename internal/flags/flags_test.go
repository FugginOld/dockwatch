package flags

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvConfig_Defaults(t *testing.T) {
	// Unset testing environments own variables, since those are not what is under test
	_ = os.Unsetenv("DOCKER_TLS_VERIFY")
	_ = os.Unsetenv("DOCKER_HOST")

	cmd := new(cobra.Command)
	SetDefaults()
	RegisterDockerFlags(cmd)

	err := EnvConfig(cmd)
	require.NoError(t, err)

	assert.Equal(t, "unix:///var/run/docker.sock", os.Getenv("DOCKER_HOST"))
	assert.Equal(t, "", os.Getenv("DOCKER_TLS_VERIFY"))
	// Re-enable this test when we've moved to github actions.
	// assert.Equal(t, DockerAPIMinVersion, os.Getenv("DOCKER_API_VERSION"))
}

func TestEnvConfig_Custom(t *testing.T) {
	cmd := new(cobra.Command)
	SetDefaults()
	RegisterDockerFlags(cmd)

	err := cmd.ParseFlags([]string{"--host", "some-custom-docker-host", "--tlsverify", "--api-version", "1.99"})
	require.NoError(t, err)

	err = EnvConfig(cmd)
	require.NoError(t, err)

	assert.Equal(t, "some-custom-docker-host", os.Getenv("DOCKER_HOST"))
	assert.Equal(t, "1", os.Getenv("DOCKER_TLS_VERIFY"))
	// Re-enable this test when we've moved to github actions.
	// assert.Equal(t, "1.99", os.Getenv("DOCKER_API_VERSION"))
}

func TestGetSecretsFromFilesWithString(t *testing.T) {
	value := "supersecretstring"
	t.Setenv("DOCKWATCH_HTTP_API_TOKEN", value)

	testGetSecretsFromFiles(t, "http-api-token", value)
}

func TestGetSecretsFromFilesWithFile(t *testing.T) {
	value := "megasecretstring"

	// Create the temporary file which will contain a secret.
	file, err := os.CreateTemp(t.TempDir(), "dockwatch-")
	require.NoError(t, err)

	// Write the secret to the temporary file.
	_, err = file.Write([]byte(value))
	require.NoError(t, err)
	require.NoError(t, file.Close())

	t.Setenv("DOCKWATCH_HTTP_API_TOKEN", file.Name())

	testGetSecretsFromFiles(t, "http-api-token", value)
}

func testGetSecretsFromFiles(t *testing.T, flagName string, expected string, args ...string) {
	cmd := new(cobra.Command)
	SetDefaults()
	RegisterSystemFlags(cmd)
	require.NoError(t, cmd.ParseFlags(args))
	GetSecretsFromFiles(cmd)
	flag := cmd.PersistentFlags().Lookup(flagName)
	require.NotNil(t, flag)
	value := flag.Value.String()

	assert.Equal(t, expected, value)
}

func TestHTTPAPIPeriodicPollsFlag(t *testing.T) {
	cmd := new(cobra.Command)
	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)

	err := cmd.ParseFlags([]string{"--http-api-periodic-polls"})
	require.NoError(t, err)

	periodicPolls, err := cmd.PersistentFlags().GetBool("http-api-periodic-polls")
	require.NoError(t, err)

	assert.Equal(t, true, periodicPolls)
}

func TestIsFile(t *testing.T) {
	assert.False(t, isFile("https://google.com"), "an URL should never be considered a file")
	assert.True(t, isFile(os.Args[0]), "the currently running binary path should always be considered a file")
}

func TestProcessFlagAliases(t *testing.T) {
	logrus.StandardLogger().ExitFunc = func(_ int) { t.FailNow() }
	cmd := new(cobra.Command)
	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)

	require.NoError(t, cmd.ParseFlags([]string{
		`--porcelain`, `v1`,
		`--interval`, `10`,
		`--trace`,
	}))
	flags := cmd.Flags()
	ProcessFlagAliases(flags)

	sched, _ := flags.GetString(`schedule`)
	assert.Equal(t, `@every 10s`, sched)

	logLevel, _ := flags.GetString(`log-level`)
	assert.Equal(t, `trace`, logLevel)
}

func TestProcessFlagAliasesLogLevelFromEnvironment(t *testing.T) {
	cmd := new(cobra.Command)
	t.Setenv("DOCKWATCH_DEBUG", `true`)

	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)

	require.NoError(t, cmd.ParseFlags([]string{}))
	flags := cmd.Flags()
	ProcessFlagAliases(flags)

	logLevel, _ := flags.GetString(`log-level`)
	assert.Equal(t, `debug`, logLevel)
}

func TestLogFormatFlag(t *testing.T) {
	cmd := new(cobra.Command)

	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)

	// Ensure the default value is Auto
	require.NoError(t, cmd.ParseFlags([]string{}))
	require.NoError(t, SetupLogging(cmd.Flags()))
	assert.IsType(t, &logrus.TextFormatter{}, logrus.StandardLogger().Formatter)

	// Test JSON format
	require.NoError(t, cmd.ParseFlags([]string{`--log-format`, `JSON`}))
	require.NoError(t, SetupLogging(cmd.Flags()))
	assert.IsType(t, &logrus.JSONFormatter{}, logrus.StandardLogger().Formatter)

	// Test Pretty format
	require.NoError(t, cmd.ParseFlags([]string{`--log-format`, `pretty`}))
	require.NoError(t, SetupLogging(cmd.Flags()))
	assert.IsType(t, &logrus.TextFormatter{}, logrus.StandardLogger().Formatter)
	textFormatter, ok := (logrus.StandardLogger().Formatter).(*logrus.TextFormatter)
	assert.True(t, ok)
	assert.True(t, textFormatter.ForceColors)
	assert.False(t, textFormatter.FullTimestamp)

	// Test LogFmt format
	require.NoError(t, cmd.ParseFlags([]string{`--log-format`, `logfmt`}))
	require.NoError(t, SetupLogging(cmd.Flags()))
	textFormatter, ok = (logrus.StandardLogger().Formatter).(*logrus.TextFormatter)
	assert.True(t, ok)
	assert.True(t, textFormatter.DisableColors)
	assert.True(t, textFormatter.FullTimestamp)

	// Test invalid format
	require.NoError(t, cmd.ParseFlags([]string{`--log-format`, `cowsay`}))
	require.Error(t, SetupLogging(cmd.Flags()))
}

func TestLogLevelFlag(t *testing.T) {
	cmd := new(cobra.Command)

	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)

	// Test invalid format
	require.NoError(t, cmd.ParseFlags([]string{`--log-level`, `gossip`}))
	require.Error(t, SetupLogging(cmd.Flags()))
}

func TestProcessFlagAliasesSchedAndInterval(t *testing.T) {
	logrus.StandardLogger().ExitFunc = func(_ int) { panic(`FATAL`) }
	cmd := new(cobra.Command)
	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)

	require.NoError(t, cmd.ParseFlags([]string{`--schedule`, `@hourly`, `--interval`, `10`}))
	flags := cmd.Flags()

	assert.PanicsWithValue(t, `FATAL`, func() {
		ProcessFlagAliases(flags)
	})
}

func TestProcessFlagAliasesCronAlias(t *testing.T) {
	logrus.StandardLogger().ExitFunc = func(_ int) { t.FailNow() }
	cmd := new(cobra.Command)
	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)

	require.NoError(t, cmd.ParseFlags([]string{`--cron`, `@hourly`}))
	flags := cmd.Flags()
	ProcessFlagAliases(flags)

	sched, _ := flags.GetString(`schedule`)
	assert.Equal(t, `@hourly`, sched)
}

func TestProcessFlagAliasesCronAndScheduleConflict(t *testing.T) {
	logrus.StandardLogger().ExitFunc = func(_ int) { panic(`FATAL`) }
	cmd := new(cobra.Command)
	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)

	require.NoError(t, cmd.ParseFlags([]string{`--cron`, `@hourly`, `--schedule`, `@daily`}))
	flags := cmd.Flags()

	assert.PanicsWithValue(t, `FATAL`, func() {
		ProcessFlagAliases(flags)
	})
}

func TestProcessFlagAliasesForceUpdateAlias(t *testing.T) {
	logrus.StandardLogger().ExitFunc = func(_ int) { t.FailNow() }
	cmd := new(cobra.Command)
	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)

	require.NoError(t, cmd.ParseFlags([]string{`--force-update`}))
	flags := cmd.Flags()
	ProcessFlagAliases(flags)

	runOnce, _ := flags.GetBool(`run-once`)
	assert.Equal(t, true, runOnce)
}

func TestProcessFlagAliasesScheduleFromEnvironment(t *testing.T) {
	cmd := new(cobra.Command)

	t.Setenv("DOCKWATCH_SCHEDULE", `@hourly`)

	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)

	require.NoError(t, cmd.ParseFlags([]string{}))
	flags := cmd.Flags()
	ProcessFlagAliases(flags)

	sched, _ := flags.GetString(`schedule`)
	assert.Equal(t, `@hourly`, sched)
}

func TestProcessFlagAliasesInvalidPorcelaineVersion(t *testing.T) {
	logrus.StandardLogger().ExitFunc = func(_ int) { panic(`FATAL`) }
	cmd := new(cobra.Command)
	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)

	require.NoError(t, cmd.ParseFlags([]string{`--porcelain`, `cowboy`}))
	flags := cmd.Flags()

	assert.PanicsWithValue(t, `FATAL`, func() {
		ProcessFlagAliases(flags)
	})
}
