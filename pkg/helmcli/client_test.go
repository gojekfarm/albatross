package helmcli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

type TestSuite struct {
	suite.Suite
	c       Client
	version string
}

func (s *TestSuite) SetupTest() {
	s.c = New()
	s.version = "0.1.0"
}

func (s *TestSuite) TestNewUpgraderSetsChartOptionsUsingFlagValues() {
	t := s.T()
	install := false
	flg := flags.UpgradeFlags{
		Version: s.version,
		Install: install,
	}
	u, _ := s.c.NewUpgrader(flg)
	newUpgrader, _ := u.(*upgrader)
	assert.Equal(t, s.version, newUpgrader.action.Version)
	assert.Equal(t, install, newUpgrader.action.Install)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
