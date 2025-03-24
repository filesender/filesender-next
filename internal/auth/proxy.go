//go:build !dev

package auth

import (
	"errors"
	"net"
	"net/http"
)

func Auth(r *http.Request) (string, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr)
	if err != nil {
		return "", err
	}
	if !tcpAddr.IP.IsLoopback() {
		return "", errors.New("REMOTE_ADDR is NOT `localhost`")
	}
	remoteUser := r.Header.Get("REMOTE_USER")
	if remoteUser == "" {
		return "", errors.New("HTTP header REMOTE_USER is NOT set")
	}

	return remoteUser, nil
}
