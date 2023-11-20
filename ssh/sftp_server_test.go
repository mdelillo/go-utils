package ssh_test

import (
	"github.com/mdelillo/go-utils/net"
	"github.com/mdelillo/go-utils/ssh"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSFTPServer(t *testing.T) {
	spec.Run(t, "SFTP Server", testSFTPServer, spec.Report(report.Terminal{}))
}

func testSFTPServer(t *testing.T, context spec.G, it spec.S) {
	var (
		server     *ssh.SFTPServer
		listenAddr string
	)

	it.Before(func() {
		var err error
		listenAddr, err = net.GetFreeAddr()
		require.NoError(t, err)

		server = ssh.NewSFTPServer(listenAddr)
	})

	it.After(func() {
		_ = server.Close()
	})

	context("NewSFTPServer", func() {
		it.Pend("creates an SFTP server", func() {

		})
	})
}
