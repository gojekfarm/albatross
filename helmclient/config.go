package helmclient

import (
	"os"

	"github.com/gojekfarm/albatross/logger"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ActionConfig struct {
	*action.Configuration
}

func NewActionConfig() *ActionConfig {
	return &ActionConfig{
		new(action.Configuration),
	}
}

// clientGetter returns a kube config that is scoped to a namespace.
// Context: The EnvSetting struct does not expose any way to set the namespace,
// so we cannot set it directly. However, it is used to create kubeclients.
// So in order to configure the kubeclient with the proper namespace, we define a custom getter
// here that sets the correct namespace in the kubeconfig
func kubeClientConfig(envconfig *EnvConfigHandler, namespace string) genericclioptions.RESTClientGetter {
	clientConfig := kube.GetConfig(envconfig.KubeConfig, envconfig.KubeContext, namespace)
	if envconfig.KubeToken != "" {
		clientConfig.BearerToken = &envconfig.KubeToken
	}
	if envconfig.KubeAPIServer != "" {
		clientConfig.APIServer = &envconfig.KubeAPIServer
	}

	return clientConfig
}

// WithEnvironment initializes the action configuration with the environment values
func (ac *ActionConfig) WithEnvironmentFlags(envconfig *EnvConfigHandler, flags Flags) {
	actionNamespace := envconfig.Namespace()
	if namespace, ok := flags["namespace"].(string); ok {
		actionNamespace = namespace
	}

	ac.Configuration.Init(
		kubeClientConfig(envconfig, actionNamespace),
		actionNamespace,
		os.Getenv("HELM_DRIVER"),
		logger.Debug,
	)

	ac.WithBaseFlags(flags)
}

// WithBaseFlags updates the action config with a base set of flags common to all actions
func (ac *ActionConfig) WithBaseFlags(flags Flags) {
}
