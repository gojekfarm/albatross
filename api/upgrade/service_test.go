package upgrade

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

func (m *mockHelmClient) NewUpgrader(fl flags.UpgradeFlags) (helmcli.Upgrader, error) {
	args := m.Called(fl)
	return args.Get(0).(helmcli.Upgrader), args.Error(1)
}

func (m *mockHelmClient) NewInstaller(fl flags.InstallFlags) (helmcli.Installer, error) {
	args := m.Called(fl)
	return args.Get(0).(helmcli.Installer), args.Error(1)
}

func (m *mockHelmClient) NewLister(fl flags.ListFlags) (helmcli.Lister, error) {
	args := m.Called(fl)
	return args.Get(0).(helmcli.Lister), args.Error(1)
}

type mockUpgrader struct{ mock.Mock }

func (m *mockUpgrader) Upgrade(ctx context.Context, relName, chart string, values map[string]interface{}) (*release.Release, error) {
	args := m.Called(ctx, relName, chart, values)
	if len(args) < 2 {
		log.Fatalf("error while mocking response for upgrade")
	}
	return args.Get(0).(*release.Release), args.Error(1)
}

func TestShouldReturnErrorOnInvalidChart(t *testing.T) {
	helmcli := new(mockHelmClient)
	upgc := new(mockUpgrader)
	service := NewService(helmcli)
	ctx := context.Background()
	req := Request{Name: "invalid_release", Chart: "stable/invalid_chart"}
	helmcli.On("NewUpgrader", mock.AnythingOfType("flags.UpgradeFlags")).Return(upgc, nil)
	release := &release.Release{Info: &release.Info{Status: release.StatusFailed}}
	upgc.On("Upgrade", ctx, req.Name, req.Chart, req.Values).Return(release, errors.New("failed to download invalid-chart"))

	resp, err := service.Upgrade(ctx, req)

	assert.EqualError(t, err, "failed to download invalid-chart")
	require.NotNil(t, resp)
	assert.Empty(t, resp.Error)
	assert.Equal(t, "failed", resp.Status)
	helmcli.AssertExpectations(t)
	upgc.AssertExpectations(t)
}

func TestShouldReturnValidResponseOnSuccess(t *testing.T) {
	helmcli := new(mockHelmClient)
	upgc := new(mockUpgrader)
	service := NewService(helmcli)
	ctx := context.Background()
	req := Request{Name: "invalid_release", Chart: "stable/invalid_chart"}
	helmcli.On("NewUpgrader", mock.AnythingOfType("flags.UpgradeFlags")).Return(upgc, nil)
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

	upgc.On("Upgrade", ctx, req.Name, req.Chart, req.Values).Return(release, nil)

	resp, err := service.Upgrade(ctx, req)

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
	upgc.AssertExpectations(t)
}
