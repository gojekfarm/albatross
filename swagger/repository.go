package swagger

import "github.com/gojekfarm/albatross/api/repository"

// AddErrorResponse body of non 2xx response
// swagger:model addRepoErrorResponseBody
type AddErrorResponse struct {
	Error string `json:"error"`
}

// AddOkResponse body of 2xx response
// swagger:model addRepoOkResponseBody
type AddOkResponse struct {
	Repository repository.Entry `json:"repository"`
}
