package uninstall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gotest.tools/assert"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/gojekfarm/albatross/pkg/logger"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) Uninstall(ctx context.Context, req Request) (Response, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(Response), args.Error(1)
}

type UninstallTestSuite struct {
	suite.Suite
	recorder    *httptest.ResponseRecorder
	server      *httptest.Server
	mockService *mockService
}

func (s *UninstallTestSuite) SetupSuite() {
	logger.Setup("default")
}

func (s *UninstallTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.mockService = new(mockService)
	handler := Handler(s.mockService)
	s.server = httptest.NewServer(handler)
}

func (s *UninstallTestSuite) TestShouldReturnReleasesWhenSuccessfulAPICall() {
	body := fmt.Sprintf(`{"release_name":"%v"}`, testReleaseName)
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/uninstall", s.server.URL), strings.NewReader(body))

	response := Response{
		Release: releaseInfo(getMockRelease()),
	}
	s.mockService.On("Uninstall", mock.Anything, mock.AnythingOfType("Request")).Return(response, nil)

	res, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), 200, res.StatusCode)
	require.NoError(s.T(), err)

	var actualResponse Response
	err = json.NewDecoder(res.Body).Decode(&actualResponse)
	assert.NilError(s.T(), err)
	expectedResponse := Response{
		Error:   "",
		Release: releaseInfo(getMockRelease()),
	}

	assert.Equal(s.T(), expectedResponse.Release.Name, actualResponse.Release.Name)
	assert.Equal(s.T(), expectedResponse.Release.Version, actualResponse.Release.Version)
	assert.Equal(s.T(), expectedResponse.Release.Namespace, actualResponse.Release.Namespace)
	assert.Equal(s.T(), expectedResponse.Release.Status, actualResponse.Release.Status)
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *UninstallTestSuite) TestShouldReturnBadRequestErrorIfItHasUnavailableReleaseName() {
	body := `{"release_name":"unknown_release"}`
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/uninstall", s.server.URL), strings.NewReader(body))
	s.mockService.On("Uninstall", mock.Anything, mock.AnythingOfType("Request")).Return(Response{}, driver.ErrReleaseNotFound)
	res, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 404, res.StatusCode)
	require.NoError(s.T(), err)
}

func (s *UninstallTestSuite) TestShouldReturnBadRequestErrorIfItHasInvalidReleaseName() {
	body := `{"release_name":""}`
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/uninstall", s.server.URL), strings.NewReader(body))
	s.mockService.On("Uninstall", mock.Anything, mock.AnythingOfType("Request")).Return(Response{}, errInvalidReleaseName)
	res, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 400, res.StatusCode)
	require.NoError(s.T(), err)
}

func (s *UninstallTestSuite) TearDownTest() {
	s.server.Close()
}

func TestUninstallApi(t *testing.T) {
	suite.Run(t, new(UninstallTestSuite))
}
