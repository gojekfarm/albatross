package flags

import "time"

type GlobalFlags struct {
	KubeContext   string `json:"kube_context,omitempty"`
	KubeToken     string `json:"kube_token,omitempty"`
	KubeAPIServer string `json:"kube_apiserver,omitempty"`
	Namespace     string
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


// UninstallFlags maps the list of options that can be passed to the helm action 
type UninstallFlags struct {
	Release string
	KeepHistory bool
	DisableHooks bool
	DryRun bool
	Timeout time.Duration
	GlobalFlags
}
