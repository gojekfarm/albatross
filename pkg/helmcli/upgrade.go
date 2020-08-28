package helmcli

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type upgrader struct {
	action      *action.Upgrade
	history     *action.History
	envSettings *cli.EnvSettings
	installer   Installer
}

// Upgrade executes the upgrade action
func (u *upgrader) Upgrade(ctx context.Context, relName, chartName string, values map[string]interface{}) (*release.Release, error) {
	// Install the release first if install is set to true
	if u.action.Install {
		u.history.Max = 1
		if _, err := u.history.Run(relName); err == driver.ErrReleaseNotFound {
			release, err := u.installer.Install(ctx, relName, chartName, values)
			if err != nil {
				return nil, err
			}

			return release, nil
		} else if err != nil {
			return nil, err
		}
	}

	chart, err := u.loadChart(chartName)
	if err != nil {
		return nil, fmt.Errorf("error loading chart: %w", err)
	}

	return u.action.Run(relName, chart, values)
}

func (u *upgrader) loadChart(chartName string) (*chart.Chart, error) {
	cp, err := u.action.LocateChart(chartName, u.envSettings)
	if err != nil {
		return nil, err
	}

	return loader.Load(cp)
}
