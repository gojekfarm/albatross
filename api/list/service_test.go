package list

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

type mockLister struct{ mock.Mock }

func (m *mockLister) List(ctx context.Context) ([]*release.Release, error) {
	args := m.Called(ctx)
	if len(args) < 1 {
		log.Fatalf("error while mocking response for list")
	}
	return args.Get(0).([]*release.Release), args.Error(1)
}

func TestShouldReturnValidResponseOnSuccess(t *testing.T) {
	cli := new(mockHelmClient)
	lic := new(mockLister)
	service := NewService(cli)
	ctx := context.Background()
	req := Request{Flags: Flags{Deployed: true}}
	cli.On("NewLister", mock.AnythingOfType("flags.ListFlags")).Return(lic, nil)
	chartloader, err := loader.Loader("../testdata/albatross")
	if err != nil {
		panic("Could not load chart")
	}

	chart, err := chartloader.Load()
	if err != nil {
		panic("Unable to load chart")
	}

	releases := []*release.Release{
		{
			Name:      "test-release",
			Namespace: "test-namespace",
			Version:   1,
			Info: &release.Info{
				FirstDeployed: time.Now(),
				Status:        release.StatusDeployed,
			},
			Chart: chart,
		},
	}

	lic.On("List", ctx).Return(releases, nil)

	resp, err := service.List(ctx, req)

	assert.NoError(t, err)
	require.NotNil(t, resp)
	rel := resp.Releases[0]
	assert.Equal(t, rel.Name, releases[0].Name)
	assert.Equal(t, rel.Namespace, releases[0].Namespace)
	assert.Equal(t, rel.Version, releases[0].Version)
	assert.Equal(t, rel.Status, releases[0].Info.Status)
	assert.Equal(t, rel.Chart, releases[0].Chart.ChartFullPath())
	assert.Equal(t, rel.Updated, releases[0].Info.FirstDeployed.Local().Time)
	assert.Equal(t, rel.AppVersion, releases[0].Chart.AppVersion())
	assert.Empty(t, resp.Error)
	cli.AssertExpectations(t)
	lic.AssertExpectations(t)
}
