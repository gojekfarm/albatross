package api_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gotest.tools/assert"

	"github.com/gojekfarm/albatross/api"
	"github.com/gojekfarm/albatross/pkg/logger"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

type UpgradeTestSuite struct {
	suite.Suite
	recorder        *httptest.ResponseRecorder
	server          *httptest.Server
	mockUpgrader    *mockUpgrader
	mockHistory     *mockHistory
	mockChartLoader *mockChartLoader
	appConfig       *cli.EnvSettings
}

func (s *UpgradeTestSuite) SetupSuite() {
	logger.Setup("default")
}

func (s *UpgradeTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.mockUpgrader = new(mockUpgrader)
	s.mockHistory = new(mockHistory)
	s.mockChartLoader = new(mockChartLoader)
	s.appConfig = &cli.EnvSettings{
		RepositoryConfig: "./testdata/helm",
		PluginsDirectory: "./testdata/helm/plugin",
	}
	service := api.NewService(s.appConfig, s.mockChartLoader, nil, nil, s.mockUpgrader, s.mockHistory)
	handler := api.Upgrade(service)
	s.server = httptest.NewServer(handler)
}

func (s *UpgradeTestSuite) TestShouldReturnDeployedStatusOnSuccessfulUpgrade() {
	chartName := "stable/redis-ha"
	body := fmt.Sprintf(`{
		"chart":"%s",
		"name": "redis-v5",
		"flags": {
			"install": false
		},
		"values": {
			"usePassword": false
		},
		"namespace": "something"}`, chartName)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/upgrade", s.server.URL), strings.NewReader(body))
	s.mockChartLoader.On("LocateChart", chartName, s.appConfig).Return("./testdata/albatross", nil)
	ucfg := api.ReleaseConfig{ChartName: chartName, Name: "redis-v5", Namespace: "something", Version: ">0.0.0-0"}
	s.mockUpgrader.On("GetInstall").Return(false)
	s.mockUpgrader.On("SetConfig", ucfg)
	release := &release.Release{Info: &release.Info{Status: release.StatusDeployed}}
	vals := map[string]interface{}{"usePassword": false}
	//TODO: pass chart object and verify values present testdata chart yml
	s.mockUpgrader.On("Run", "redis-v5", mock.AnythingOfType("*chart.Chart"), vals).Return(release, nil)

	resp, err := http.DefaultClient.Do(req)

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	expectedResponse := `{"status":"deployed"}` + "\n"
	respBody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(s.T(), expectedResponse, string(respBody))
	require.NoError(s.T(), err)
	s.mockUpgrader.AssertExpectations(s.T())
	s.mockChartLoader.AssertExpectations(s.T())
}

func (s *UpgradeTestSuite) TestShouldReturnInternalServerErrorOnFailure() {
	chartName := "stable/redis-ha"
	body := fmt.Sprintf(`{
    "chart":"%s",
	"name": "redis-v5",
	"flags": {
	    "install": true,
        "version": "7.5.4"
	},
    "namespace": "something"}`, chartName)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/install", s.server.URL), strings.NewReader(body))
	ucfg := api.ReleaseConfig{ChartName: chartName, Name: "redis-v5", Namespace: "something", Version: "7.5.4", Install: true}
	s.mockUpgrader.On("SetConfig", ucfg)
	s.mockChartLoader.On("LocateChart", chartName, s.appConfig).Return("./testdata/albatross", errors.New("Invalid chart"))

	resp, err := http.DefaultClient.Do(req)

	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
	expectedResponse := `{"error":"error in locating chart: Invalid chart"}` + "\n"
	respBody, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(s.T(), expectedResponse, string(respBody))
	require.NoError(s.T(), err)
	s.mockUpgrader.AssertExpectations(s.T())
	s.mockChartLoader.AssertExpectations(s.T())
}

func (s *UpgradeTestSuite) TearDownTest() {
	s.server.Close()
}

func TestUpgradeAPI(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}
