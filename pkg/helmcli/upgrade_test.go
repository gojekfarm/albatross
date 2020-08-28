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

func fakeUpgradeConfiguration(t *testing.T) *action.Configuration {
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

func TestUpgradeShouldFailForNonExistentReleaseWithoutInstall(t *testing.T) {
	config := fakeUpgradeConfiguration(t)
	u := &upgrader{
		action:      action.NewUpgrade(config),
		history:     action.NewHistory(config),
		envSettings: cli.New(),
		installer: &installer{
			action:      action.NewInstall(config),
			envSettings: cli.New(),
		},
	}

	values := map[string]interface{}{
		"test": "test",
	}
	_, err := u.Upgrade(context.Background(), "test-release", "../../api/testdata/albatross", values)

	assert.Error(t, err)
	assert.EqualError(t, err, "\"test-release\" has no deployed releases")
}

func TestUpgradeShouldSucceedForNonExistentReleaseWithInstall(t *testing.T) {
	config := fakeUpgradeConfiguration(t)
	u := &upgrader{
		action:      action.NewUpgrade(config),
		history:     action.NewHistory(config),
		envSettings: cli.New(),
		installer: &installer{
			action:      action.NewInstall(config),
			envSettings: cli.New(),
		},
	}

	u.action.Install = true

	values := map[string]interface{}{
		"test": "test",
	}
	release, err := u.Upgrade(context.Background(), "test-release", "../../api/testdata/albatross", values)

	assert.NoError(t, err)
	assert.Equal(t, release.Name, "test-release")
}

func TestUpgradeShouldFailForInvalidChart(t *testing.T) {
	config := fakeUpgradeConfiguration(t)
	u := &upgrader{
		action:      action.NewUpgrade(config),
		history:     action.NewHistory(config),
		envSettings: cli.New(),
		installer: &installer{
			action:      action.NewInstall(config),
			envSettings: cli.New(),
		},
	}

	values := map[string]interface{}{
		"test": "test",
	}

	// TODO: See if we can override registry client to mock this instead of reyling on filesystem
	_, err := u.Upgrade(context.Background(), "test-release", "../../api/testdata/albatrossdne", values)

	assert.Error(t, err)
	assert.EqualError(t, err, "error loading chart: path \"../../api/testdata/albatrossdne\" not found")
}

func TestUpgradeShouldReturnUpgradedReleaseOnSuccess(t *testing.T) {
	config := fakeUpgradeConfiguration(t)

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

	u := &upgrader{
		action:      action.NewUpgrade(config),
		history:     action.NewHistory(config),
		envSettings: cli.New(),
		installer: &installer{
			action:      action.NewInstall(config),
			envSettings: cli.New(),
		},
	}

	values := map[string]interface{}{
		"test": "test",
	}

	release, err := u.Upgrade(context.Background(), "test-release", "../../api/testdata/albatross", values)

	assert.NoError(t, err)
	assert.Equal(t, release.Name, existingRelease.Name)
	assert.Equal(t, release.Version, existingRelease.Version+1)
}
