package uninstall

import (
	"context"
	"fmt"
	"time"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"

	"helm.sh/helm/v3/pkg/release"
)

const defaultTimeout = 300 * time.Second

type Service struct {
	cli helmcli.Client
}

// Uninstall a release according to the request provided and fails if req is incorrect.
func (s Service) Uninstall(ctx context.Context, req Request) (Response, error) {
	var timeout time.Duration
	if req.Timeout < 1 {
		timeout = defaultTimeout
	} else {
		timeout = time.Second * time.Duration(req.Timeout)
	}
	unInstallFlags := flags.UninstallFlags{
		Release:      req.releaseName,
		KeepHistory:  req.KeepHistory,
		DryRun:       req.DryRun,
		DisableHooks: req.DisableHooks,
		Timeout:      timeout,
		GlobalFlags:  req.GlobalFlags,
	}
	u, err := s.cli.NewUninstaller(unInstallFlags)
	if err != nil {
		return Response{}, fmt.Errorf("error while initializing uninstaller: %w", err)
	}
	resp, err := u.Uninstall(ctx, req.releaseName)
	if err != nil {
		if resp != nil {
			return responseWithStatus(resp.Release), err
		}
		return Response{}, err
	}
	return responseWithStatus(resp.Release), nil
}

func responseWithStatus(rel *release.Release) Response {
	resp := Response{}
	if rel != nil && rel.Info != nil {
		resp.Release = releaseInfo(rel)
		resp.Status = rel.Info.Status.String()
	}
	return resp
}

func releaseInfo(rel *release.Release) *Release {
	return &Release{
		Name:       rel.Name,
		Namespace:  rel.Namespace,
		Version:    rel.Version,
		Updated:    rel.Info.FirstDeployed.Local().Time,
		Status:     rel.Info.Status,
		Chart:      rel.Chart.ChartFullPath(),
		AppVersion: rel.Chart.AppVersion(),
	}
}

// NewService returns an uninstall service.
func NewService(cli helmcli.Client) Service {
	return Service{cli}
}
