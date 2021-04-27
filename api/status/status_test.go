package status

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gotest.tools/assert"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/logger"

	"helm.sh/helm/v3/pkg/release"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) Status(ctx context.Context, req Request) (*Release, error) {
	args := m.Called(ctx, req)
	if args.Get(0) != nil {
		return args.Get(0).(*Release), args.Error(1)
	}
	return nil, args.Error(1)
}

type TestSuite struct {
	suite.Suite
	recorder    *httptest.ResponseRecorder
	server      *httptest.Server
	mockService *mockService
}

func (s *TestSuite) SetupSuite() {
	logger.Setup("default")
}

func (s *TestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.mockService = new(mockService)
	router := mux.NewRouter()
	router.Handle("/clusters/{cluster}/namespaces/{namespace}/releases/{release_name}", Handler(s.mockService)).Methods(http.MethodGet)
	s.server = httptest.NewServer(router)
}

func (s *TestSuite) TestShouldReturnReleasesWhenSuccessfulAPICall() {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeFromStr, _ := time.Parse(layout, str)
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/clusters/staging/namespaces/test/releases/mysql-test?revision=1", s.server.URL), nil)
	expectedRequestStruct := Request{
		name:    "mysql-test",
		Version: 1,
		GlobalFlags: flags.GlobalFlags{
			KubeContext: "staging",
			Namespace:   "test",
		},
	}
	response := Release{
		Name:       "test-release",
		Namespace:  "test",
		Version:    1,
		Updated:    timeFromStr,
		Status:     release.StatusDeployed,
		AppVersion: "0.1",
	}

	s.mockService.On("Status", mock.Anything, expectedRequestStruct).Return(&response, nil)

	res, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), 200, res.StatusCode)
	require.NoError(s.T(), err)

	var actualResponse Release
	err = json.NewDecoder(res.Body).Decode(&actualResponse)

	assert.Equal(s.T(), response, actualResponse)
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *TestSuite) TestShouldReturnBadRequestErrorIfItHasInvalidCharacter() {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/clusters/staging/namespaces/test/releases/mysql-test?vision=2", s.server.URL), nil)

	res, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), 400, res.StatusCode)
	require.NoError(s.T(), err)
}

func (s *TestSuite) TestShouldReturnInternalServerErrorIfListServiceReturnsError() {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/clusters/staging/namespaces/test/releases/mysql-test?revision=2", s.server.URL), nil)
	expectedRequestStruct := Request{
		name:    "mysql-test",
		Version: 2,
		GlobalFlags: flags.GlobalFlags{
			KubeContext: "staging",
			Namespace:   "test",
		},
	}
	errorMsg := "test error"
	listError := errors.New(errorMsg)
	s.mockService.On("Status", mock.Anything, expectedRequestStruct).Return(nil, listError).Once()

	res, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), 500, res.StatusCode)
	require.NoError(s.T(), err)
	var actualResponse ErrorResponse
	err = json.NewDecoder(res.Body).Decode(&actualResponse)
	require.NoError(s.T(), err)
	expectedResponse := ErrorResponse{
		Error: errorMsg,
	}
	assert.Equal(s.T(), expectedResponse.Error, actualResponse.Error)
	s.mockService.AssertExpectations(s.T())
}

func (s *TestSuite) TearDownTest() {
	s.server.Close()
}

func TestListAPI(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
