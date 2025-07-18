package log

type options struct {
	appName    string
	appVersion string
}

type Option func(*options)

// WithAppName sets the application name
func WithAppName(name string) Option {
	return func(o *options) {
		o.appName = name
	}
}

// WithAppVersion sets the application version
func WithAppVersion(version string) Option {
	return func(o *options) {
		o.appVersion = version
	}
}
