package swagger

// AddErrorResponse body of non 2xx response
// swagger:model addRepoErrorResponseBody
type AddErrorResponse struct {
	Error string `json:"error"`
}
