package helmcli

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

type uninstaller struct {
	action      *action.Uninstall
	envSettings *cli.EnvSettings
}

// List runs the list operation.
func (u *uninstaller) Uninstall(ctx context.Context, releaseName string) (*release.UninstallReleaseResponse, error) {
	return u.action.Run(releaseName)
}
