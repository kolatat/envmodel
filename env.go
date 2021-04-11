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
}

func NewEnvironment(option ...*Option) *Environment {
	opt := getOption(option...)

	var env Environment
	env.EnvKey = EnvKey
	if "" == opt.AppName {
		env.EnvKey = strings.ToUpper(opt.AppName) + "_" + EnvKey
	}

	env.Env = os.Getenv(env.EnvKey)
	if "" == env.Env {
		env.Env = EnvDevelopment
	}

	return &env
}
