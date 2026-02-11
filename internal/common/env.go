package common

var EnvironmentConfig Environment //nolint:gochecknoglobals

type Environment struct {
	CurrentEnvironment string `env:"ENVIRONMENT" envDefault:"production"`
}

func (e *Environment) IsTest() bool {
	return e.CurrentEnvironment == "test"
}
