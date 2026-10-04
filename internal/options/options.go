package options

type Environment string

const (
	EnvironmentProduction  Environment = "production"
	EnvironmentDevelopment Environment = "development"
)

type AuthMode string

const (
	AuthModeSingleUser      AuthMode = "single-user"
	AuthModeIdentityToken   AuthMode = "identity-token"
	AuthModeForwardedHeader AuthMode = "forwarded-header"
)

type Auth struct {
	Mode           AuthMode
	IdentitySecret string
}

type Driver string

const (
	DriverPostgres Driver = "postgres"
	DriverSQLite   Driver = "sqlite"
)

type Database struct {
	Driver Driver
	Source string
}

type Options struct {
	Environment Environment
	Auth        Auth
	Database    Database
}

type Option func(*Options)

func New(opts ...Option) *Options {
	o := &Options{
		Environment: EnvironmentProduction,
		Auth:        Auth{Mode: AuthModeSingleUser},
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func WithEnvironment(env Environment) Option {
	return func(o *Options) {
		o.Environment = env
	}
}

func WithSingleUser() Option {
	return func(o *Options) {
		o.Auth = Auth{Mode: AuthModeSingleUser}
	}
}

func WithIdentitySecret(secret string) Option {
	return func(o *Options) {
		o.Auth = Auth{Mode: AuthModeIdentityToken, IdentitySecret: secret}
	}
}

func WithForwardedHeader() Option {
	return func(o *Options) {
		o.Auth = Auth{Mode: AuthModeForwardedHeader}
	}
}

func WithPostgres(dsn string) Option {
	return func(o *Options) {
		o.Database = Database{Driver: DriverPostgres, Source: dsn}
	}
}

func WithSQLite(path string) Option {
	return func(o *Options) {
		o.Database = Database{Driver: DriverSQLite, Source: path}
	}
}

func (o *Options) IsDevelopment() bool {
	return o.Environment == EnvironmentDevelopment
}

func (o *Options) IsMultiUser() bool {
	return o.Auth.Mode != AuthModeSingleUser
}
