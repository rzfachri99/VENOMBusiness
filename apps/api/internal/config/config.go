package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Env            string
	HTTPAddr       string
	DatabaseDriver string
	DatabaseURL    string
	AppOrigin      string
	CookieSecure   bool
}

func Load() (Config, error) {
	cfg := Config{
		Env:            strings.ToLower(env("VENOM_ENV", "development")),
		HTTPAddr:       env("VENOM_HTTP_ADDR", ":8080"),
		DatabaseDriver: strings.ToLower(env("VENOM_DATABASE_DRIVER", "sqlite")),
		DatabaseURL:    env("VENOM_DATABASE_URL", "file:venombusiness.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"),
		AppOrigin:      strings.TrimRight(env("VENOM_APP_ORIGIN", "http://localhost:5173"), "/"),
	}
	cfg.CookieSecure = strings.EqualFold(env("VENOM_COOKIE_SECURE", "false"), "true")
	if cfg.Env != "development" && cfg.Env != "production" && cfg.Env != "test" {
		return Config{}, fmt.Errorf("VENOM_ENV must be development, test, or production")
	}
	if cfg.DatabaseDriver != "sqlite" && cfg.DatabaseDriver != "postgres" {
		return Config{}, fmt.Errorf("VENOM_DATABASE_DRIVER must be sqlite or postgres")
	}
	if cfg.DatabaseDriver == "postgres" && !strings.HasPrefix(cfg.DatabaseURL, "postgres://") && !strings.HasPrefix(cfg.DatabaseURL, "postgresql://") {
		return Config{}, fmt.Errorf("VENOM_DATABASE_URL must be a postgres:// URL when using postgres")
	}
	u, err := url.Parse(cfg.AppOrigin)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return Config{}, fmt.Errorf("VENOM_APP_ORIGIN must be an absolute URL")
	}
	if cfg.Env == "production" {
		if !cfg.CookieSecure {
			return Config{}, fmt.Errorf("VENOM_COOKIE_SECURE must be true in production")
		}
		if u.Scheme != "https" {
			return Config{}, fmt.Errorf("VENOM_APP_ORIGIN must use https in production")
		}
		if cfg.DatabaseDriver == "sqlite" && strings.Contains(cfg.DatabaseURL, "mode=memory") {
			return Config{}, fmt.Errorf("in-memory SQLite is not allowed in production")
		}
		if cfg.DatabaseDriver == "postgres" {
			du, _ := url.Parse(cfg.DatabaseURL)
			if strings.EqualFold(du.Query().Get("sslmode"), "disable") {
				return Config{}, fmt.Errorf("PostgreSQL sslmode=disable is not allowed in production")
			}
		}
	}
	return cfg, nil
}
func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
