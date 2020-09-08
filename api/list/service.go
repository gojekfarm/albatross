package list

import (
	"context"

	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

type Service struct {
	cli helmcli.Client
}

func (s Service) List(ctx context.Context, req Request) (Response, error) {
	listflags := flags.ListFlags{
		GlobalFlags: req.Flags.GlobalFlags,
	}
	lcli := s.cli.NewLister(listflags)
	releases, err := lcli.List(ctx)
	if err != nil {
		return Response{}, err
	}

	respReleases := []Release{}
	for _, release := range releases {
		respReleases = append(respReleases, releaseInfo(release))
	}

	resp := Response{Releases: respReleases}
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

func NewService(cli helmcli.Client) Service {
	return Service{cli}
}
