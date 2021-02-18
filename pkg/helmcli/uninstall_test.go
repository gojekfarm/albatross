package helmcli

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
)

const testReleaseName = "test-release-albatross"

func TestUninstallShouldFailForInvalidRelease(t *testing.T) {
	actionConfig := fakeUninstallConfiguration(t)
	u := &uninstaller{
		action:      action.NewUninstall(actionConfig),
		envSettings: cli.New(),
	}
	_, err := u.Uninstall(context.Background(), testReleaseName+"-incorrect")
	assert.Error(t, err)
}

func TestUninstallShouldSucceedForValidRelease(t *testing.T) {
	actionConfig := fakeUninstallConfiguration(t)
	u := &uninstaller{
		action:      action.NewUninstall(actionConfig),
		envSettings: cli.New(),
	}
	response, err := u.Uninstall(context.Background(), testReleaseName)
	assert.Nil(t, err)
	assert.NotNil(t, response)
}

func TestUninstallShouldRemoveTheRelease(t *testing.T) {
	actionConfig := fakeUninstallConfiguration(t)
	u := &uninstaller{
		action:      action.NewUninstall(actionConfig),
		envSettings: cli.New(),
	}
	l := &lister{
		action:      action.NewList(actionConfig),
		envSettings: cli.New(),
	}
	releaseList, err := l.List(context.Background())
	assert.Len(t, releaseList, 1)
	require.NoError(t, err)
	resp, err := u.Uninstall(context.Background(), testReleaseName)
	assert.NotNil(t, resp)
	require.NoError(t, err)
	releaseList, err = l.List(context.Background())
	assert.Len(t, releaseList, 0)
	assert.Nil(t, err)
}

func fakeUninstallConfiguration(t *testing.T) *action.Configuration {
	newStorage := storage.Init(driver.NewMemory())
	err := newStorage.Create(getMockRelease())
	require.NoError(t, err)

	return &action.Configuration{
		Releases: newStorage,
		KubeClient: &kubefake.FailingKubeClient{
			PrintingKubeClient: kubefake.PrintingKubeClient{
				Out: ioutil.Discard,
			},
		},
		Capabilities: chartutil.DefaultCapabilities,
		Log: func(format string, v ...interface{}) {
			t.Helper()
			t.Logf(format, v...)
		},
	}
}

func getMockRelease() *release.Release {
	releaseOptions := &release.MockReleaseOptions{
		Name:      testReleaseName,
		Version:   1,
		Namespace: "default",
		Chart:     nil,
		Status:    release.StatusDeployed,
	}

	return release.Mock(releaseOptions)
}
