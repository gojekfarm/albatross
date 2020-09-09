package install

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

func (s Service) Install(ctx context.Context, req Request) (Response, error) {
	installflags := flags.InstallFlags{
		DryRun:      req.Flags.DryRun,
		Version:     req.Flags.Version,
		GlobalFlags: req.Flags.GlobalFlags,
	}
	icli, err := s.cli.NewInstaller(installflags)
	if err != nil {
		return Response{}, fmt.Errorf("error while initializing the installer: %s", err)
	}

	rel, err := icli.Install(ctx, req.Name, req.Chart, req.Values)
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
