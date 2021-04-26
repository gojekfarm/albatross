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
	router := mux.NewRouter()
	router.Handle("/clusters/{cluster}/namespaces/{namespace}/releases/{release_name}", Handler(s.mockService)).Methods(http.MethodPut)
	s.server = httptest.NewServer(router)
}

func (s *InstallerTestSuite) TestShouldReturnDeployedStatusOnSuccessfulInstall() {
	chartName := "stable/redis-ha"
	body := fmt.Sprintf(`{"chart":"%s", "values": {"replicas": 2}, "flags": {}}`, chartName)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/clusters/minikube/namespaces/albatross/releases/redis-v5", s.server.URL), strings.NewReader(body))
	response := Response{
		Status: release.StatusDeployed.String(),
	}
	requestStruct := Request{
		Chart: chartName,
		Name:  "redis-v5",
		Values: map[string]interface{}{
			"replicas": float64(2),
		},
		Flags: Flags{
			GlobalFlags: flags.GlobalFlags{
				Namespace:   "albatross",
				KubeContext: "minikube",
			},
		},
	}
	s.mockService.On("Install", mock.Anything, requestStruct).Return(response, nil)

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
	body := fmt.Sprintf(`{"chart":"%s"}`, chartName)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/clusters/minikube/namespaces/albatross/releases/redis-v5", s.server.URL), strings.NewReader(body))
	requestStruct := Request{
		Chart: chartName,
		Name:  "redis-v5",
		Flags: Flags{
			GlobalFlags: flags.GlobalFlags{
				Namespace:   "albatross",
				KubeContext: "minikube",
			},
		},
	}
	s.mockService.On("Install", mock.Anything, requestStruct).Return(Response{}, errors.New("invalid chart"))

	resp, err := http.DefaultClient.Do(req)

	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
	expectedResponse := `{"error":"invalid chart"}` + "\n"
	respBody, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(s.T(), expectedResponse, string(respBody))
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *InstallerTestSuite) TestReturnShouldBadRequestOnInvalidRequest() {
	chartName := "stable/redis-ha"
	body := fmt.Sprintf(`{"chart":"%s}`, chartName)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/clusters/minikube/namespaces/albatross/releases/redis-v5", s.server.URL), strings.NewReader(body))

	resp, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
	require.NoError(s.T(), err)
}

func (s *InstallerTestSuite) TearDownTest() {
	s.server.Close()
}

func TestInstallAPI(t *testing.T) {
	suite.Run(t, new(InstallerTestSuite))
}
