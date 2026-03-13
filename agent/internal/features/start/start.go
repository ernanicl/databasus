package start

import (
	"errors"
	"log/slog"

	"databasus-agent/internal/config"
)

func Run(cfg *config.Config, log *slog.Logger) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	log.Info("start: stub — not yet implemented",
		"dbId", cfg.DbID,
		"hasToken", cfg.Token != "",
	)

	return nil
}

func validateConfig(cfg *config.Config) error {
	if cfg.DatabasusHost == "" {
		return errors.New("argument databasus-host is required")
	}

	if cfg.DbID == "" {
		return errors.New("argument db-id is required")
	}

	if cfg.Token == "" {
		return errors.New("argument token is required")
	}

	return nil
}
