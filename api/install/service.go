package install

import (
	"context"

	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

// TODO: Move the service interface to a common place for all apis
type service interface {
	Install(ctx context.Context, req Request) (Response, error)
}

type Service struct{}

func (s Service) Install(ctx context.Context, req Request) (Response, error) {
	installflags := flags.InstallFlags{
		DryRun:      req.Flags.DryRun,
		Version:     req.Flags.Version,
		GlobalFlags: req.Flags.GlobalFlags,
	}
	icli := helmcli.NewInstaller(installflags)
	release, err := icli.Install(ctx, req.Name, req.Chart, req.Values)
	if err != nil {
		return Response{}, err
	}
	resp := Response{Status: release.Info.Status.String(), Release: releaseInfo(release)}
	if req.Flags.DryRun {
		resp.Data = release.Manifest
	}
	return resp, nil
}

func releaseInfo(release *release.Release) Release {
	return Release{
		Name:       release.Name,
		Namespace:  release.Namespace,
		Version:    release.Version,
		Updated:    release.Info.FirstDeployed.Local().Time,
		Status:     release.Info.Status,
		Chart:      release.Chart.ChartFullPath(),
		AppVersion: release.Chart.AppVersion(),
	}
}
