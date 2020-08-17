package helmclient

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// UpgradeResult represents the results for an upgrade action.
// It contains the underlying Release
type UpgradeResult struct {
	helmrelease *release.Release
	Release     *Release
	Data        string
	Status      string
}

// NewUpgradeResult returns a new instance of UpgradeResult
func NewUpgradeResult(release *release.Release, upgrader *Upgrader) *UpgradeResult {
	result := &UpgradeResult{
		helmrelease: release,
		Release:     NewRelease(release),
		Status:      release.Info.Status.String(),
	}

	if upgrader.operation.Flags.DryRun {
		result.Data = release.Manifest
	}

	return result
}

// Upgrader handles state for an upgrade action
// It defines a single public Run method to execute the install
type Upgrader struct {
	operation *UpgradeOperation
}

// NewUpgrader returns a new instance of Upgrader struct
func NewUpgrader(operation *UpgradeOperation) *Upgrader {
	return &Upgrader{
		operation: operation,
	}
}

// newUpgradeAction returns an upgrade action instance using the actionconfig
func (upgrader *Upgrader) newUpgradeAction(actionconfig *ActionConfig) *action.Upgrade {
	upgrade := action.NewUpgrade(actionconfig.Configuration)
	upgrade.Namespace = upgrader.operation.Flags.Namespace
	upgrade.DryRun = upgrader.operation.Flags.DryRun
	upgrade.Install = upgrader.operation.Flags.Install
	upgrade.Version = upgrader.operation.Flags.Version
	return upgrade
}

// Run executes the upgrade action
func (upgrader *Upgrader) Run() (*UpgradeResult, error) {
	flags := upgrader.operation.Flags
	envconfig := NewEnvConfig(flags.GlobalFlags)
	actionconfig := NewActionConfig(envconfig, flags.GlobalFlags)
	upgrade := upgrader.newUpgradeAction(actionconfig)

	// Install the release first if install is set to true
	if flags.Install {
		history := action.NewHistory(actionconfig.Configuration)
		history.Max = 1
		if _, err := history.Run(upgrader.operation.Name); err == driver.ErrReleaseNotFound {
			result, err := upgrader.installRelease()
			if err != nil {
				return nil, err
			}

			return NewUpgradeResult(result.helmrelease, upgrader), nil
		} else if err != nil {
			return nil, err
		}
	}

	chart, err := upgrader.LoadChart(upgrade, envconfig)
	if err != nil {
		return nil, err
	}

	release, err := upgrade.Run(upgrader.operation.Name, chart, upgrader.operation.Values)
	if err != nil {
		return nil, err
	}

	return NewUpgradeResult(release, upgrader), nil
}

// install runs the install action before an upgrade
// This does not check if the install flag is set, caller should check the flag
func (upgrader *Upgrader) installRelease() (*InstallResult, error) {
	operation := &InstallOperation{
		Name:   upgrader.operation.Name,
		Chart:  upgrader.operation.Chart,
		Values: upgrader.operation.Values,
		Flags: &InstallFlags{
			InstallUpgradeFlags: upgrader.operation.Flags.InstallUpgradeFlags,
		},
	}

	installer := NewInstaller(operation)
	result, err := installer.Run()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// LoadChart loads the chart with the given name
func (upgrader *Upgrader) LoadChart(upgrade *action.Upgrade, envconfig *EnvConfig) (*chart.Chart, error) {
	cp, err := upgrade.LocateChart(upgrader.operation.Chart, envconfig.EnvSettings)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	return chart, nil
}
