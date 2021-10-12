package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/helmcli/repository"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepositoryClient struct{ mock.Mock }

func (m *mockRepositoryClient) NewAdder(addFlags flags.AddFlags) (repository.Adder, error) {
	args := m.Called(addFlags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repository.Adder), args.Error(1)
}

type mockAdder struct{ mock.Mock }

func (m *mockAdder) Add(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestServiceAddSuccessful(t *testing.T) {
	mockCli := new(mockRepositoryClient)
	adder := new(mockAdder)
	s := NewService(mockCli)
	req := AddRequest{
		Name: "repoName",
		URL:  "https://gojek.github.io/charts/incubator/",
	}
	addFlags := flags.AddFlags{
		Name: "repoName",
		URL:  "https://gojek.github.io/charts/incubator/",
	}
	mockCli.On("NewAdder", addFlags).Return(adder, nil).Once()
	adder.On("Add", mock.Anything).Return(nil).Once()

	err := s.Add(context.Background(), req)

	require.NoError(t, err)
}

func TestServiceNewAdderError(t *testing.T) {
	mockCli := new(mockRepositoryClient)
	s := NewService(mockCli)
	req := AddRequest{
		Name: "repoName",
		URL:  "https://gojek.github.io/charts/incubator/",
	}
	addFlags := flags.AddFlags{
		Name: "repoName",
		URL:  "https://gojek.github.io/charts/incubator/",
		// Username: "user",
		// Password: "password",
	}
	adderError := errors.New("failed creating adder")
	mockCli.On("NewAdder", addFlags).Return(nil, adderError)

	err := s.Add(context.Background(), req)

	require.Error(t, adderError, err)
}

func TestServiceAddError(t *testing.T) {
	mockCli := new(mockRepositoryClient)
	adder := new(mockAdder)
	s := NewService(mockCli)
	req := AddRequest{
		Name: "repoName",
		URL:  "https://gojek.github.io/charts/incubator/",
	}
	addFlags := flags.AddFlags{
		Name: "repoName",
		URL:  "https://gojek.github.io/charts/incubator/",
		// Username: "user",
		// Password: "password",
	}
	addError := errors.New("error while adding repo")
	mockCli.On("NewAdder", addFlags).Return(adder, nil).Once()
	adder.On("Add", mock.Anything).Return(addError).Once()

	err := s.Add(context.Background(), req)

	require.Error(t, addError, err)
}
