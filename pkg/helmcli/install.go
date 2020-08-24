package helmcli

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

type installer struct {
	action      *action.Install
	envSettings *cli.EnvSettings
}

func (i *installer) Install(ctx context.Context, relName, chartName string, values map[string]interface{}) (*release.Release, error) {
	i.action.ReleaseName = relName

	chart, err := i.loadChart(chartName)
	if err != nil {
		return nil, err
	}

	return i.action.Run(chart, values)
}

func (i *installer) loadChart(chartName string) (*chart.Chart, error) {
	cp, err := i.action.LocateChart(chartName, i.envSettings)
	if err != nil {
		return nil, err
	}

	return loader.Load(cp)
}
