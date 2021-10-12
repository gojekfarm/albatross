package repository

import (
	"context"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/helmcli/repository"
)

type Service struct {
	cli repository.Client
}

func (s Service) Add(ctx context.Context, req AddRequest) error {
	addFlags := flags.AddFlags{
		Name:        req.Name,
		URL:         req.URL,
		Username:    req.Username,
		Password:    req.Password,
		ForceUpdate: req.ForceUpdate,
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

func NewService(cli repository.Client) Service {
	return Service{cli}
}
