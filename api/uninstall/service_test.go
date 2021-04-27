package uninstall

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

const testReleaseName = "test-release-name"

var (
	errNewUninstallerError  = errors.New("new uninstaller error")
	errUninstallActionError = errors.New("uninstall action error")
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

func (m *mockHelmClient) NewUninstaller(fl flags.UninstallFlags) (helmcli.Uninstaller, error) {
	args := m.Called(fl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(helmcli.Uninstaller), args.Error(1)
}

func (m *mockHelmClient) NewStatusGiver(fl flags.StatusFlags) (helmcli.StatusGiver, error) {
	args := m.Called(fl)
	return args.Get(0).(helmcli.StatusGiver), args.Error(1)
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
	releaseOptions := &release.MockReleaseOptions{
		Name:      testReleaseName,
		Version:   1,
		Namespace: "default",
		Chart:     nil,
		Status:    release.StatusDeployed,
	}
	uninstallFlags := flags.UninstallFlags{
		Release: testReleaseName,
		Timeout: defaultTimeout,
	}
	mockRelease := release.Mock(releaseOptions)
	uiResponse := release.UninstallReleaseResponse{Release: mockRelease}
	cli.On("NewUninstaller", uninstallFlags).Times(1).Return(uic, nil)
	uic.On("Uninstall", ctx, testReleaseName).Times(1).Return(&uiResponse, nil)

	resp, err := service.Uninstall(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Release)
	assert.NotEmpty(t, resp.Status)
	assert.Empty(t, resp.Error)
	rel := resp.Release
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

func TestShouldHandleNewUninstallerFailureWithError(t *testing.T) {
	cli := new(mockHelmClient)
	service := NewService(cli)
	ctx := context.Background()
	req := Request{ReleaseName: testReleaseName, Timeout: 2}
	uninstallFlags := flags.UninstallFlags{
		Release:     testReleaseName,
		GlobalFlags: flags.GlobalFlags{},
		Timeout:     time.Second * 2,
	}
	cli.On("NewUninstaller", uninstallFlags).Times(1).Return(nil, errNewUninstallerError)

	resp, err := service.Uninstall(ctx, req)

	assert.True(t, errors.Is(err, errNewUninstallerError))
	assert.Nil(t, resp.Release)
	require.NotNil(t, resp)
	cli.AssertExpectations(t)
}

func TestShouldReturnResponseAndProperErrorWhenReleaseIsNotFound(t *testing.T) {
	cli := new(mockHelmClient)
	uic := new(mockUninstaller)
	service := NewService(cli)
	ctx := context.Background()
	globalFlag := flags.GlobalFlags{KubeContext: "minikube"}
	req := Request{ReleaseName: testReleaseName, GlobalFlags: globalFlag}
	uninstallFlags := flags.UninstallFlags{Release: testReleaseName, GlobalFlags: globalFlag, Timeout: defaultTimeout}
	cli.On("NewUninstaller", uninstallFlags).Times(1).Return(uic, nil)
	uic.On("Uninstall", ctx, testReleaseName).Times(1).Return(nil, driver.ErrReleaseNotFound)

	resp, err := service.Uninstall(ctx, req)

	assert.Error(t, err)
	require.NotNil(t, resp)
	assert.Nil(t, resp.Release)
	assert.True(t, errors.Is(err, driver.ErrReleaseNotFound))
	cli.AssertExpectations(t)
	uic.AssertExpectations(t)
}

func TestShouldReturnResponseAndProperErrorWhenUninstallActionFails(t *testing.T) {
	cli := new(mockHelmClient)
	uic := new(mockUninstaller)
	service := NewService(cli)
	ctx := context.Background()
	req := Request{ReleaseName: testReleaseName, KeepHistory: true, DryRun: true, DisableHooks: true}
	uninstallFlags := flags.UninstallFlags{Release: testReleaseName, KeepHistory: true, DryRun: true, DisableHooks: true, Timeout: defaultTimeout}
	cli.On("NewUninstaller", uninstallFlags).Times(1).Return(uic, nil)
	uic.On("Uninstall", ctx, testReleaseName).Times(1).Return(&release.UninstallReleaseResponse{}, errUninstallActionError)

	resp, err := service.Uninstall(ctx, req)

	assert.Error(t, err)
	require.NotNil(t, resp)
	assert.Nil(t, resp.Release)
	assert.True(t, errors.Is(err, errUninstallActionError))
	cli.AssertExpectations(t)
	uic.AssertExpectations(t)
}
