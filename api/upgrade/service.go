package upgrade

import (
	"context"

	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

type upgrader interface {
	Upgrade(ctx context.Context, release, chart string, values map[string]interface{}) (*release.Release, error)
}

type Service struct {
	cli helmcli.Client
}

func (s Service) Upgrade(ctx context.Context, req Request) (Response, error) {
	upgradeflags := flags.UpgradeFlags{
		DryRun:      req.Flags.DryRun,
		Version:     req.Flags.Version,
		Install:     req.Flags.Install,
		GlobalFlags: req.Flags.GlobalFlags,
	}

	ucli := s.cli.NewUpgrader(upgradeflags)
	release, err := ucli.Upgrade(ctx, req.Name, req.Chart, req.Values)
	if err != nil {
		return responseWithStatus(release), err
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

func responseWithStatus(rel *release.Release) Response {
	resp := Response{}
	if rel != nil && rel.Info != nil {
		resp.Status = rel.Info.Status.String()
	}
	return resp
}

func NewService(cli helmcli.Client) Service {
	return Service{cli}
}
