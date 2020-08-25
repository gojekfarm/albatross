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
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
)

func fakeInstallConfiguration(t *testing.T) *action.Configuration {
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

func TestInstallShouldFailForInvalidChart(t *testing.T) {
	config := fakeInstallConfiguration(t)
	u := &installer{
		action:      action.NewInstall(config),
		envSettings: cli.New(),
	}

	values := map[string]interface{}{
		"test": "test",
	}

	// TODO: See if we can override registry client to mock this instead of reyling on filesystem
	_, err := u.Install(context.Background(), "test-release", "../../api/testdata/albatrossdne", values)

	assert.Error(t, err)
	assert.EqualError(t, err, "path \"../../api/testdata/albatrossdne\" not found")
}

func TestInstallShouldReturnInstalledReleaseOnSuccess(t *testing.T) {
	config := fakeInstallConfiguration(t)

	u := &installer{
		action:      action.NewInstall(config),
		envSettings: cli.New(),
	}

	values := map[string]interface{}{
		"test": "test",
	}

	release, err := u.Install(context.Background(), "test-release", "../../api/testdata/albatross", values)

	assert.NoError(t, err)
	assert.Equal(t, release.Name, "test-release")
}
