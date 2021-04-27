package status

import (
	"context"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

type Service struct {
	cli helmcli.Client
}

func (s Service) Status(ctx context.Context, req Request) (*Release, error) {
	flg := flags.StatusFlags{
		Version:     req.Version,
		GlobalFlags: req.GlobalFlags,
	}

	statusGiver, err := s.cli.NewStatusGiver(flg)
	if err != nil {
		return nil, err
	}

	rel, err := statusGiver.Status(ctx, req.name)
	if err != nil {
		return nil, err
	}

	return &Release{
		Name:       rel.Name,
		Namespace:  rel.Namespace,
		Version:    rel.Version,
		Updated:    rel.Info.FirstDeployed.Local().Time,
		Status:     rel.Info.Status,
		Chart:      rel.Chart.ChartFullPath(),
		AppVersion: rel.Chart.AppVersion(),
	}, err
}

func NewService(cli helmcli.Client) Service {
	return Service{cli}
}
