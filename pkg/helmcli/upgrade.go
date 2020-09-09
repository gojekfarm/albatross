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
		if _, runErr := u.history.Run(relName); runErr == driver.ErrReleaseNotFound {
			rel, err := u.installer.Install(ctx, relName, chartName, values)
			if err != nil {
				return nil, err
			}

			return rel, nil
		} else if runErr != nil {
			return nil, runErr
		}
	}

	ch, err := u.loadChart(chartName)
	if err != nil {
		return nil, fmt.Errorf("error loading chart: %w", err)
	}

	return u.action.Run(relName, ch, values)
}

func (u *upgrader) loadChart(chartName string) (*chart.Chart, error) {
	cp, err := u.action.LocateChart(chartName, u.envSettings)
	if err != nil {
		return nil, err
	}

	return loader.Load(cp)
}
