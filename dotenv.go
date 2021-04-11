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

func suppressOnLoad(err error) error {
	if err == nil {
		return nil
	}
	// all nbd errors can be ignored here
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func LoadDotEnv(environment ...*Environment) error {
	var env *Environment
	if len(environment) > 0 {
		env = environment[0]
	} else {
		env = NewEnvironment()
	}

	fileEnv := "." + env.Env

	if err := suppressOnLoad(godotenv.Load(fileRoot + fileEnv + fileLocal)); err != nil {
		return err
	}
	// the not EnvTest check is present on joho/godotenv docs, idk why
	// if EnvTest != env.Env {
	if err := suppressOnLoad(godotenv.Load(fileRoot + fileLocal)); err != nil {
		return err
	}
	// }
	if err := suppressOnLoad(godotenv.Load(fileRoot + fileEnv)); err != nil {
		return err
	}
	if err := suppressOnLoad(godotenv.Load()); err != nil {
		return err
	}

	return nil
}
