package helmclient

import (
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
)

// TODO: There is a lot of duplication here.
// compose install and upgrade flags into install/upgrde flags
type UpgradeFlags struct {
	DryRun  bool
	Install bool
	Version string
	*GlobalFlags
}

// NewUpgradeFlags returns an instance of UpgradeFlags
// It sets up the global flags before setting up the local flags
func NewUpgradeFlags(flagmap FlagMap) (*UpgradeFlags, error) {
	globalFlags, err := NewGlobalFlags(flagmap)
	if err != nil {
		return nil, err
	}

	flags := &UpgradeFlags{GlobalFlags: globalFlags}
	if err := flags.update(flagmap); err != nil {
		return nil, err
	}

	return flags, nil
}

// update method updates the UpgradeFlags with local flags
// TODO use list of setters
func (uflags *UpgradeFlags) update(flagmap FlagMap) error {
	if err := uflags.setDryRunFlag(flagmap); err != nil {
		return err
	}

	if err := uflags.setInstallFlag(flagmap); err != nil {
		return err
	}

	if err := uflags.setVersionFlag(flagmap); err != nil {
		return err
	}

	return nil
}

// dryRunFlagSetter validates and sets the DryRun flag in the Flag struct
func (uflags *UpgradeFlags) setDryRunFlag(flagmap FlagMap) error {
	dryRun, ok := flagmap["dry-run"]

	if ok {
		dryRun, ok := dryRun.(bool)
		if !ok {
			return &InvalidFlagValueError{FlagName: "dry-run"}
		}

		uflags.DryRun = dryRun
	}

	return nil
}

// installFlagSetter validates and sets the DryRun flag in the Flag struct
func (uflags *UpgradeFlags) setInstallFlag(flagmap FlagMap) error {
	install, ok := flagmap["install"]

	if ok {
		install, ok := install.(bool)
		if !ok {
			return &InvalidFlagValueError{FlagName: "install"}
		}

		uflags.Install = install
	}

	return nil
}

// versionFlagSetter validates and sets the DryRun flag in the Flag struct
func (uflags *UpgradeFlags) setVersionFlag(flagmap FlagMap) error {
	version, ok := flagmap["version"]

	if ok {
		version, ok := version.(string)
		if !ok {
			return &InvalidFlagValueError{FlagName: "version"}
		}

		uflags.Version = version
	}

	return nil
}

// UpgradeResult represents the results for an upgrade action.
// It contains the underlying Release
type UpgradeResult struct {
	release *release.Release
	Data    string
	Status  string
}

// Upgrader handles state for an upgrade action
// It defines a single public Run method to execute the install
type Upgrader struct {
	ReleaseName string
	ChartName   string
	Flags       *UpgradeFlags
}

// NewUpgrader returns a new instance of Upgrader struct
func NewUpgrader(releaseName string, chartName string, flagmap FlagMap) (*Upgrader, error) {
	flags, err := NewUpgradeFlags(flagmap)

	if err != nil {
		return nil, err
	}

	return &Upgrader{
		ReleaseName: releaseName,
		ChartName:   chartName,
		Flags:       flags,
	}, nil
}

func (upgrader *Upgrader) newUpgradeAction(actionconfig *ActionConfig) *action.Upgrade {
	upgrade := action.NewUpgrade(actionconfig.Configuration)
	upgrade.Namespace = upgrader.Flags.Namespace
	upgrade.DryRun = upgrader.Flags.DryRun
	upgrade.Install = upgrader.Flags.Install
	upgrade.Version = upgrader.Flags.Version
	return upgrade
}

func (upgrader *Upgrader) Run(values Values) (*UpgradeResult, error) {
	envconfig := NewEnvConfig(upgrader.Flags.GlobalFlags)
	actionconfig := NewActionConfig(envconfig, upgrader.Flags.GlobalFlags)
	upgrade := upgrader.newUpgradeAction(actionconfig)

	chart, err := upgrader.LoadChart(upgrade, envconfig)
	if err != nil {
		return nil, err
	}

	// Install the release first if install is set to true
	// TODO: Add check for history
	if upgrader.Flags.Install {
		installer, err := NewInstaller(
			upgrader.ReleaseName,
			upgrader.ChartName,
			upgrader.Flags.flagmap,
		)

		if err != nil {
			return nil, err
		}

		if _, err := installer.Run(values); err != nil {
			return nil, err
		}
	}

	fmt.Printf("+%v", values)

	release, err := upgrade.Run(upgrader.ReleaseName, chart, values)
	if err != nil {
		return nil, err
	}

	return upgrader.wrappedResult(release, upgrade), nil
}

// LoadChart loads the chart with the given name
func (upgrader *Upgrader) LoadChart(upgrade *action.Upgrade, envconfig *EnvConfig) (*chart.Chart, error) {
	cp, err := upgrade.LocateChart(upgrader.ChartName, envconfig.EnvSettings)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

func (upgrader *Upgrader) wrappedResult(release *release.Release, upgrade *action.Upgrade) *UpgradeResult {
	result := &UpgradeResult{
		release: release,
		Status:  release.Info.Status.String(),
	}

	if upgrade.DryRun {
		result.Data = release.Manifest
	}

	return result
}
