package flags

type GlobalFlags struct {
	KubeCtx       string
	KubeToken     string
	KubeAPIServer string
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
