package helmcli

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli/config"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

// Lister acts as an entrypoint for the list action
type Lister struct {
	action      *action.List
	envSettings *cli.EnvSettings
}

// NewLister returns a new Lister instance
func NewLister(flg flags.ListFlags) *Lister {
	envconfig := config.NewEnvConfig(&flg.GlobalFlags)
	actionconfig := config.NewActionConfig(envconfig, &flg.GlobalFlags)

	list := action.NewList(actionconfig.Configuration)
	list.AllNamespaces = flg.AllNamespaces
	list.Deployed = flg.Deployed
	list.Failed = flg.Failed
	list.Pending = flg.Pending
	list.Uninstalling = flg.Uninstalling
	list.Uninstalled = flg.Uninstalled

	return &Lister{
		action:      list,
		envSettings: envconfig.EnvSettings,
	}
}

// List runs the list operation
func (l *Lister) List(ctx context.Context) ([]*release.Release, error) {
	return l.action.Run()
}
