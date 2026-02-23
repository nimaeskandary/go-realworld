package http_handler_types

import (
	"net/http"
)

type HttpHandler interface {
	GetHandler() http.Handler
}
