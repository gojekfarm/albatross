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
	"helm.sh/helm/v3/pkg/time"
)

func fakeListConfiguration(t *testing.T) *action.Configuration {
	return &action.Configuration{
		Releases: storage.Init(driver.NewMemory()),
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

func TestListShouldReturnListOfReleasesOnSuccess(t *testing.T) {
	config := fakeInstallConfiguration(t)
	// Mark that the release is already created
	existingRelease := &release.Release{
		Name:      "test-release",
		Namespace: "test-namespace",
		Version:   1,
		Info: &release.Info{
			FirstDeployed: time.Now(),
			Status:        release.StatusDeployed,
		},
	}
	config.Releases.Create(existingRelease)

	l := &lister{
		action:      action.NewList(config),
		envSettings: cli.New(),
	}

	releases, err := l.List(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, releases[0].Name, "test-release")
}
