package repository

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gojekfarm/albatross/pkg/logger"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gotest.tools/assert"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) Add(ctx context.Context, req AddRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

type RepoAddTestSuite struct {
	suite.Suite
	recorder    *httptest.ResponseRecorder
	server      *httptest.Server
	mockService *mockService
}

func (s *RepoAddTestSuite) SetupSuite() {
	logger.Setup("default")
}

func (s *RepoAddTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.mockService = new(mockService)
	router := mux.NewRouter()
	path := fmt.Sprintf("/repositories/{%s}", NAME)
	router.Handle(path, AddHandler(s.mockService)).Methods(http.MethodPut)
	s.server = httptest.NewServer(router)
}

func (s *RepoAddTestSuite) TestRepoAddSuccessFul() {
	repoName := "gojek-incubator"
	urlName := "https://gojek.github.io/charts/incubator/"
	body := fmt.Sprintf(`{"url":"%s", "username":"admin", "password":"123", 
	"allow_deprecated_repos":true, "force_update": true, "skip_tls_verify": true}`, urlName)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/repositories/%s", s.server.URL, repoName), strings.NewReader(body))
	request := AddRequest{
		Name:                  repoName,
		URL:                   urlName,
		ForceUpdate:           true,
		InsecureSkipTLSverify: true,
	}

	s.mockService.On("Add", mock.Anything, request).Return(nil)

	resp, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	expectedResponse := `{"message":"Repo gojek-incubator added successfully with url: https://gojek.github.io/charts/incubator/"}` + "\n"
	respBody, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(s.T(), expectedResponse, string(respBody))
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *RepoAddTestSuite) TestRepoAddInvalidRequest() {
	repoName := "gojek-incubator"
	body := `{"username":"admin", "password":"123"}`

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/repositories/%s", s.server.URL, repoName), strings.NewReader(body))

	resp, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
	expectedResponse := `{"error":"url cannot be empty"}` + "\n"
	respBody, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(s.T(), expectedResponse, string(respBody))
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func (s *RepoAddTestSuite) TestRepoAddFailure() {
	repoName := "gojek-incubator"
	urlName := "https://gojek.github.io/charts/incubator/"
	body := fmt.Sprintf(`{"url":"%s", "username":"admin", "password":"123"}`, urlName)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/repositories/%s", s.server.URL, repoName), strings.NewReader(body))
	request := AddRequest{
		Name: repoName,
		URL:  urlName,
	}

	s.mockService.On("Add", mock.Anything, request).Return(errors.New("error adding repository"))

	resp, err := http.DefaultClient.Do(req)
	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
	expectedResponse := `{"error":"error adding repository"}` + "\n"
	respBody, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(s.T(), expectedResponse, string(respBody))
	require.NoError(s.T(), err)
	s.mockService.AssertExpectations(s.T())
}

func TestInstallAPI(t *testing.T) {
	suite.Run(t, new(RepoAddTestSuite))
}
