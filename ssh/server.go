package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"net"
	"strings"
)

// TODO: add logger

type Server struct {
	listenAddr      string
	sshServerConfig *ssh.ServerConfig
	debugWriter     io.Writer
	shutdown        chan interface{}
}

type serverConfig struct {
	username       string
	password       string
	authorizedKeys [][]byte
	hostKeys       [][]byte
	debugWriter    io.Writer
}

type ServerOption func(s *serverConfig)

func WithBasicAuth(username, password string) ServerOption {
	// TODO: test me
	return func(s *serverConfig) {
		s.username = username
		s.password = password
	}
}

func WithAuthorizedKeyBytes(authorizedKey []byte) ServerOption {
	// TODO: test me
	return func(config *serverConfig) {
		config.authorizedKeys = append(config.authorizedKeys, authorizedKey)
	}
}

func WithHostKeyBytes(hostKey []byte) ServerOption {
	// TODO: test me
	return func(config *serverConfig) {
		config.hostKeys = append(config.hostKeys, hostKey)
	}
}

func WithDebugWriter(writer io.Writer) ServerOption {
	// TODO: test me
	return func(config *serverConfig) {
		config.debugWriter = writer
	}
}

func NewServer(listenAddr string, options ...ServerOption) (*Server, error) {
	config := serverConfig{
		debugWriter: io.Discard,
	}

	for _, option := range options {
		option(&config)
	}

	sshServerConfig := &ssh.ServerConfig{}

	if config.username != "" && config.password != "" {
		sshServerConfig.PasswordCallback = func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == config.username && string(pass) == config.password {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		}
	}

	if len(config.authorizedKeys) > 0 {
		var authorizedKeysMap map[string]bool
		for _, authorizedKeysBytes := range config.authorizedKeys {
			for len(authorizedKeysBytes) > 0 {
				publicKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
				if err != nil {
					return nil, fmt.Errorf("failed to parse authorized key: %w", err)
				}

				authorizedKeysMap[string(publicKey.Marshal())] = true
				authorizedKeysBytes = rest
			}
		}

		sshServerConfig.PublicKeyCallback = func(c ssh.ConnMetadata, publicKey ssh.PublicKey) (*ssh.Permissions, error) {
			if authorizedKeysMap[string(publicKey.Marshal())] {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"pubkey-fingerprint": ssh.FingerprintSHA256(publicKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		}
	}

	for _, privateKey := range config.hostKeys {
		hostKey, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse host key: %w", err)
		}

		sshServerConfig.AddHostKey(hostKey)
	}

	return &Server{
		listenAddr:      listenAddr,
		sshServerConfig: sshServerConfig,
		debugWriter:     config.debugWriter,
	}, nil
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen for incoming connections: %w", err)
	}
	defer listener.Close()
	s.debug("Listening on %v", listener.Addr())

	for {
		nConn, err := listener.Accept()
		if err != nil {
			s.debug("failed to accept incoming connection: %s", err.Error())
			continue
		}

		go s.handleConnection(nConn)
	}
}

func (s *Server) handleConnection(nConn net.Conn) {
	conn, chans, reqs, err := ssh.NewServerConn(nConn, s.sshServerConfig)
	if err != nil {
		s.debug("failed to create server conn: %s", err.Error())
	}
	s.debug("handling connection: %v", conn)
	//s.debug("logged in with key %s", conn.Permissions.Extensions["pubkey-fingerprint"])

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			err = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			if err != nil {
				s.debug("failed to reject channel: %s", err.Error())
			}
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			s.debug("failed to accept channel: %s", err.Error())
			return
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				_ = req.Reply(req.Type == "shell", nil)
			}
		}(requests)

		term := terminal.NewTerminal(channel, "> ")

		go func() {
			defer channel.Close()
			for {
				line, err := term.ReadLine()
				if err != nil {
					break
				}
				s.debug(line)
			}
		}()
	}
}

func (s *Server) Close() error {
	return nil
}

func (s *Server) debug(format string, a ...any) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	_, _ = fmt.Fprintf(s.debugWriter, format, a...)
}

//func (s *Server) handleChannel(channel ssh.NewChannel) error {
//	if channel.ChannelType() != "session" {
//		err := channel.Reject(ssh.UnknownChannelType, "unknown channel type")
//		if err != nil {
//			return fmt.Errorf("failed to reject channel: %w", err)
//		}
//		return nil
//	}
//
//	acceptedChannel, requests, err := channel.Accept()
//	if err != nil {
//		return fmt.Errorf("failed to accept channel: %w", err)
//	}
//
//	// Sessions have out-of-band requests such as "shell",
//	// "pty-req" and "env".  Here we handle only the
//	// "subsystem" request.
//	go func(in <-chan *ssh.Request) {
//		for req := range in {
//			ok := false
//			switch req.Type {
//			case "subsystem":
//				if string(req.Payload[4:]) == "sftp" {
//					ok = true
//				}
//			}
//			err := req.Reply(ok, nil)
//			if err != nil {
//				return
//			}
//		}
//	}(requests)
//
//	serverOptions := []sftp.ServerOption{
//		sftp.WithDebug(debugStream),
//	}
//
//	if readOnly {
//		serverOptions = append(serverOptions, sftp.ReadOnly())
//		fmt.Fprintf(debugStream, "Read-only server\n")
//	} else {
//		fmt.Fprintf(debugStream, "Read write server\n")
//	}
//
//	server, err := sftp.NewServer(
//		acceptedChannel,
//		serverOptions...,
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if err := server.Serve(); err == io.EOF {
//		server.Close()
//		log.Print("sftp client exited session.")
//	} else if err != nil {
//		log.Fatal("sftp server completed with error:", err)
//	}
//}
