package envmodel

type Option struct {
	AppName string
	DotEnv  bool
	Namespace string
}

func getOption(option ...*Option) *Option {
	if len(option) > 0 {
		return option[0]
	}
	return &Option{}
}
