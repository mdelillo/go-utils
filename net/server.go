package net

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	listenAddr string
	handler    http.Handler
	enableTLS  bool
	certPath   string
	keyPath    string
}

type ServerOption func(s *Server)

func WithTLS(certPath, keyPath string) ServerOption {
	return func(s *Server) {
		s.enableTLS = true
		s.certPath = certPath
		s.keyPath = keyPath
	}
}

func NewServer(listenAddr string, handler http.Handler, options ...ServerOption) Server {
	server := Server{listenAddr: listenAddr, handler: handler}

	for _, option := range options {
		option(&server)
	}

	return server
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{Addr: s.listenAddr, Handler: s.handler}

	var err error
	if s.enableTLS {
		err = s.httpServer.ListenAndServeTLS(s.certPath, s.keyPath)
	} else {
		err = s.httpServer.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Shutdown() error {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			return err
		}

		s.httpServer = nil
	}

	return nil
}
