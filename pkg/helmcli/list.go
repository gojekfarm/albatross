package helmcli

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

type lister struct {
	action      *action.List
	envSettings *cli.EnvSettings
}

// List runs the list operation.
func (l *lister) List(ctx context.Context) ([]*release.Release, error) {
	return l.action.Run()
}
