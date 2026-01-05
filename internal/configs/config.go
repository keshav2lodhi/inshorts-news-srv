package configs

import (
	"errors"
	"os"

	"inshorts.com/inshorts-news-srv/internal/base"
)

func Initialize() error {
	if os.Getenv(base.EnvLogLevel) == "" {
		_ = os.Setenv("LOG_LEVEL", "info")
	}
	if os.Getenv(base.EnvESUserName) == "" {
		return errors.New("ES_USERNAME env variable is not given")
	}
	if os.Getenv(base.EnvESPassword) == "" {
		return errors.New("ES_PASSWORD env variable is not given")
	}
	if os.Getenv(base.EnvESUrl) == "" {
		return errors.New("ES_URL env variable is not given")
	}
	if os.Getenv(base.EnvOpenRouterAPIKey) == "" {
		return errors.New("OPENROUTER_API_KEY env variable is not given")
	}
	if os.Getenv(base.EnvAPIKey) == "" {
		return errors.New("API_KEY env variable is not given")
	}
	return nil
}
