package install

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/time"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

// To satisfy the client interface, we have to define all methods(NewUpgrade, NewInstaller) on the mock struct
// TODO: Find a way to isolate interface only for upgrade
type mockHelmClient struct{ mock.Mock }

func (m *mockHelmClient) NewUpgrader(fl flags.UpgradeFlags) helmcli.Upgrader {
	return m.Called().Get(0).(helmcli.Upgrader)
}

func (m *mockHelmClient) NewInstaller(fl flags.InstallFlags) helmcli.Installer {
	return m.Called().Get(0).(helmcli.Installer)
}

func (m *mockHelmClient) NewLister(fl flags.ListFlags) helmcli.Lister {
	return m.Called().Get(0).(helmcli.Lister)
}

type mockInstaller struct{ mock.Mock }

func (m *mockInstaller) Install(ctx context.Context, relName, chart string, values map[string]interface{}) (*release.Release, error) {
	args := m.Called(ctx, relName, chart, values)
	if len(args) < 2 {
		log.Fatalf("error while mocking response for install")
	}
	return args.Get(0).(*release.Release), args.Error(1)
}

func TestShouldReturnErrorOnInvalidChart(t *testing.T) {
	helmcli := new(mockHelmClient)
	inc := new(mockInstaller)
	service := NewService(helmcli)
	ctx := context.Background()
	req := Request{Name: "invalid_release", Chart: "stable/invalid_chart"}
	helmcli.On("NewInstaller").Return(inc)
	release := &release.Release{Info: &release.Info{Status: release.StatusFailed}}
	inc.On("Install", ctx, req.Name, req.Chart, req.Values).Return(release, errors.New("failed to download invalid-chart"))

	resp, err := service.Install(ctx, req)

	assert.EqualError(t, err, "failed to download invalid-chart")
	require.NotNil(t, resp)
	assert.Empty(t, resp.Error)
	assert.Equal(t, "failed", resp.Status)
	helmcli.AssertExpectations(t)
	inc.AssertExpectations(t)
}

func TestShouldReturnValidResponseOnSuccess(t *testing.T) {
	helmcli := new(mockHelmClient)
	inc := new(mockInstaller)
	service := NewService(helmcli)
	ctx := context.Background()
	req := Request{Name: "invalid_release", Chart: "stable/invalid_chart"}
	helmcli.On("NewInstaller").Return(inc)
	chartloader, err := loader.Loader("../testdata/albatross")
	if err != nil {
		panic("Could not load chart")
	}

	chart, err := chartloader.Load()
	if err != nil {
		panic("Unable to load chart")
	}

	release := &release.Release{
		Name:      "test-release",
		Namespace: "test-namespace",
		Version:   1,
		Info: &release.Info{
			FirstDeployed: time.Now(),
			Status:        release.StatusDeployed,
		},
		Chart: chart,
	}

	inc.On("Install", ctx, req.Name, req.Chart, req.Values).Return(release, nil)

	resp, err := service.Install(ctx, req)

	assert.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, resp.Name, release.Name)
	assert.Equal(t, resp.Namespace, release.Namespace)
	assert.Equal(t, resp.Version, release.Version)
	assert.Equal(t, resp.Status, release.Info.Status.String())
	assert.Equal(t, resp.Chart, release.Chart.ChartFullPath())
	assert.Equal(t, resp.Updated, release.Info.FirstDeployed.Local().Time)
	assert.Equal(t, resp.AppVersion, release.Chart.AppVersion())
	assert.Empty(t, resp.Error)
	helmcli.AssertExpectations(t)
	inc.AssertExpectations(t)
}
