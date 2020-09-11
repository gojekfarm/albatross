package helmcli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/time"
)

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
	if err := config.Releases.Create(existingRelease); err != nil {
		t.Error(err)
	}

	l := &lister{
		action:      action.NewList(config),
		envSettings: cli.New(),
	}

	releases, err := l.List(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, releases[0].Name, "test-release")
}
