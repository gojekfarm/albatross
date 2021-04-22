package list

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

func (m *mockService) List(ctx context.Context, req Request) (Response, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(Response), args.Error(1)
}

type ListTestSuite struct {
	suite.Suite
	recorder    *httptest.ResponseRecorder
	server      *httptest.Server
	mockService *mockService
}

func (s *ListTestSuite) SetupSuite() {
	logger.Setup("default")
}

func (s *ListTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.mockService = new(mockService)
	router := mux.NewRouter()
	router.Handle("/clusters/{cluster}/releases", Handler(s.mockService)).Methods(http.MethodGet)
	router.Handle("/clusters/{cluster}/namespaces/{namespace}/releases", Handler(s.mockService)).Methods(http.MethodGet)
	s.server = httptest.NewServer(router)
}

func (s *ListTestSuite) TestShouldReturnReleasesWhenSuccessfulAPICall() {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeFromStr, _ := time.Parse(layout, str)
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/clusters/staging/releases?deployed=true", s.server.URL), nil)
	expectedRequestStruct := Request{
		Flags: Flags{
			AllNamespaces: true,
			Deployed:      true,
			GlobalFlags: flags.GlobalFlags{
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

func (s *ListTestSuite) TestShouldReturnReleasesWhenSuccessfulAPICallNamespace() {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeFromStr, _ := time.Parse(layout, str)
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/clusters/staging/namespaces/test/releases?deployed=true", s.server.URL), nil)
	expectedRequestStruct := Request{
		Flags: Flags{
			Deployed: true,
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

func (s *ListTestSuite) TestShouldReturnNoContentWhenNoReleasesAreAvailable() {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/clusters/staging/namespaces/test/releases?deployed=true", s.server.URL), nil)
	expectedRequestStruct := Request{
		Flags: Flags{
			Deployed: true,
			GlobalFlags: flags.GlobalFlags{
				Namespace:   "test",
				KubeContext: "staging",
			},
		},
	}
	response := Response{
		Releases: []Release{},
	}
	s.mockService.On("List", mock.Anything, expectedRequestStruct).Return(response, nil).Once()

	res, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), 204, res.StatusCode)
	require.NoError(s.T(), err)

	s.mockService.AssertExpectations(s.T())
}
func (s *ListTestSuite) TestShouldReturnBadRequestErrorIfItHasInvalidCharacter() {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/clusters/staging/releases?deply=test", s.server.URL), nil)

	res, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), 400, res.StatusCode)
	require.NoError(s.T(), err)
}

func (s *ListTestSuite) TestShouldReturnInternalServerErrorIfListServiceReturnsError() {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/clusters/staging/namespaces/test/releases?deployed=true", s.server.URL), nil)
	expectedRequestStruct := Request{
		Flags: Flags{
			Deployed: true,
			GlobalFlags: flags.GlobalFlags{
				Namespace:   "test",
				KubeContext: "staging",
			},
		},
	}
	response := Response{}
	errorMsg := "test error"
	listError := errors.New(errorMsg)
	s.mockService.On("List", mock.Anything, expectedRequestStruct).Return(response, listError).Once()

	res, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), 500, res.StatusCode)
	require.NoError(s.T(), err)
	var actualResponse Response
	err = json.NewDecoder(res.Body).Decode(&actualResponse)
	require.NoError(s.T(), err)
	expectedResponse := Response{
		Error: errorMsg,
	}
	assert.Equal(s.T(), expectedResponse.Error, actualResponse.Error)
	s.mockService.AssertExpectations(s.T())
}

func (s *ListTestSuite) TearDownTest() {
	s.server.Close()
}

func TestListAPI(t *testing.T) {
	suite.Run(t, new(ListTestSuite))
}
