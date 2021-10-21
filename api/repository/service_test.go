package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/helmcli/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/repo"
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

func (m *mockAdder) Add(ctx context.Context) (*repo.Entry, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repo.Entry), args.Error(1)
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
	mockAdd := &repo.Entry{
		Name: "repoName",
		URL:  "https://gojek.github.io/charts/incubator/",
	}
	mockCli.On("NewAdder", addFlags).Return(adder, nil).Once()
	adder.On("Add", mock.Anything).Return(mockAdd, nil).Once()

	resp, err := s.Add(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.Name, resp.Name)
	assert.Equal(t, req.URL, resp.URL)
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

	resp, err := s.Add(context.Background(), req)

	require.Error(t, adderError, err)
	assert.Equal(t, Entry{}, resp)
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
	adder.On("Add", mock.Anything).Return(nil, addError).Once()

	resp, err := s.Add(context.Background(), req)

	require.Error(t, addError, err)
	assert.Equal(t, Entry{}, resp)
}
