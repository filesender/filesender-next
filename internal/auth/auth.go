// Package auth contains authentication methods (as an interface) for the FileSender application
package auth

import "net/http"

// Auth is an interface containing the authentication method
type Auth interface {
	UserAuth(r *http.Request) (string, error)
}
