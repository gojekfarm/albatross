package uninstall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gotest.tools/assert"
	"helm.sh/helm/v3/pkg/release"
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
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/uninstall", s.server.URL), strings.NewReader(body))
	requestSturct := Request{
		ReleaseName: testReleaseName,
	}
	releaseOptions := &release.MockReleaseOptions{
		Name:      testReleaseName,
		Version:   1,
		Namespace: "default",
		Chart:     nil,
		Status:    release.StatusDeployed,
	}
	mockRelease := releaseInfo(release.Mock(releaseOptions))
	response := Response{
		Release: mockRelease,
	}
	s.mockService.On("Uninstall", mock.Anything, requestSturct).Times(1).Return(response, nil)

	res, err := http.DefaultClient.Do(req)

	assert.Equal(s.T(), 200, res.StatusCode)
	require.NoError(s.T(), err)

	var actualResponse Response
	err = json.NewDecoder(res.Body).Decode(&actualResponse)
	assert.NilError(s.T(), err)
	assert.Equal(s.T(), mockRelease.Name, actualResponse.Release.Name)
	assert.Equal(s.T(), mockRelease.Version, actualResponse.Release.Version)
	assert.Equal(s.T(), mockRelease.Namespace, actualResponse.Release.Namespace)
	assert.Equal(s.T(), mockRelease.Status, actualResponse.Release.Status)
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *UninstallTestSuite) TestShouldReturnNotFoundErrorIfItHasUnavailableReleaseName() {
	unavailableReleaseName := "unknown_release"
	body := fmt.Sprintf(`{"release_name":"%v"}`, unavailableReleaseName)
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/uninstall", s.server.URL), strings.NewReader(body))
	requestStruct := Request{
		ReleaseName: unavailableReleaseName,
	}
	s.mockService.On("Uninstall", mock.Anything, requestStruct).Times(1).Return(Response{}, driver.ErrReleaseNotFound)

	res, err := http.DefaultClient.Do(req)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), 404, res.StatusCode)
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *UninstallTestSuite) TestShouldReturnInternalServerErrorIfUninstallThrowsUnknownError() {
	body := fmt.Sprintf(`{"release_name":"%v"}`, testReleaseName)
	errMsg := "Test error Message"
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/uninstall", s.server.URL), strings.NewReader(body))
	requestSturct := Request{
		ReleaseName: testReleaseName,
	}
	releaseOptions := &release.MockReleaseOptions{
		Name:      testReleaseName,
		Version:   1,
		Namespace: "default",
		Chart:     nil,
		Status:    release.StatusDeployed,
	}
	mockRelease := releaseInfo(release.Mock(releaseOptions))
	response := Response{
		Release: mockRelease,
	}
	s.mockService.On("Uninstall", mock.Anything, requestSturct).Times(1).Return(response, errors.New(errMsg))

	res, err := http.DefaultClient.Do(req)

	assert.Equal(s.T(), 500, res.StatusCode)
	require.NoError(s.T(), err)

	var actualResponse Response
	err = json.NewDecoder(res.Body).Decode(&actualResponse)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), errMsg, actualResponse.Error)
	s.mockService.AssertExpectations(s.T())
}

func (s *UninstallTestSuite) TestShouldReturnBadRequestErrorIfItHasInvalidReleaseName() {
	body := `{"release_name":""}`
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/uninstall", s.server.URL), strings.NewReader(body))

	res, err := http.DefaultClient.Do(req)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), res)
	var actualResponse Response
	err = json.NewDecoder(res.Body).Decode(&actualResponse)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), actualResponse.Error, errInvalidReleaseName.Error())
	assert.Equal(s.T(), 400, res.StatusCode)
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *UninstallTestSuite) TearDownTest() {
	s.server.Close()
}

func TestUninstallApi(t *testing.T) {
	suite.Run(t, new(UninstallTestSuite))
}
