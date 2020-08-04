package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"albatross/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PingTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	server   *httptest.Server
}

func (s *PingTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	handler := api.Ping()
	s.server = httptest.NewServer(handler)
}

func (s *PingTestSuite) TestShouldReturnPongWhenPingCall() {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/ping", s.server.URL), strings.NewReader(""))
	res, httpErr := http.DefaultClient.Do(req)
	var pingResponse api.PingResponse
	err := json.NewDecoder(res.Body).Decode(&pingResponse)

	assert.Equal(s.T(), api.PingResponse{"", "pong"}, pingResponse)
	assert.Equal(s.T(), 200, res.StatusCode)
	require.NoError(s.T(), httpErr)
	require.NoError(s.T(), err)
}

func (s *PingTestSuite) TearDownTest() {
	s.server.Close()
}

func TestPingAPI(t *testing.T) {
	suite.Run(t, new(PingTestSuite))
}
