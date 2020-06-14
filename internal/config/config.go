package config

import "github.com/kelseyhightower/envconfig"

// Config is the application's runtime environment.
type Config struct {
	Port       int    `envconfig:"port" required:"true"`
	PGPort     int    `envconfig:"pg_port" required:"true"`
	PGHost     string `envconfig:"pg_host" required:"true"`
	PGDatabase string `envconfig:"pg_database" required:"true"`
	PGUser     string `envconfig:"pg_user" required:"true"`
	PGPassword string `envconfig:"pg_password" required:"true"`
	PGSSLMode  string `envconfig:"pg_sslmode" required:"true"`
}

// New returns a new Config instance.
func New() (*Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return nil, err
	}
	return &config, nil
}
