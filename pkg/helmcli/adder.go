package helmcli

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/logger"

	"github.com/gofrs/flock"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

const timeout time.Duration = 30 * time.Second

type adder struct {
	flags.AddFlags
	settings *cli.EnvSettings
}

func (o *adder) Add(ctx context.Context) error {
	err := o.checkPrerequisite()
	if err != nil {
		return err
	}

	// Acquire a file lock for process synchronization
	fileLock := flock.New(strings.Replace(o.RepoFile, filepath.Ext(o.RepoFile), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer check(fileLock.Unlock)
	}
	if err != nil {
		return err
	}

	c := repo.Entry{
		Name:                  o.Name,
		URL:                   o.URL,
		Username:              o.Username,
		Password:              o.Password,
		CertFile:              o.CertFile,
		KeyFile:               o.KeyFile,
		CAFile:                o.CaFile,
		InsecureSkipTLSverify: o.InsecureSkipTLSverify,
	}

	f, err := o.initialiseRepoFile(c)
	if err != nil {
		return err
	}

	err = o.initialiseChartsFromRepository(c)
	if err != nil {
		return err
	}

	f.Update(&c)

	if err := f.WriteFile(o.RepoFile, 0644); err != nil {
		return err
	}

	return nil
}

func (o *adder) checkPrerequisite() error {
	// Ensure the file directory exists as it is required for file locking
	err := os.MkdirAll(filepath.Dir(o.RepoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func (o *adder) initialiseRepoFile(c repo.Entry) (*repo.File, error) {
	b, err := ioutil.ReadFile(o.RepoFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, err
	}

	// If the repo exists do one of two things:
	// 1. If the configuration for the name is the same continue without error
	// 2. When the config is different require --force-update
	if !o.ForceUpdate && f.Has(o.Name) {
		existing := f.Get(o.Name)
		if c != *existing {
			// The input coming in for the name is different from what is already
			// configured. Return an error.
			return nil, fmt.Errorf("repository name (%s) already exists, please specify a different name", o.Name)
		}

		// The add is idempotent so do nothing
		return &f, nil
	}

	return &f, nil
}

func (o *adder) initialiseChartsFromRepository(c repo.Entry) error {
	r, err := repo.NewChartRepository(&c, getter.All(o.settings))
	if err != nil {
		return err
	}

	if o.RepoCache != "" {
		r.CachePath = o.RepoCache
	}
	if _, err := r.DownloadIndexFile(); err != nil {
		return fmt.Errorf("%w looks like %v is not a valid chart repository or cannot be reached", err, o.URL)
	}
	return nil
}

func check(f func() error) {
	if err := f(); err != nil {
		logger.Errorf("Error while %v", err)
	}
}
