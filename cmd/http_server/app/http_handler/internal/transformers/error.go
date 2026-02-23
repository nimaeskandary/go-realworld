package transformers

import (
	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
)

func ToApiError(errs ...error) api_gen.GenericErrorJSONResponse {
	// we have to define this struct here again because in the openapi spec this is defined as an anonymous object and not a named component
	type apiErrors struct {
		Body []string `json:"body"`
	}
	errorMessages := make([]string, len(errs))
	for i, err := range errs {
		errorMessages[i] = err.Error()
	}
	return api_gen.GenericErrorJSONResponse{
		Errors: apiErrors{
			Body: errorMessages,
		},
	}
}
