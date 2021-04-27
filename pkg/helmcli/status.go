package helmcli

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

type statusGiver struct {
	action      *action.Status
	envSettings *cli.EnvSettings
}

func (s *statusGiver) Status(ctx context.Context, releaseName string) (*release.Release, error) {
	return s.action.Run(releaseName)
}
