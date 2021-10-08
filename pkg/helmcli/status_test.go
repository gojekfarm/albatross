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

func fakeStatusConfiguration(t *testing.T) *action.Configuration {
	newStorage := storage.Init(driver.NewMemory())
	err := newStorage.Create(
		release.Mock(
			&release.MockReleaseOptions{
				Name:      testReleaseName,
				Version:   1,
				Namespace: "default",
				Chart:     nil,
				Status:    release.StatusDeployed,
			}))
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

func TestStatusShouldFailForInvalidRelease(t *testing.T) {
	actionConfig := fakeStatusConfiguration(t)
	u := &statusGiver{
		action:      action.NewStatus(actionConfig),
		envSettings: cli.New(),
	}

	_, err := u.Status(context.Background(), testReleaseName+"-incorrect")

	assert.Error(t, err)
}

func TestStatusShouldSucceedForValidRelease(t *testing.T) {
	actionConfig := fakeStatusConfiguration(t)
	u := &statusGiver{
		action:      action.NewStatus(actionConfig),
		envSettings: cli.New(),
	}
	response, err := u.Status(context.Background(), testReleaseName)

	assert.Nil(t, err)
	assert.NotNil(t, response)
}
