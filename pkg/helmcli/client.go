package helmcli

import (
	"context"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli/config"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

type Client interface {
	NewUpgrader(flags.UpgradeFlags) (Upgrader, error)
	NewInstaller(flags.InstallFlags) (Installer, error)
	NewLister(flags.ListFlags) (Lister, error)
	NewUninstaller(flags.UninstallFlags) (Uninstaller, error)
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

type Uninstaller interface {
	Uninstall(ctx context.Context, releaseName string) (*release.UninstallReleaseResponse, error)
}

func New() Client {
	return helmClient{}
}

type helmClient struct{}

const hooksTimeout time.Duration = time.Minute * 5

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
	upgrade.Version = flg.Version

	return &upgrader{
		action:      upgrade,
		envSettings: envconfig.EnvSettings,
		history:     history,
		installer:   installer,
	}, nil
}

// NewInstaller returns a new instance of Installer struct.
func (c helmClient) NewInstaller(flg flags.InstallFlags) (Installer, error) {
	envconfig := config.NewEnvConfig(&flg.GlobalFlags)
	actionconfig, err := config.NewActionConfig(envconfig, &flg.GlobalFlags)
	if err != nil {
		return nil, err
	}

	install := action.NewInstall(actionconfig.Configuration)
	install.Namespace = flg.Namespace
	install.DryRun = flg.DryRun
	install.Version = flg.Version

	return &installer{
		action:      install,
		envSettings: envconfig.EnvSettings,
	}, nil
}

// NewLister returns a new Lister instance.
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

func (c helmClient) NewUninstaller(flg flags.UninstallFlags) (Uninstaller, error) {
	envconfig := config.NewEnvConfig(&flg.GlobalFlags)
	actionconfig, err := config.NewActionConfig(envconfig, &flg.GlobalFlags)
	if err != nil {
		return nil, err
	}

	uninstall := action.NewUninstall(actionconfig.Configuration)
	uninstall.KeepHistory = flg.KeepHistory
	uninstall.DisableHooks = flg.DisableHooks
	uninstall.DryRun = flg.DryRun
	uninstall.Timeout = hooksTimeout

	return &uninstaller{
		action:      uninstall,
		envSettings: envconfig.EnvSettings,
	}, nil
}
