package middlewares

import (
	"errors"
	"fmt"
	"net"
	"net/http"
)

type HeaderAuth struct {
	HeaderName string
}

func (s *HeaderAuth) AuthUser(r *http.Request) (string, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr)
	if err != nil {
		return "", err
	}
	if !tcpAddr.IP.IsLoopback() {
		return "", errors.New("REMOTE_ADDR is NOT `localhost`")
	}
	remoteUser := r.Header.Get(s.HeaderName)
	if remoteUser == "" {
		return "", fmt.Errorf("HTTP header %s NOT set", s.HeaderName)
	}

	return remoteUser, nil
}
