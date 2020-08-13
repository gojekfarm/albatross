package helmclient

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
)

type Installer struct {
	*EnvConfigHandler
	*ActionConfig
	action *action.Install
	chart  *chart.Chart
}

// InstallRunner defines the minimal contract for an installer
// TODO: Define it somewhere else or and remove it if not required
type InstallRunner interface {
	Setup(name string, chart string, values Values, flags Flags)
	Run() (*release.Release, error)
}

func NewInstaller() *Installer {
	return &Installer{
		NewEnvConfigHandler(),
		NewActionConfig(),
		new(action.Install),
		new(chart.Chart),
	}
}

// NewInstall returns a new instance of the installer
func (i *Installer) Setup(name string, chartName string, flags Flags) {
	i.EnvConfigHandler.WithEnvFlags(flags)

	i.ActionConfig.WithEnvironmentFlags(i.EnvConfigHandler, flags)
	i.ActionConfig.WithBaseFlags(flags)

	i.action = action.NewInstall(i.ActionConfig.Configuration)
	i.action.ReleaseName = name
	i.SetFlags(flags)

	chart, err := i.loadChart(chartName)
	if err != nil {
		// TODO: Remove panic later
		panic("Failed to read the chart")
	}

	i.chart = chart
}

// loadChart returns the loaded chart.Chart instance
func (i *Installer) loadChart(name string) (*chart.Chart, error) {
	cp, err := i.action.LocateChart(name, i.EnvConfigHandler.EnvSettings)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

// SetFlags updates the install action with the proper flags
// TODO: Find a cleaner to way to map flag keys with flags without reflection
func (i *Installer) SetFlags(flags Flags) {
	if namespace, ok := flags["namespace"].(string); ok {
		i.action.Namespace = namespace
	}

	if dryRun, ok := flags["dry-run"].(bool); ok {
		i.action.DryRun = dryRun
	}
}

func (i *Installer) Run(values Values) (*release.Release, error) {
	return i.action.Run(i.chart, values)
}
