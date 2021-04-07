package list

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (m *mockService) List(ctx context.Context, req Request) (Response, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(Response), args.Error(1)
}

type ListTestSuite struct {
	suite.Suite
	recorder      *httptest.ResponseRecorder
	server        *httptest.Server
	mockService   *mockService
	restfulServer *httptest.Server
}

func (s *ListTestSuite) SetupSuite() {
	logger.Setup("default")
}

func (s *ListTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.mockService = new(mockService)
	handler := Handler(s.mockService)
	s.server = httptest.NewServer(handler)
	router := mux.NewRouter()
	router.Handle("/releases/{kube_context}", RestHandler(s.mockService)).Methods(http.MethodGet)
	s.restfulServer = httptest.NewServer(router)
}

func (s *ListTestSuite) TestShouldReturnReleasesWhenSuccessfulAPICall() {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeFromStr, _ := time.Parse(layout, str)
	body := `{"release_status":"deployed"}`
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/list", s.server.URL), strings.NewReader(body))

	response := Response{
		Releases: []Release{
			{
				Name:       "test-release",
				Namespace:  "test",
				Version:    1,
				Updated:    timeFromStr,
				Status:     release.StatusDeployed,
				AppVersion: "0.1",
			},
		},
	}
	s.mockService.On("List", mock.Anything, mock.AnythingOfType("Request")).Return(response, nil)

	res, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), 200, res.StatusCode)
	require.NoError(s.T(), err)

	var actualResponse Response
	err = json.NewDecoder(res.Body).Decode(&actualResponse)

	expectedResponse := Response{
		Error:    "",
		Releases: response.Releases,
	}

	assert.Equal(s.T(), expectedResponse.Releases[0], actualResponse.Releases[0])
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *ListTestSuite) TestShouldReturnReleasesWhenSuccessfulAPICallRestful() {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeFromStr, _ := time.Parse(layout, str)
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/releases/staging?all_namespaces=false&deployed=true&namespace=test", s.restfulServer.URL), nil)
	expectedRequestStruct := Request{
		Flags: Flags{
			AllNamespaces: false,
			Deployed:      true,
			GlobalFlags: flags.GlobalFlags{
				Namespace:   "test",
				KubeContext: "staging",
			},
		},
	}
	response := Response{
		Releases: []Release{
			{
				Name:       "test-release",
				Namespace:  "test",
				Version:    1,
				Updated:    timeFromStr,
				Status:     release.StatusDeployed,
				AppVersion: "0.1",
			},
		},
	}
	s.mockService.On("List", mock.Anything, expectedRequestStruct).Return(response, nil)

	res, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), 200, res.StatusCode)
	require.NoError(s.T(), err)

	var actualResponse Response
	err = json.NewDecoder(res.Body).Decode(&actualResponse)

	expectedResponse := Response{
		Error:    "",
		Releases: response.Releases,
	}

	assert.Equal(s.T(), expectedResponse.Releases[0], actualResponse.Releases[0])
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *ListTestSuite) TestShouldReturnBadRequestErrorIfItHasInvalidCharacter() {
	body := `{"release_status":"unknown""""}`
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/list", s.server.URL), strings.NewReader(body))

	res, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), 400, res.StatusCode)
	require.NoError(s.T(), err)
}

func (s *ListTestSuite) TestShouldReturnBadRequestErrorIfItHasInvalidCharacterRestful() {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/releases/staging?namespce=test", s.restfulServer.URL), nil)

	res, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), 400, res.StatusCode)
	require.NoError(s.T(), err)
}

func (s *ListTestSuite) TearDownTest() {
	s.server.Close()
}

func TestListAPI(t *testing.T) {
	suite.Run(t, new(ListTestSuite))
}
