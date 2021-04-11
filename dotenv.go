package envmodel

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

const (
	fileRoot  = ".env"
	fileLocal = ".local"
)

func LoadDotEnv(environment ...*Environment) error {
	var env *Environment
	if len(environment) > 0 {
		env = environment[0]
	} else {
		env = NewEnvironment()
	}

	suppressOnLoad := func(filename string) error {
		err := godotenv.Load(filename)
		if err == nil {
			return nil
		} else {
			env.option.Logger.Warn().Err(err).Msgf("while loading %q", filename)
		}
		// all nbd errors can be ignored here
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	fileEnv := "." + env.Env

	if err := suppressOnLoad(fileRoot + fileEnv + fileLocal); err != nil {
		return err
	}
	// the not EnvTest check is present on joho/godotenv docs, idk why
	// if EnvTest != env.Env {
	if err := suppressOnLoad(fileRoot + fileLocal); err != nil {
		return err
	}
	// }
	if err := suppressOnLoad(fileRoot + fileEnv); err != nil {
		return err
	}
	if err := suppressOnLoad(fileRoot); err != nil {
		return err
	}

	return nil
}
