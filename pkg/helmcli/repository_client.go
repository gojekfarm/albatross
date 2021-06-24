package helmcli

import (
	"context"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"

	"helm.sh/helm/v3/pkg/cli"
)

type RepositoryClient interface {
	NewAdder(flags.AddFlags) (Adder, error)
}

type Adder interface {
	Add(ctx context.Context) error
}

type repoClient struct{}

func (c repoClient) NewAdder(addFlags flags.AddFlags) (Adder, error) {
	settings := cli.New()
	addFlags.RepoCache = settings.RepositoryCache
	addFlags.RepoFile = settings.RepositoryConfig
	newAdder := adder{AddFlags: addFlags, settings: settings}
	return &newAdder, nil
}

func NewRepoClient() RepositoryClient {
	return repoClient{}
}
