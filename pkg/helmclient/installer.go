package helmclient

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
)

// InstallFlags encapsulates all install specific flags.
// To parse the install specific flags properly, flagsetter must be defined
// for the respective flags(or a consolidated flag setter).
type InstallFlags struct {
	DryRun bool
	*GlobalFlags
}

// NewInstallFlags returns an instance of InstallFlags
// It sets up the global flags before setting up the local flags
func NewInstallFlags(flagmap FlagMap) (*InstallFlags, error) {
	globalFlags, err := NewGlobalFlags(flagmap)
	if err != nil {
		return nil, err
	}

	flags := &InstallFlags{GlobalFlags: globalFlags}
	if err := flags.update(flagmap); err != nil {
		return nil, err
	}

	return flags, nil
}

// update method updates the InstallFlags with local flags
func (iflags *InstallFlags) update(flagmap FlagMap) error {
	return iflags.setDryRunFlag(flagmap)
}

// dryRunFlagSetter validates and sets the DryRun flag in the Flag struct
func (iflags *InstallFlags) setDryRunFlag(flagmap FlagMap) error {
	dryRun, ok := flagmap["dry-run"]

	if ok {
		dryRun, ok := dryRun.(bool)
		if !ok {
			return &InvalidFlagValueError{FlagName: "dry-run"}
		}

		iflags.DryRun = dryRun
	}

	return nil
}

// InstallResult represents the results for an install action.
// It contains the underlying Release
type InstallResult struct {
	release *release.Release
	Data    string
	Status  string
}

// Installer handles state for an install action
// It defines a single public Run method to execute the install
type Installer struct {
	ReleaseName string
	ChartName   string
	Flags       *InstallFlags
}

// NewInstaller returns a new instance of Installer struct
func NewInstaller(releaseName string, chartName string, flagmap FlagMap) (*Installer, error) {
	flags, err := NewInstallFlags(flagmap)

	if err != nil {
		return nil, err
	}

	return &Installer{
		ReleaseName: releaseName,
		ChartName:   chartName,
		Flags:       flags,
	}, nil
}

// newInstallAction returns a new instance of action.Install based on the action config
func (installer *Installer) newInstallAction(actionconfig *ActionConfig) *action.Install {
	install := action.NewInstall(actionconfig.Configuration)
	install.ReleaseName = installer.ReleaseName
	install.Namespace = installer.Flags.Namespace
	install.DryRun = installer.Flags.DryRun
	return install
}

// Run runs the helm install action. It creates appropriate env and install configs,
// populates the appropriate install options, loads the chart and executes the
// final install operation.
func (installer *Installer) Run(values Values) (*InstallResult, error) {
	envconfig := NewEnvConfig(installer.Flags.GlobalFlags)
	actionconfig := NewActionConfig(envconfig, installer.Flags.GlobalFlags)
	install := installer.newInstallAction(actionconfig)

	chart, err := installer.LoadChart(install, envconfig)
	if err != nil {
		return nil, err
	}

	release, err := install.Run(chart, values)
	if err != nil {
		return nil, err
	}

	return wrappedResult(release, install), nil
}

// LoadChart loads the chart with the given name
func (installer *Installer) LoadChart(install *action.Install, envconfig *EnvConfig) (*chart.Chart, error) {
	cp, err := install.LocateChart(installer.ChartName, envconfig.EnvSettings)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

// wrappedResult returns an InstallResult instance
// This can be moved elsewhere
func wrappedResult(release *release.Release, install *action.Install) *InstallResult {
	result := &InstallResult{
		release: release,
		Status:  release.Info.Status.String(),
	}

	if install.DryRun {
		result.Data = release.Manifest
	}

	return result
}

// InstallRunner defines the minimal contract for an installer
// TODO: Define it somewhere else or and remove it if not required
type InstallRunner interface {
	Run() (*release.Release, error)
}
