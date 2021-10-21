package repository

import (
	"context"
	"fmt"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/helmcli/repository"
	"github.com/gojekfarm/albatross/pkg/logger"

	"helm.sh/helm/v3/pkg/repo"
)

type Service struct {
	cli repository.Client
}

func (s Service) Add(ctx context.Context, req AddRequest) (Entry, error) {
	addFlags := flags.AddFlags{
		Name:        req.Name,
		URL:         req.URL,
		ForceUpdate: req.ForceUpdate,
	}

	adder, err := s.cli.NewAdder(addFlags)
	if err != nil {
		return Entry{}, err
	}

	entry, err := adder.Add(ctx)
	if err != nil {
		return Entry{}, err
	}
	return getEntry(entry)
}

func NewService(cli repository.Client) Service {
	return Service{cli}
}

func getEntry(entry *repo.Entry) (Entry, error) {
	if entry != nil {
		logger.Infof("Repository %s with URL: %s has been added", entry.Name, entry.URL)
		return Entry{
			Name:     entry.Name,
			URL:      entry.URL,
			Username: entry.Username,
			Password: entry.Password,
		}, nil
	}

	return Entry{}, fmt.Errorf("couldn't get repository from user")
}
