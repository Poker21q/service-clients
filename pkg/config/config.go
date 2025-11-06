package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	dotEnvPath = "./.env"
)

type Config struct {
	auth     Auth
	server   Server
	postgres Postgres
}

func (c Config) Server() Server {
	return c.server
}

func (c Config) Postgres() Postgres {
	return c.postgres
}

func (c Config) Auth() Auth { return c.auth }

func Load() (config Config, err error) {
	cfg, err := loadFromDotEnv()
	if err != nil {
		return Config{}, fmt.Errorf("failed to load config from .env: %w", err)
	}

	cfg, err = overrideFromCommandLineFlags(cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to override config from flags: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func loadFromDotEnv() (Config, error) {
	if err := godotenv.Load(dotEnvPath); err != nil {
		return Config{}, fmt.Errorf("failed to load .env: %w", err)
	}

	config := Config{
		auth: Auth{
			secret: os.Getenv("AUTH_SECRET"),
		},
		server: Server{
			host: os.Getenv("SERVER_HOST"),
			port: os.Getenv("SERVER_PORT"),
		},
		postgres: Postgres{
			host:     os.Getenv("POSTGRES_HOST"),
			port:     os.Getenv("POSTGRES_PORT"),
			user:     os.Getenv("POSTGRES_USER"),
			password: os.Getenv("POSTGRES_PASSWORD"),
			db:       os.Getenv("POSTGRES_DB"),
		},
	}

	return config, nil
}

func overrideFromCommandLineFlags(baseConfig Config) (Config, error) {
	authSecret := flag.String("auth-secret", baseConfig.auth.secret, "auth secret")
	serverHost := flag.String("server-host", baseConfig.server.host, "Server host")
	serverPort := flag.String("server-port", baseConfig.server.port, "Server port")
	postgresHost := flag.String("postgres-host", baseConfig.postgres.host, "PostgreSQL host")
	postgresPort := flag.String("postgres-port", baseConfig.postgres.port, "PostgreSQL port")
	postgresUser := flag.String("postgres-user", baseConfig.postgres.user, "PostgreSQL user")
	postgresPassword := flag.String("postgres-password", baseConfig.postgres.password, "PostgreSQL password")
	postgresDB := flag.String("postgres-db", baseConfig.postgres.db, "PostgreSQL database name")

	flag.Parse()

	config := Config{
		auth: Auth{
			secret: *authSecret,
		},
		server: Server{
			host: *serverHost,
			port: *serverPort,
		},
		postgres: Postgres{
			host:     *postgresHost,
			port:     *postgresPort,
			user:     *postgresUser,
			password: *postgresPassword,
			db:       *postgresDB,
		},
	}

	return config, nil
}

func validateConfig(cfg Config) error {
	if cfg.server.host == "" {
		return fmt.Errorf("server host is required")
	}
	if cfg.server.port == "" {
		return fmt.Errorf("server port is required")
	}
	if cfg.postgres.host == "" {
		return fmt.Errorf("postgres host is required")
	}
	if cfg.postgres.port == "" {
		return fmt.Errorf("postgres port is required")
	}
	if cfg.postgres.user == "" {
		return fmt.Errorf("postgres user is required")
	}
	if cfg.postgres.password == "" {
		return fmt.Errorf("postgres password is required")
	}
	if cfg.postgres.db == "" {
		return fmt.Errorf("postgres database name is required")
	}
	return nil
}
