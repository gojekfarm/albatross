package upgrade

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/logger"

	"helm.sh/helm/v3/pkg/release"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) Upgrade(ctx context.Context, req Request) (Response, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(Response), args.Error(1)
}

type UpgradeTestSuite struct {
	suite.Suite
	recorder    *httptest.ResponseRecorder
	server      *httptest.Server
	mockService *mockService
}

func (s *UpgradeTestSuite) SetupSuite() {
	logger.Setup("default")
}

func (s *UpgradeTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.mockService = new(mockService)
	router := mux.NewRouter()
	router.Handle("/clusters/{cluster}/namespaces/{namespace}/releases/{release_name}", Handler(s.mockService)).Methods(http.MethodPut)
	s.server = httptest.NewServer(router)
}

func (s *UpgradeTestSuite) TestShouldReturnDeployedStatusOnSuccessfulUpgrade() {
	chartName := "stable/redis-ha"
	body := fmt.Sprintf(`{
		"chart":"%s",
		"flags": {
			"install": true
		},
		"values": {
			"usePassword": false
		}}`, chartName)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/clusters/staging/namespaces/something/releases/redis-v5", s.server.URL), strings.NewReader(body))
	requestStruct := Request{
		name:  "redis-v5",
		Chart: chartName,
		Flags: Flags{
			Install: true,
			GlobalFlags: flags.GlobalFlags{
				Namespace:   "something",
				KubeContext: "staging",
			},
		},
		Values: map[string]interface{}{
			"usePassword": false,
		},
	}
	response := Response{
		Status: release.StatusDeployed.String(),
	}
	s.mockService.On("Upgrade", mock.Anything, requestStruct).Return(response, nil)

	resp, err := http.DefaultClient.Do(req)

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *UpgradeTestSuite) TestShouldReturnInternalServerErrorOnFailure() {
	chartName := "stable/redis-ha"
	body := fmt.Sprintf(`{
    "chart":"%s",
	"flags": {
	    "install": true, "namespace": "something2", "version": "7.5.4"
	}}`, chartName)
	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/clusters/staging-context/namespaces/something/releases/redis-v5", s.server.URL), strings.NewReader(body))
	requestStruct := Request{
		name:  "redis-v5",
		Chart: chartName,
		Flags: Flags{
			Install: true,
			Version: "7.5.4",
			GlobalFlags: flags.GlobalFlags{
				Namespace:   "something",
				KubeContext: "staging-context",
			},
		},
	}
	s.mockService.On("Upgrade", mock.Anything, requestStruct).Return(Response{}, errors.New("invalid chart"))

	resp, err := http.DefaultClient.Do(req)

	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
	require.NoError(s.T(), err)
}

func (s *UpgradeTestSuite) TestShouldBadRequestOnInvalidRequest() {
	chartName := "stable/redis-ha"
	body := fmt.Sprintf(`{
    "chart":"%s",
	"name": "redis-v5",
	"flags": {
	    "install": true, "namespace": true, "version": 7.5.4
	}}`, chartName)
	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/clusters/staging-context/namespaces/something/releases/redis-v5", s.server.URL), strings.NewReader(body))

	resp, err := http.DefaultClient.Do(req)

	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
	require.NoError(s.T(), err)
}

func (s *UpgradeTestSuite) TearDownTest() {
	s.server.Close()
}

func TestUpgradeAPI(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}
