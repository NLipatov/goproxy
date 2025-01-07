package config

import "fmt"

type EnvVarNotSetError struct {
	envVarName string
}

func NewEnvVarNotSetError(envVarName string) EnvVarNotSetError {
	err := EnvVarNotSetError{
		envVarName: envVarName,
	}

	return err
}

func (e EnvVarNotSetError) Error() string {
	return fmt.Sprintf("%s env var not set", e.envVarName)
}
