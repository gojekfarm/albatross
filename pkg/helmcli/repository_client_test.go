package helmcli

import (
	"os"
	"testing"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"

	"github.com/stretchr/testify/suite"
)

type RepositoryClientTestSuite struct {
	suite.Suite
	c RepositoryClient
}

const (
	configPath string = "./testdata/repositories.yaml"
	cachePath  string = "./testdata/repositories"
)

func (s *RepositoryClientTestSuite) SetupTest() {
	s.c = NewRepoClient()
	os.Setenv("HELM_REPOSITORY_CONFIG", configPath)
	os.Setenv("HELM_REPOSITORY_CACHE", cachePath)
}

func (s *RepositoryClientTestSuite) TestNewAdder() {
	addFlags := flags.AddFlags{
		Name:     "repo-name",
		URL:      "http://helm-repository.com",
		Username: "abcd",
		Password: "1234",
	}
	newAdder, err := s.c.NewAdder(addFlags)

	assertion := s.Assert()
	assertion.NoError(err)
	adderStruct, ok := newAdder.(*adder)
	assertion.True(ok)
	assertion.Equal(addFlags.Name, adderStruct.Name)
	assertion.Equal(addFlags.URL, adderStruct.URL)
	assertion.Equal(addFlags.Username, adderStruct.Username)
	assertion.Equal(addFlags.Password, adderStruct.Password)
	assertion.Equal(cachePath, adderStruct.RepoCache)
	assertion.Equal(configPath, adderStruct.RepoFile)
	assertion.NotNil(adderStruct.settings)
}

func (s *RepositoryClientTestSuite) TearDownTest() {
	os.Unsetenv("HELM_REPOSITORY_CONFIG")
	os.Unsetenv("HELM_REPOSITORY_CACHE")
}

func TestRepoClient(t *testing.T) {
	suite.Run(t, new(RepositoryClientTestSuite))
}
