package middlewares

import (
	"net/http"
)

type Auth interface {
	AuthUser(r *http.Request) (string, error)
}
