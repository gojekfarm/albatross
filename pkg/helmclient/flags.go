package helmclient

// GlobalFlags is an inventory of all supported flags.
// It exposes methods to validate and get specific flag values.
type GlobalFlags struct {
	KubeCtx       string
	KubeToken     string
	KubeAPIServer string
	Namespace     string
}

// InstallUpgradeFlags defines flags that are common to both install and upgrade actions
type InstallUpgradeFlags struct {
	DryRun  bool `json:"dry-run,omitempty"`
	Version string
	*GlobalFlags
}

// NewInstallUpgradeFlags return a new instance of InstallUpgradeflags.
// Ideally, this need to be directly used. Instead this is references in the concrete
// install and upgrade flag stuct
func NewInstallUpgradeFlags() *InstallUpgradeFlags {
	return &InstallUpgradeFlags{
		GlobalFlags: &GlobalFlags{},
	}
}

// InstallFlags defines all flags specific to install action
type InstallFlags struct {
	*InstallUpgradeFlags
}

// NewInstallFlags returns an instance of InstallFlags
func NewInstallFlags() *InstallFlags {
	return &InstallFlags{
		InstallUpgradeFlags: NewInstallUpgradeFlags(),
	}
}

// UpgradeFlags defines all flags specific to upgrade action
type UpgradeFlags struct {
	Install bool
	*InstallUpgradeFlags
}

// NewUpgradeFlags returns an instance of UpgradeFlags
func NewUpgradeFlags() *UpgradeFlags {
	return &UpgradeFlags{
		InstallUpgradeFlags: NewInstallUpgradeFlags(),
	}
}
