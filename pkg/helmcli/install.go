package helmcli

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli/config"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

type Installer struct {
	action      *action.Install
	envSettings *cli.EnvSettings
}

// NewInstaller returns a new instance of Installer struct
func NewInstaller(flg flags.InstallFlags) *Installer {
	envconfig := config.NewEnvConfig(&flg.GlobalFlags)
	actionconfig := config.NewActionConfig(envconfig, &flg.GlobalFlags)

	install := action.NewInstall(actionconfig.Configuration)
	install.Namespace = flg.Namespace
	install.DryRun = flg.DryRun

	return &Installer{
		action:      install,
		envSettings: envconfig.EnvSettings,
	}
}

func (i *Installer) Install(ctx context.Context, relName, chartName string, values map[string]interface{}) (*release.Release, error) {
	i.action.ReleaseName = relName

	chart, err := i.loadChart(chartName)
	if err != nil {
		return nil, err
	}

	return i.action.Run(chart, values)
}

func (i *Installer) loadChart(chartName string) (*chart.Chart, error) {
	cp, err := i.action.LocateChart(chartName, i.envSettings)
	if err != nil {
		return nil, err
	}

	return loader.Load(cp)
}
