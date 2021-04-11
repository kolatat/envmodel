package envmodel

import (
	"github.com/rs/zerolog"
)

type Option struct {
	AppName   string
	DotEnv    bool
	Namespace string
	Logger    *zerolog.Logger
}

func getOption(option ...*Option) *Option {
	if len(option) > 0 {
		if option[0].Logger == nil {
			logger := zerolog.Nop()
			option[0].Logger = &logger
		}
		return option[0]
	}
	return &Option{}
}
