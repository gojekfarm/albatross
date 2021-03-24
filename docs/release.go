package docs

import (
	"github.com/gojekfarm/albatross/api/list"
	"github.com/gojekfarm/albatross/api/uninstall"
)

// UninstallResponse stub for swagger route for uninstall
// swagger:response uninstallResponse
type UninstallResponse struct {
	//in: body
	Body uninstall.Response
}

// UninstallRequest stub for swagger route for uninstall
// swagger:parameters uninstallRelease
type UninstallRequest struct {
	//in: body
	Body uninstall.Request
}

// ListRequest stub for swagger route for list
// swagger:parameters listRelease
type ListRequest struct {
	//in: body
	Body list.Request
}

// ListResponse stub for swagger route for List
// swagger:response listResponse
type ListResponse struct {
	//in: body
	Body list.Response
}