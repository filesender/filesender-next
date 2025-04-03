package auth

import (
	"errors"
	"net"
	"net/http"
)

type ProxyAuth struct{}

func (s *ProxyAuth) UserAuth(r *http.Request) (string, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr)
	if err != nil {
		return "", err
	}
	if !tcpAddr.IP.IsLoopback() {
		return "", errors.New("REMOTE_ADDR is NOT `localhost`")
	}
	remoteUser := r.Header.Get("X-Remote-User")
	if remoteUser == "" {
		return "", errors.New("HTTP header X-Remote-User is NOT set")
	}

	return remoteUser, nil
}
