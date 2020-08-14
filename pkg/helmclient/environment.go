package helmclient

import (
	"helm.sh/helm/v3/pkg/cli"
)

// EnvConfigHandler serves as a proxy to cli.EnvSettings.
// The methods on this struct take care of updating the EnvSettings struct
// with appropriate values
type EnvConfigHandler struct {
	*cli.EnvSettings
}

func NewEnvConfigHandler() *EnvConfigHandler {
	return &EnvConfigHandler{
		cli.New(),
	}
}

// WithFlags sets the appropriate config members corresponding to the flags argument
// There is gotacha here, the EnvSettings does not expose the namespace as a publicly
// writable field and takes it from the environment. The problem here is that we cannot
// set the namespace here, which means that the namespace needs to be set in individual actions.
func (config *EnvConfigHandler) SetEnvFlags(flags Flags) {
	if kubectx, ok := flags["kube-context"]; ok {
		config.KubeContext = kubectx.(string)
	}
}
