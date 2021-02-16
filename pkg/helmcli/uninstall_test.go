package helmcli

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
)

const testReleaseName = "test-release-albatross"

func TestUninstallShouldFailForInvalidRelease(t *testing.T){
	actionConfig := fakeUninstallConfiguration(t)
	u := &uninstaller{
		action: action.NewUninstall(actionConfig),
		envSettings: cli.New(),
	}
	_, err := u.Uninstall(context.Background(), testReleaseName+"-incorrect")
	assert.Error(t, err)
}

func TestUninstallShouldSucceedForValidRelease(t *testing.T){
	actionConfig := fakeUninstallConfiguration(t)
	u := &uninstaller{
		action: action.NewUninstall(actionConfig),
		envSettings: cli.New(),
	}
	response, err := u.Uninstall(context.Background(), testReleaseName)
	assert.Nil(t, err)
	assert.NotNil(t, response)
}

func fakeUninstallConfiguration(t *testing.T) *action.Configuration{
	newStorage := storage.Init(driver.NewMemory())
	newStorage.Create(getMockRelease())
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

func getMockRelease() *release.Release{
	releaseOptions := &release.MockReleaseOptions{
		Name: testReleaseName,
		Version: 1,
		Namespace: "default",
		Chart: nil,
		Status: release.StatusDeployed,
	}

	return release.Mock(releaseOptions)
}