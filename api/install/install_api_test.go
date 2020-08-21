package install

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/gojekfarm/albatross/pkg/logger"

	"helm.sh/helm/v3/pkg/release"
)

type mockService struct {
	mock.Mock
}

func (s *mockService) Install(ctx context.Context, req Request) (Response, error) {
	args := s.Called(ctx, req)
	return args.Get(0).(Response), args.Error(1)
}

type InstallerTestSuite struct {
	suite.Suite
	recorder    *httptest.ResponseRecorder
	server      *httptest.Server
	mockService *mockService
}

func (s *InstallerTestSuite) SetupSuite() {
	logger.Setup("default")
}

func (s *InstallerTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.mockService = new(mockService)
	handler := Handler(s.mockService)
	s.server = httptest.NewServer(handler)
}

func (s *InstallerTestSuite) TestShouldReturnDeployedStatusOnSuccessfulInstall() {
	chartName := "stable/redis-ha"
	body := fmt.Sprintf(`{"chart":"%s", "name": "redis-v5", "values": {"replicas": 2}, "flags": {"namespace": "albatross"}}`, chartName)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/install", s.server.URL), strings.NewReader(body))
	response := Response{
		Status: release.StatusDeployed.String(),
	}

	s.mockService.On("Install", mock.Anything, mock.AnythingOfType("Request")).Return(response, nil)

	resp, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	expectedResponse := `{"status":"deployed"}` + "\n"
	respBody, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(s.T(), expectedResponse, string(respBody))
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *InstallerTestSuite) TestShouldReturnInternalServerErrorOnFailure() {
	chartName := "stable/redis-ha"
	body := fmt.Sprintf(`{"chart":"%s", "name": "redis-v5"}`, chartName)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/install", s.server.URL), strings.NewReader(body))
	s.mockService.On("Install", mock.Anything, mock.AnythingOfType("Request")).Return(Response{}, errors.New("Invalid chart"))

	resp, err := http.DefaultClient.Do(req)

	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
	expectedResponse := `{"error":"Invalid chart"}` + "\n"
	respBody, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(s.T(), expectedResponse, string(respBody))
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *InstallerTestSuite) TestReturnShouldBadRequestOnInvalidRequest() {
	chartName := "stable/redis-ha"
	body := fmt.Sprintf(`{"chart":"%s", "name": "redis-v5}`, chartName)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/install", s.server.URL), strings.NewReader(body))
	s.mockService.On("Install", mock.Anything, mock.AnythingOfType("Request")).Return(Release{}, nil)

	resp, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
	require.NoError(s.T(), err)
	s.mockService.AssertNotCalled(s.T(), "Install")
}

func (s *InstallerTestSuite) TearDownTest() {
	s.server.Close()
}

func TestInstallAPI(t *testing.T) {
	suite.Run(t, new(InstallerTestSuite))
}
