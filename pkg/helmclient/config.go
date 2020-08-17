package helmclient

import (
	"os"

	"github.com/gojekfarm/albatross/pkg/logger"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// ActionConfig acts as a proxy to helm package's action configuration.
// It defines methods to set the default/common action config members
type ActionConfig struct {
	*action.Configuration
}

// NewActionConfig returns a new instance of actionconfig
func NewActionConfig(envconfig *EnvConfig, flags *GlobalFlags) *ActionConfig {
	config := &ActionConfig{
		new(action.Configuration),
	}

	config.setFlags(envconfig, flags)
	return config
}

// kubeClientConfig returns a kube config that is scoped to a namespace.
// Context: The EnvSetting struct does not expose any way to set the namespace,
// so we cannot set it directly. However, it is used to create kubeclients.
// So in order to configure the kubeclient with the proper namespace, we define a custom getter
// here that sets the correct namespace in the kubeconfig
func kubeClientConfig(envconfig *EnvConfig, namespace string) genericclioptions.RESTClientGetter {
	clientConfig := kube.GetConfig(envconfig.KubeConfig, envconfig.KubeContext, namespace)

	if envconfig.KubeToken != "" {
		clientConfig.BearerToken = &envconfig.KubeToken
	}

	if envconfig.KubeAPIServer != "" {
		clientConfig.APIServer = &envconfig.KubeAPIServer
	}

	return clientConfig
}

// setFlags initializes the action configuration with proper config flags
func (ac *ActionConfig) setFlags(envconfig *EnvConfig, flags *GlobalFlags) {
	// TODO: Can we remove the check altogether?
	// actionNamespace := envconfig.Namespace()
	// if flags.Namespace != "" {
	// 	actionNamespace = flags.Namespace
	// }

	actionNamespace := flags.Namespace
	ac.Configuration.Init(
		kubeClientConfig(envconfig, actionNamespace),
		actionNamespace,
		os.Getenv("HELM_DRIVER"),
		logger.Debugf,
	)
}
