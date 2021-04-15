package flags

import "time"

// GlobalFlags flags which give context about kubernetes cluster to connect to
// swagger:model globalFlags
type GlobalFlags struct {
	KubeContext string `json:"-"`
	// required: false
	KubeToken string `json:"kube_token,omitempty"`
	// required: false
	KubeAPIServer string `json:"kube_apiserver,omitempty"`
	// required: true
	// example: default
	Namespace string `json:"-" schema:"namespace"`
}

type UpgradeFlags struct {
	DryRun  bool
	Install bool
	Version string
	GlobalFlags
}

type InstallFlags struct {
	DryRun  bool
	Version string
	GlobalFlags
}

type ListFlags struct {
	AllNamespaces bool
	Deployed      bool
	Failed        bool
	Pending       bool
	Uninstalled   bool
	Uninstalling  bool
	GlobalFlags
}

// UninstallFlags maps the list of options that can be passed to the helm action.
type UninstallFlags struct {
	Release      string
	KeepHistory  bool
	DisableHooks bool
	DryRun       bool
	Timeout      time.Duration
	GlobalFlags
}

type AddFlags struct {
	Name                 string
	URL                  string
	Username             string
	Password             string
	ForceUpdate          bool
	AllowDeprecatedRepos bool

	CertFile              string
	KeyFile               string
	CaFile                string
	InsecureSkipTLSverify bool

	RepoFile  string
	RepoCache string
}
