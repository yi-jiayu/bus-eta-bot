package busetabot

import (
	"os"

	"github.com/pkg/errors"
)

const (
	devEnvironment        = "dev"
	stagingEnvironment    = "staging"
	productionEnvironment = "prod"

	EnvironmentDev        = devEnvironment
	EnvironmentStaging    = stagingEnvironment
	EnvironmentProduction = productionEnvironment
)

var (
	errNotFound = errors.New("not found")
)

var namespace = GetBotEnvironment()

func GetBotEnvironment() string {
	switch os.Getenv("BOT_ENVIRONMENT") {
	case stagingEnvironment:
		return stagingEnvironment
	case productionEnvironment:
		return productionEnvironment
	default:
		return devEnvironment
	}
}
