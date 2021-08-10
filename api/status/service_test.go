package status

import (
	"context"
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
// TODO: Find a way to isolate interface only for upgrade.
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

func (m *mockHelmClient) NewStatusGiver(fl flags.StatusFlags) (helmcli.StatusGiver, error) {
	args := m.Called(fl)
	return args.Get(0).(helmcli.StatusGiver), args.Error(1)
}

func (m *mockHelmClient) NewUninstaller(fl flags.UninstallFlags) (helmcli.Uninstaller, error) {
	args := m.Called(fl)
	return args.Get(0).(helmcli.Uninstaller), args.Error(1)
}

type mockStatusGiver struct{ mock.Mock }

func (m *mockStatusGiver) Status(ctx context.Context, releaseName string) (*release.Release, error) {
	args := m.Called(ctx, releaseName)
	if len(args) < 1 {
		log.Fatalf("error while mocking response for list")
	}
	return args.Get(0).(*release.Release), args.Error(1)
}

func TestShouldReturnValidResponseOnSuccess(t *testing.T) {
	cli := new(mockHelmClient)
	sic := new(mockStatusGiver)
	service := NewService(cli)
	ctx := context.Background()
	req := Request{name: "test-release", Version: 1, GlobalFlags: flags.GlobalFlags{
		KubeContext: "abc", Namespace: "test",
	}}
	statusFlags := flags.StatusFlags{
		Version: 1,
		GlobalFlags: flags.GlobalFlags{
			KubeContext: "abc", Namespace: "test",
		},
	}
	chartloader, err := loader.Loader("../testdata/albatross")

	if err != nil {
		panic("Could not load chart")
	}

	chart, err := chartloader.Load()
	if err != nil {
		panic("Unable to load chart")
	}

	cli.On("NewStatusGiver", statusFlags).Return(sic, nil).Once()

	releases := release.Release{
		Name:      "test-release",
		Namespace: "test",
		Version:   1,
		Info: &release.Info{
			FirstDeployed: time.Now(),
			Status:        release.StatusDeployed,
		},
		Chart: chart,
	}

	sic.On("Status", ctx, "test-release").Return(&releases, nil).Once()

	resp, err := service.Status(ctx, req)

	assert.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, resp.Name, releases.Name)
	assert.Equal(t, resp.Namespace, releases.Namespace)
	assert.Equal(t, resp.Version, releases.Version)
	assert.Equal(t, resp.Status, releases.Info.Status)
	assert.Equal(t, resp.Chart, releases.Chart.ChartFullPath())
	assert.Equal(t, resp.Updated, releases.Info.FirstDeployed.Local().Time)
	assert.Equal(t, resp.AppVersion, releases.Chart.AppVersion())
	cli.AssertExpectations(t)
	sic.AssertExpectations(t)
}
