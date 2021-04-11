package envmodel

import (
	"os"
	"strings"
)

const (
	EnvDevelopment = "development"
	EnvTest        = "test"

	EnvKey = "ENV"
)

type Environment struct {
	EnvKey string
	Env    string

	option *Option
}

func NewEnvironment(option ...*Option) *Environment {
	opt := getOption(option...)

	var env Environment
	env.option = opt
	env.EnvKey = EnvKey
	if "" == opt.AppName {
		env.EnvKey = strings.ToUpper(opt.AppName) + "_" + EnvKey
	}

	env.Env = os.Getenv(env.EnvKey)
	if "" == env.Env {
		env.Env = EnvDevelopment
		opt.Logger.Info().Str("environment", env.Env).Msg("environment assumed")
	} else {
		opt.Logger.Info().Str("environment", env.Env).Msgf("environment set by %q", env.EnvKey)
	}

	return &env
}
