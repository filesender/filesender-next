package middlewares

import (
	"net/http"
)

type Auth interface {
	User(r *http.Request) (string, error)
}
