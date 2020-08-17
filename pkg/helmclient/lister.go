package helmclient

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
)

// ListResult is the result of the list operation
type ListResult struct {
	Releases []*Release
}

// NewListResult returns an instance of ListResult.
// It format the result according to the contract
func NewListResult(releases []*release.Release) *ListResult {
	rels := []*Release{}
	for _, release := range releases {
		rels = append(rels, NewRelease(release))
	}

	return &ListResult{
		Releases: rels,
	}
}

// Lister acts as an entrypoint for the list action
// It has listoperation instance member which keeps the state of the operation
type Lister struct {
	operation *ListOperation
}

// NewLister returns a new Lister instance
func NewLister(operation *ListOperation) *Lister {
	return &Lister{
		operation: operation,
	}
}

// newListAction returns a new instance of action.List based on the action config
// It sets the appropriate list action members based on ListOperation
func (lister *Lister) newListAction(actionconfig *ActionConfig) *action.List {
	list := action.NewList(actionconfig.Configuration)
	list.AllNamespaces = lister.operation.AllNamespaces
	list.Deployed = lister.operation.Deployed
	list.Failed = lister.operation.Failed
	list.Pending = lister.operation.Pending
	list.Uninstalling = lister.operation.Uninstalling
	list.Uninstalled = lister.operation.Uninstalled
	return list
}

// Run runs the list operation
func (lister *Lister) Run() (*ListResult, error) {
	envconfig := NewEnvConfig(lister.operation.GlobalFlags)
	actionconfig := NewActionConfig(envconfig, lister.operation.GlobalFlags)
	list := lister.newListAction(actionconfig)

	releases, err := list.Run()
	if err != nil {
		return nil, err
	}

	return NewListResult(releases), nil
}
