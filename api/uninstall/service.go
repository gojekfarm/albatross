package uninstall

import (
	"context"
	"fmt"
	"errors"
	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
)

type Service struct {
	cli helmcli.Client
}
// NewService returns an uninstall service
func NewService(cli helmcli.Client) Service{
	return Service{cli}
}
// Uninstall a release according to the request provided and fails if req is incorrect
func (s Service) Uninstall(ctx context.Context, req Request) (Response, error) {

	unInstallFlags := &flags.UninstallFlags{
		Release:      req.ReleaseName,
		KeepHistory:  req.KeepHistory,
		DryRun:       req.Dryrun,
		DisableHooks: req.DisableHooks,
		GlobalFlags:  req.GlobalFlags,
	}
	u, err := s.cli.NewUninstaller(*unInstallFlags)

	if err != nil {
		return Response{}, fmt.Errorf("error while initializing uninstaller: %s", err)
	}

	resp, err := u.Uninstall(ctx, req.ReleaseName)

	if err != nil && resp != nil{
		return responseWithStatus(resp.Release), err
	}

	return Response{
		Status:  string(resp.Release.Info.Status),
		Release: releaseInfo(resp.Release),
	}, nil

}

func responseWithStatus(rel *release.Release) Response {
	resp := Response{}
	if rel != nil && rel.Info != nil {
		resp.Status = rel.Info.Status.String()
	}
	return resp
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

func checkReleaseName(u *action.Uninstall, releaseName string) error{
	if err := validateReleaseName(releaseName); err != nil {
		return err
	}
	
	return nil
}

func validateReleaseName(releaseName string) error {
	if releaseName == "" {
		return errors.New("No release name provided")
	}

	if !action.ValidName.MatchString(releaseName) {
		return errors.New("Invalid release name")
	}

	return nil
}
