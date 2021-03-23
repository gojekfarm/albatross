package docs

import "github.com/gojekfarm/albatross/api/uninstall"

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
