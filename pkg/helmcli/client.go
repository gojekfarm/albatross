package helmcli

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli/config"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

type Client interface {
	NewUpgrader(flags.UpgradeFlags) (Upgrader, error)
	NewInstaller(flags.InstallFlags) (Installer, error)
	NewLister(flags.ListFlags) (Lister, error)
}

type Upgrader interface {
	Upgrade(ctx context.Context, relName, chartName string, values map[string]interface{}) (*release.Release, error)
}

type Installer interface {
	Install(ctx context.Context, relName, chartName string, values map[string]interface{}) (*release.Release, error)
}

type Lister interface {
	List(ctx context.Context) ([]*release.Release, error)
}

func New() Client {
	return helmClient{}
}

type helmClient struct{}

func (c helmClient) NewUpgrader(flg flags.UpgradeFlags) (Upgrader, error) {
	//TODO: ifpossible envconfig could be moved to actionconfig new, remove pointer usage of globalflags
	envconfig := config.NewEnvConfig(&flg.GlobalFlags)
	actionconfig, err := config.NewActionConfig(envconfig, &flg.GlobalFlags)
	if err != nil {
		return nil, err
	}

	upgrade := action.NewUpgrade(actionconfig.Configuration)
	history := action.NewHistory(actionconfig.Configuration)
	installer, err := c.NewInstaller(flags.InstallFlags{
		DryRun:      flg.DryRun,
		Version:     flg.Version,
		GlobalFlags: flg.GlobalFlags,
	})
	if err != nil {
		return nil, err
	}

	upgrade.Namespace = flg.Namespace
	upgrade.Install = flg.Install
	upgrade.DryRun = flg.DryRun

	return &upgrader{
		action:      upgrade,
		envSettings: envconfig.EnvSettings,
		history:     history,
		installer:   installer,
	}, nil
}

// NewInstaller returns a new instance of Installer struct
func (c helmClient) NewInstaller(flg flags.InstallFlags) (Installer, error) {
	envconfig := config.NewEnvConfig(&flg.GlobalFlags)
	actionconfig, err := config.NewActionConfig(envconfig, &flg.GlobalFlags)
	if err != nil {
		return nil, err
	}

	install := action.NewInstall(actionconfig.Configuration)
	install.Namespace = flg.Namespace
	install.DryRun = flg.DryRun

	return &installer{
		action:      install,
		envSettings: envconfig.EnvSettings,
	}, nil
}

// NewLister returns a new Lister instance
func (c helmClient) NewLister(flg flags.ListFlags) (Lister, error) {
	envconfig := config.NewEnvConfig(&flg.GlobalFlags)
	actionconfig, err := config.NewActionConfig(envconfig, &flg.GlobalFlags)
	if err != nil {
		return nil, err
	}

	list := action.NewList(actionconfig.Configuration)
	list.AllNamespaces = flg.AllNamespaces
	list.Deployed = flg.Deployed
	list.Failed = flg.Failed
	list.Pending = flg.Pending
	list.Uninstalling = flg.Uninstalling
	list.Uninstalled = flg.Uninstalled

	return &lister{
		action:      list,
		envSettings: envconfig.EnvSettings,
	}, nil
}
