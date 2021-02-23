package uninstall

import (
	"context"
	"log"
	"testing"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

const testReleaseName = "test-release-name"

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

func (m *mockHelmClient) NewUninstaller(fl flags.UninstallFlags) (helmcli.Uninstaller, error) {
	args := m.Called(fl)
	return args.Get(0).(helmcli.Uninstaller), args.Error(1)
}

type mockUninstaller struct{ mock.Mock }

func (m *mockUninstaller) Uninstall(ctx context.Context, releaseName string) (*release.UninstallReleaseResponse, error) {
	args := m.Called(ctx, releaseName)
	if len(args) < 1 {
		log.Fatalf("error while mocking response for uninstall")
	}
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*release.UninstallReleaseResponse), args.Error(1)
}

func TestShouldReturnValidResponseOnSuccess(t *testing.T) {
	cli := new(mockHelmClient)
	uic := new(mockUninstaller)
	service := NewService(cli)
	ctx := context.Background()
	req := Request{ReleaseName: testReleaseName}
	cli.On("NewUninstaller", mock.AnythingOfType("flags.UninstallFlags")).Return(uic, nil)

	releaseOptions := &release.MockReleaseOptions{
		Name:      testReleaseName,
		Version:   1,
		Namespace: "default",
		Chart:     nil,
		Status:    release.StatusDeployed,
	}
	mockRelease := release.Mock(releaseOptions)
	uiResponse := release.UninstallReleaseResponse{Release: mockRelease}
	uic.On("Uninstall", ctx, testReleaseName).Return(&uiResponse, nil)

	resp, err := service.Uninstall(ctx, req)

	assert.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Status)
	rel := resp.Release
	assert.Equal(t, resp.Error, "")
	assert.Equal(t, rel.Name, mockRelease.Name)
	assert.Equal(t, rel.Namespace, mockRelease.Namespace)
	assert.Equal(t, rel.Version, mockRelease.Version)
	assert.Equal(t, rel.Status, mockRelease.Info.Status)
	assert.Equal(t, rel.Chart, mockRelease.Chart.ChartFullPath())
	assert.Equal(t, rel.Updated, mockRelease.Info.FirstDeployed.Local().Time)
	assert.Equal(t, rel.AppVersion, mockRelease.Chart.AppVersion())
	assert.Empty(t, resp.Error)
	cli.AssertExpectations(t)
	uic.AssertExpectations(t)
}

func TestShouldNotCrashOnFailure(t *testing.T) {
	cli := new(mockHelmClient)
	uic := new(mockUninstaller)
	service := NewService(cli)
	ctx := context.Background()
	req := Request{ReleaseName: testReleaseName}

	cli.On("NewUninstaller", mock.AnythingOfType("flags.UninstallFlags")).Return(uic, nil)
	uic.On("Uninstall", ctx, testReleaseName).Return(nil, driver.ErrReleaseNotFound)

	resp, err := service.Uninstall(ctx, req)

	assert.Error(t, err)
	require.NotNil(t, resp)
	cli.AssertExpectations(t)
	uic.AssertExpectations(t)
}
