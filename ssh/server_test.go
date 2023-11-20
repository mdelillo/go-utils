package ssh_test

import (
	"github.com/mdelillo/go-utils/net"
	"github.com/mdelillo/go-utils/ssh"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	spec.Run(t, "Server", testServer, spec.Report(report.Terminal{}))
}

func testServer(t *testing.T, context spec.G, it spec.S) {
	var (
		server     *ssh.Server
		listenAddr string
	)

	it.Before(func() {
		var err error
		listenAddr, err = net.GetFreeAddr()
		require.NoError(t, err)

		server = ssh.NewServer(listenAddr)
	})

	it.After(func() {
		_ = server.Close()
	})

	context("NewServer", func() {
		it("creates a server", func() {
			serverDone := make(chan interface{})
			go func() {
				err := server.Start()
				require.NoError(t, err)
				close(serverDone)
			}()

			err := net.WaitForServerToBeAvailable(listenAddr, 5*time.Second)
			require.NoError(t, err)

			// test SSH server

			err = server.Close()
			require.NoError(t, err)

			assert.Eventually(t, readFromChannel(serverDone), time.Second, 10*time.Millisecond)

			assert.False(t, net.ServerIsAvailable(listenAddr))
		})
	})
}

func readFromChannel(channel chan interface{}) func() bool {
	return func() bool {
		for {
			select {
			case <-channel:
				return true
			default:
				return false
			}
		}
	}
}
