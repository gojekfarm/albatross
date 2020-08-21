package upgrade

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/release"
)

func TestShouldReturnErrorOnInvalidChart(t *testing.T) {
	helmcli := new(mockHelmClient)
	upgc := new(mockUpgrader)
	service := NewService(helmcli)
	ctx := context.Background()
	req := Request{Name: "invalid_release", Chart: "stable/invalid_chart"}
	helmcli.On("NewUpgrader").Return(upgc)
	release := &release.Release{Info: &release.Info{Status: release.StatusFailed}}
	upgc.On("Upgrade", ctx, req.Name, req.Chart, req.Values).Return(release, errors.New("failed to download invalid-chart"))

	resp, err := service.Upgrade(ctx, req)

	assert.EqualError(t, err, "failed to download invalid-chart")
	require.NotNil(t, resp)
	assert.Empty(t, resp.Error)
	assert.Equal(t, "failed", resp.Status)
	helmcli.AssertExpectations(t)
	upgc.AssertExpectations(t)
}

type mockHelmClient struct{ mock.Mock }

func (m *mockHelmClient) NewUpgrader(fl flags.UpgradeFlags) helmcli.Upgrader {
	return m.Called().Get(0).(upgrader)
}

type mockUpgrader struct{ mock.Mock }

func (m *mockUpgrader) Upgrade(ctx context.Context, relName, chart string, values map[string]interface{}) (*release.Release, error) {
	args := m.Called(ctx, relName, chart, values)
	if len(args) < 2 {
		log.Fatalf("error while mocking response for upgrade")
	}
	return args.Get(0).(*release.Release), args.Error(1)
}
