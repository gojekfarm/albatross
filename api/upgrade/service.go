package upgrade

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

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

	ucli, err := s.cli.NewUpgrader(upgradeflags)
	if err != nil {
		return Response{}, fmt.Errorf("error while initializing upgrader: %s", err)
	}

	rel, err := ucli.Upgrade(ctx, req.Name, req.Chart, req.Values)
	if err != nil {
		return responseWithStatus(rel), err
	}
	resp := Response{Status: rel.Info.Status.String(), Release: releaseInfo(rel)}
	if req.Flags.DryRun {
		resp.Data = rel.Manifest
	}
	return resp, nil
}

func releaseInfo(rel *release.Release) Release {
	return Release{
		Name:       rel.Name,
		Namespace:  rel.Namespace,
		Version:    rel.Version,
		Updated:    rel.Info.FirstDeployed.Local().Time,
		Status:     rel.Info.Status,
		Chart:      rel.Chart.ChartFullPath(),
		AppVersion: rel.Chart.AppVersion(),
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
