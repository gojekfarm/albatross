package config

import (
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"

	"helm.sh/helm/v3/pkg/cli"
)

// EnvConfig serves as a proxy to cli.EnvSettings.
// The methods on this struct take care of updating the EnvSettings struct
// with appropriate values
type EnvConfig struct {
	*cli.EnvSettings
}

func NewEnvConfig(flg *flags.GlobalFlags) *EnvConfig {
	envconfig := &EnvConfig{
		cli.New(),
	}

	envconfig.setEnvFlags(flg)
	return envconfig
}

// setEnvFlags sets the appropriate config members corresponding to the flags argument
// There is gotacha here, the EnvSettings does not expose the namespace as a publicly
// writable field and takes it from the environment. The problem here is that we cannot
// set the namespace here, which means that the namespace needs to be set in individual actions.
func (config *EnvConfig) setEnvFlags(flg *flags.GlobalFlags) {
	config.KubeContext = flg.KubeContext
}
