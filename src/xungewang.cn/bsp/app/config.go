package app

import (
	"github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
)

// Config stores the application-wide configurations
var Config appConfig

type appConfig struct {
	// the server address. Defaults to ':8000'
	HttpServerAddr string `mapstructure:"http_server_addr"`

	// log level, can be one of "debug", "info" (default), "warn", "fatal", and "panic"
	LogLevel string `mapstructure:"log_level"`

	// datasource name
	DSN string `mapstructure:"db_dsn"`

	// connection pool
	DbMaxOpenConns    int `mapstructure:"db_max_open_conns"`
	DbMaxIdleConns    int `mapstructure:"db_max_idle_conns"`
	DbConnMaxLifetime int `mapstructure:"db_conn_max_lifetime"` // in seconds
}

func (config appConfig) Validate() error {
	return validation.ValidateStruct(&config,
		validation.Field(&config.HttpServerAddr, validation.Required),
		validation.Field(&config.LogLevel, validation.Required,
			validation.In("debug", "info", "warn", "warning", "fatal", "panic")),
		validation.Field(&config.DSN, validation.Required),
	)
}

// LoadConfig load configuration from given paths(directories, actually)
// "app.yaml" under given directories wil be searched.
func LoadConfig(paths ...string) error {
	// set config filename - app.yaml (which will be loaded from within given paths)
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")

	// config can be sourced from environment variables
	viper.SetEnvPrefix("bsp") // env vars are prefixed with "BSP"
	viper.AutomaticEnv()

	// setup default values (in case no value for some key)
	viper.SetDefault("http_server_addr", ":8000")
	viper.SetDefault("log_level", "info")
	// it seems that viper has a splendid weird behavior, it there's no config path given
	// (via AddConfigPath below), the Env variable won't work if there's no default value
	// for it. To overcome it, we have to set default value for 'db_dsn' here.
	viper.SetDefault("db_dsn", "")
	viper.SetDefault("db_max_open_conns", 0)
	viper.SetDefault("db_max_idle_conns", 0)
	viper.SetDefault("db_conn_max_lifetime", 0)

	// read config from paths in file system.
	if len(paths) > 0 {
		for _, path := range paths {
			viper.AddConfigPath(path)
		}
		// In case no/non-existing config path provided, ReadInConfig() gives an error (of ConfigFileNotFoundError),
		// but since we can also load config from env, we can safely ignore that error.
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}

	if err := viper.Unmarshal(&Config); err != nil {
		return err
	}

	return Config.Validate()
}
