package helmcli

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

type AdderTestSuite struct {
	suite.Suite
}

const (
	testConfigPath        string = "./testdata/config.yaml"
	testCachePath         string = "./testdata/cache"
	testRegistryConfig    string = "./testdata/registry.json"
	testSampleConfigPath1 string = "./testdata/config_sample.yaml"
	testSampleConfigPath2 string = "./testdata/config_sample2.yaml"
)

func (s *AdderTestSuite) SetupTest() {
	os.Setenv("HELM_REPOSITORY_CONFIG", testConfigPath)
	os.Setenv("HELM_REGISTRY_CONFIG", testRegistryConfig)
	os.Setenv("HELM_REPOSITORY_CACHE", testCachePath)
}

func initialiseAdder() *adder {
	settings := cli.New()
	adder := &adder{
		AddFlags: flags.AddFlags{
			Name:      "influxdata",
			URL:       "https://helm.influxdata.com/",
			RepoFile:  settings.RepositoryConfig,
			RepoCache: settings.RepositoryCache,
		},
		settings: settings,
	}
	return adder
}

func (s *AdderTestSuite) TestAddRepo() {
	newAdder := initialiseAdder()

	err := newAdder.Add(context.Background())

	suiteAssertion := s.Assert()
	suiteAssertion.NoError(err)
	b, err := ioutil.ReadFile(newAdder.RepoFile)
	suiteAssertion.NoError(err)
	var f repo.File
	expectedRepo := &repo.Entry{
		Name: "influxdata",
		URL:  "https://helm.influxdata.com/",
	}
	suiteAssertion.NoError(yaml.Unmarshal(b, &f))
	repoFoundCount := 0
	for _, repo := range f.Repositories {
		if repo.Name == expectedRepo.Name {
			suiteAssertion.Equal(expectedRepo, repo)
			repoFoundCount++
		}
	}
	if repoFoundCount == 0 {
		suiteAssertion.FailNow("Did not find added repo in configuration")
	} else if repoFoundCount > 1 {
		suiteAssertion.FailNow("Duplicate repo found")
	}
}

func (s *AdderTestSuite) TestAddDuplicateRepo() {
	err := copyUtil(testSampleConfigPath1, testConfigPath)
	require.NoError(s.T(), err)
	newAdder := initialiseAdder()
	newAdder.Username = "abcd"
	newAdder.Password = "1234"
	err = newAdder.Add(context.Background())
	assert.Error(s.T(), err)
	assert.Equal(s.T(), "repository name (influxdata) already exists, please specify a different name", err.Error())
}

func (s *AdderTestSuite) TestAddDuplicateRepoForceUpdate() {
	err := copyUtil(testSampleConfigPath1, testConfigPath)
	require.NoError(s.T(), err)
	newAdder := initialiseAdder()
	newAdder.Username = "abcd"
	newAdder.Password = "1234"
	newAdder.ForceUpdate = true

	err = newAdder.Add(context.Background())

	suiteAssertion := s.Assert()
	suiteAssertion.NoError(err)
	b, err := ioutil.ReadFile(newAdder.RepoFile)
	suiteAssertion.NoError(err)
	var f repo.File
	expectedRepo := &repo.Entry{
		Name:     "influxdata",
		URL:      "https://helm.influxdata.com/",
		Username: "abcd",
		Password: "1234",
	}
	suiteAssertion.NoError(yaml.Unmarshal(b, &f))
	repoFoundCount := 0
	for _, repo := range f.Repositories {
		if repo.Name == expectedRepo.Name {
			suiteAssertion.Equal(expectedRepo, repo)
			repoFoundCount++
		}
	}
	if repoFoundCount == 0 {
		suiteAssertion.FailNow("Did not find added repo in configuration")
	} else if repoFoundCount > 1 {
		suiteAssertion.FailNow("Duplicate repo found")
	}
}

// func (s *AdderTestSuite) TestAddRepoToExistingFile() {

// }

func (s *AdderTestSuite) TestAddingToExistingConfig() {
	err := copyUtil(testSampleConfigPath2, testConfigPath)
	require.NoError(s.T(), err)
	newAdder := initialiseAdder()
	err = newAdder.Add(context.Background())
	require.NoError(s.T(), err)
	b, err := ioutil.ReadFile(newAdder.RepoFile)
	assert.NoError(s.T(), err)
	var f repo.File
	expectedRepo := &repo.Entry{
		Name: "influxdata",
		URL:  "https://helm.influxdata.com/",
	}
	assert.NoError(s.T(), yaml.Unmarshal(b, &f))
	repoFoundCount := 0
	for _, repo := range f.Repositories {
		if repo.Name == expectedRepo.Name {
			assert.Equal(s.T(), expectedRepo, repo)
			repoFoundCount++
		}
	}
	assert.NotEqual(s.T(), 1, len(f.Repositories))
	if repoFoundCount == 0 {
		assert.FailNow(s.T(), "Did not find added repo in configuration")
	} else if repoFoundCount > 1 {
		assert.FailNow(s.T(), "Duplicate repo found")
	}
}

func (s *AdderTestSuite) TearDownTest() {
	os.Unsetenv("HELM_REPOSITORY_CONFIG")
	os.Unsetenv("HELM_REPOSITORY_CACHE")
	os.Unsetenv("HELM_REGISTRY_CONFIG")
	err := os.Remove(testConfigPath)
	if err != nil {
		s.FailNow("Failed to delete config file", err)
	}
	err = os.RemoveAll(testCachePath)
	if err != nil {
		s.FailNow("Failed to delete cache folder", err)
	}
}

func TestAdderSuite(t *testing.T) {
	suite.Run(t, new(AdderTestSuite))
}

func copyUtil(sourceFileName, destinationFileName string) error {
	srcFile, err := os.Open(sourceFileName)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	if err != nil {
		return err
	}

	destFile, err := os.Create(destinationFileName)
	if err != nil {
		return err
	}
	defer destFile.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	return err
}
