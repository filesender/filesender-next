package auth

import "net/http"

type Auth interface {
	UserAuth(r *http.Request) (string, error)
}
