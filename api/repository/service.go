package repository

import (
	"context"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/helmcli/repository"
	"helm.sh/helm/v3/pkg/repo"
)

type Service struct {
	cli repository.Client
}

func (s Service) Add(ctx context.Context, req AddRequest) (AddResponse, error) {
	addFlags := flags.AddFlags{
		Name:        req.Name,
		URL:         req.URL,
		ForceUpdate: req.ForceUpdate,
	}

	adder, err := s.cli.NewAdder(addFlags)
	if err != nil {
		return AddResponse{}, err
	}

	entry, err := adder.Add(ctx)
	if err != nil {
		return AddResponse{}, err
	}
	return AddResponse{Repository: getEntry(entry)}, nil
}

func NewService(cli repository.Client) Service {
	return Service{cli}
}

func getEntry(entry *repo.Entry) *Entry {
	if entry != nil {
		return &Entry{
			Name:     entry.Name,
			URL:      entry.URL,
			Username: entry.Username,
			Password: entry.Password,
		}
	}
	return nil
}
