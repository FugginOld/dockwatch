package meta

var (
	// Version is the compile-time set version of Dockwatch
	Version = "dev"

	// UserAgent is the http client identifier derived from Version
	UserAgent string
)

func init() {
	UserAgent = "Dockwatch/" + Version
}
