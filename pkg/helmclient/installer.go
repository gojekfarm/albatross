package helmclient

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
)

// InstallResult represents the results for an install action.
// It contains the underlying Release
type InstallResult struct {
	helmrelease *release.Release
	Release     *Release
	Data        string
	Status      string
}

// NewInstallResult returns a new instance of InstallResult
func NewInstallResult(release *release.Release, installer *Installer) *InstallResult {
	result := &InstallResult{
		helmrelease: release,
		Release:     NewRelease(release),
		Status:      release.Info.Status.String(),
	}

	if installer.operation.Flags.DryRun {
		result.Data = release.Manifest
	}

	return result
}

// Installer handles state for an install action
// It defines a single public Run method to execute the install
type Installer struct {
	operation *InstallOperation
}

// NewInstaller returns a new instance of Installer struct
func NewInstaller(operation *InstallOperation) *Installer {
	return &Installer{
		operation: operation,
	}
}

// newInstallAction returns a new instance of action.Install based on the action config
func (installer *Installer) newInstallAction(actionconfig *ActionConfig) *action.Install {
	install := action.NewInstall(actionconfig.Configuration)
	install.ReleaseName = installer.operation.Name
	install.Namespace = installer.operation.Flags.Namespace
	install.DryRun = installer.operation.Flags.DryRun
	return install
}

// Run runs the helm install action. It creates appropriate env and install configs,
// populates the appropriate install options, loads the chart and executes the
// final install operation.
func (installer *Installer) Run() (*InstallResult, error) {
	flags := installer.operation.Flags
	envconfig := NewEnvConfig(flags.GlobalFlags)
	actionconfig := NewActionConfig(envconfig, flags.GlobalFlags)
	install := installer.newInstallAction(actionconfig)

	chart, err := installer.LoadChart(install, envconfig)
	if err != nil {
		return nil, err
	}

	release, err := install.Run(chart, installer.operation.Values)
	if err != nil {
		return nil, err
	}

	return NewInstallResult(release, installer), nil
}

// LoadChart returns the chart with the given name
func (installer *Installer) LoadChart(install *action.Install, envconfig *EnvConfig) (*chart.Chart, error) {
	cp, err := install.LocateChart(installer.operation.Chart, envconfig.EnvSettings)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	return chart, nil
}
