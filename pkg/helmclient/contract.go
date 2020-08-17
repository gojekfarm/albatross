package helmclient

import (
	"time"

	"helm.sh/helm/v3/pkg/release"
)

// ListOperation defines the contract for a list action
type ListOperation struct {
	AllNamespaces bool `json:"all-namespaces,omitempty"`
	Deployed      bool `json:"deployed,omitempty"`
	Failed        bool `json:"failed,omitempty"`
	Pending       bool `json:"pending,omitempty"`
	Uninstalled   bool `json:"uninstalled,omitempty"`
	Uninstalling  bool `json:"uninstalling,omitempty"`

	*GlobalFlags
}

// NewListOperation return an instance of ListOperation
func NewListOperation() *ListOperation {
	return &ListOperation{
		GlobalFlags: &GlobalFlags{},
	}
}

// Release represents the Release contract for the interactions with helmclient package
type Release struct {
	Name       string         `json:"name"`
	Namespace  string         `json:"namespace"`
	Version    int            `json:"version"`
	Updated    time.Time      `json:"updated_at,omitempty"`
	Status     release.Status `json:"status"`
	Chart      string         `json:"chart"`
	AppVersion string         `json:"app_version"`
}

// NewRelease returns an instance of Release that wraps release.Release
func NewRelease(release *release.Release) *Release {
	return &Release{
		Name:       release.Name,
		Namespace:  release.Namespace,
		Version:    release.Version,
		Updated:    release.Info.FirstDeployed.Local().Time,
		Status:     release.Info.Status,
		Chart:      release.Chart.ChartFullPath(),
		AppVersion: release.Chart.AppVersion(),
	}
}

// InstallOperation defines the contract for the install action
type InstallOperation struct {
	Name   string
	Chart  string
	Values map[string]interface{}
	Flags  *InstallFlags
}

// NewInstallOperation returns an instance of InstallOperation after initializing
// the global flags
func NewInstallOperation() *InstallOperation {
	return &InstallOperation{
		Flags: NewInstallFlags(),
	}
}

// UpgradeOperation defines the contract for the upgrade action
type UpgradeOperation struct {
	Name   string
	Chart  string
	Values map[string]interface{}
	Flags  *UpgradeFlags
}

// NewUpgradeOperation returns an instance of UpgradeOperation after initializing
// the global flags
func NewUpgradeOperation() *UpgradeOperation {
	return &UpgradeOperation{
		Flags: NewUpgradeFlags(),
	}
}
