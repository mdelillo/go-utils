package ssh

import (
	"github.com/pkg/sftp"
	"io"
)

type SFTPServer struct {
	listenAddr      string
	readOnly        bool
	enableAllocator bool
	debugWriter     io.Writer
}

type SFTPServerOption func(s *SFTPServer)

func NewSFTPServer(listenAddr string, options ...SFTPServerOption) *SFTPServer {
	server := SFTPServer{listenAddr: listenAddr}

	for _, option := range options {
		option(&server)
	}

	return &server
}

func (s *SFTPServer) Start() error {
	return nil
}

func (s *SFTPServer) Close() error {
	return nil
}

func (s *SFTPServer) sftpServerOptions() []sftp.ServerOption {
	var options []sftp.ServerOption

	if s.readOnly {
		options = append(options, sftp.ReadOnly())
	}
	if s.enableAllocator {
		options = append(options, sftp.WithAllocator())
	}
	if s.debugWriter != nil {
		options = append(options, sftp.WithDebug(s.debugWriter))
	}

	return options
}
