package helmcli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

type TestSuite struct {
	suite.Suite
	c Client
}

func (s *TestSuite) SetupTest() {
	s.c = New()
}

func (s *TestSuite) TestNewUpgraderSetsChartOptionsUsingFlagValues() {
	t := s.T()
	version := "0.1.0"
	dryRun := false
	install := false
	globalFlags := flags.GlobalFlags{
		Namespace: "namespace",
	}
	flg := flags.UpgradeFlags{
		Version:     version,
		Install:     install,
		DryRun:      dryRun,
		GlobalFlags: globalFlags,
	}
	u, err := s.c.NewUpgrader(flg)
	newUpgrader, ok := u.(*upgrader)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, version, newUpgrader.action.Version)
	assert.Equal(t, install, newUpgrader.action.Install)
	assert.Equal(t, dryRun, newUpgrader.action.DryRun)
	assert.Equal(t, globalFlags.Namespace, newUpgrader.action.Namespace)
}

func (s *TestSuite) TestNewInstallerSetsChartOptionsUsingFlagValues() {
	t := s.T()
	version := "0.1.0"
	dryRun := false
	globalFlags := flags.GlobalFlags{
		Namespace: "namespace",
	}
	flg := flags.InstallFlags{
		Version:     version,
		DryRun:      dryRun,
		GlobalFlags: globalFlags,
	}
	i, err := s.c.NewInstaller(flg)
	newInstaller, ok := i.(*installer)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, version, newInstaller.action.Version)
	assert.Equal(t, dryRun, newInstaller.action.DryRun)
	assert.Equal(t, globalFlags.Namespace, newInstaller.action.Namespace)
}

func (s *TestSuite) TestNewUninstallerUsingFlagValues() {
	t := s.T()
	dryRun := true
	keepHistory := true
	disableHooks := true
	globalFlags := flags.GlobalFlags{
		Namespace: "minikube",
	}
	uiFlags := flags.UninstallFlags{
		GlobalFlags:  globalFlags,
		DryRun:       dryRun,
		KeepHistory:  keepHistory,
		DisableHooks: disableHooks,
	}

	u, err := s.c.NewUninstaller(uiFlags)

	newUninstaller, ok := u.(*uninstaller)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, dryRun, newUninstaller.action.DryRun)
	assert.Equal(t, disableHooks, newUninstaller.action.DisableHooks)
	assert.Equal(t, keepHistory, newUninstaller.action.KeepHistory)
	assert.Equal(t, globalFlags.KubeContext, newUninstaller.envSettings.KubeContext)
}

func (s *TestSuite) TestNewStatusGiverUsingFlagValues() {
	t := s.T()
	globalFlags := flags.GlobalFlags{
		Namespace:   "minikube",
		KubeContext: "staging",
	}
	uiFlags := flags.StatusFlags{
		GlobalFlags: globalFlags,
		Version:     1,
	}

	u, err := s.c.NewStatusGiver(uiFlags)

	newStatusGiver, ok := u.(*statusGiver)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 1, newStatusGiver.action.Version)
	assert.Equal(t, globalFlags.KubeContext, newStatusGiver.envSettings.KubeContext)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
