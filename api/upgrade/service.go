package upgrade

import (
	"context"

	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

type service interface {
	Upgrade(ctx context.Context, req Request) (Response, error)
}

type Service struct{}

func (s Service) Upgrade(ctx context.Context, req Request) (Response, error) {
	upgradeflags := flags.UpgradeFlags{
		DryRun:      req.Flags.DryRun,
		Version:     req.Flags.Version,
		Install:     req.Flags.Install,
		GlobalFlags: req.Flags.GlobalFlags,
	}

	ucli := helmcli.NewUpgrader(upgradeflags)
	release, err := ucli.Upgrade(ctx, req.Name, req.Chart, req.Values)
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
