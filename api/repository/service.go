package repository

import (
	"context"

	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
)

type Service struct {
	cli helmcli.RepositoryClient
}

func (s Service) Add(ctx context.Context, req AddRequest) error {
	addFlags := flags.AddFlags{
		Name:                 req.Name,
		URL:                  req.URL,
		AllowDeprecatedRepos: req.AllowDeprecatedRepos,
		Username:             req.Username,
		Password:             req.Password,
		ForceUpdate:          req.ForceUpdate,
	}

	adder, err := s.cli.NewAdder(addFlags)
	if err != nil {
		return err
	}

	err = adder.Add(ctx)
	if err != nil {
		return err
	}
	return nil
}

func NewService(cli helmcli.RepositoryClient) Service {
	return Service{cli}
}
