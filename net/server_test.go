package net_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mdelillo/go-utils/certs"
	"github.com/mdelillo/go-utils/net"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetServer(t *testing.T) {
	spec.Run(t, "Server", testServer, spec.Report(report.Terminal{}))
}

func testServer(t *testing.T, context spec.G, it spec.S) {
	var listenAddr string

	it.Before(func() {
		var err error
		listenAddr, err = net.GetFreeAddr()
		require.NoError(t, err)
	})

	context("NewServer", func() {
		it("creates a server", func() {
			handlerResponse := "response from test handler"
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprint(w, handlerResponse)
			})

			server := net.NewServer(listenAddr, handler)

			serverDone := make(chan interface{})
			go func() {
				err := server.Start()
				require.NoError(t, err)
				close(serverDone)
			}()

			err := net.WaitForServerToBeAvailable(listenAddr, 5*time.Second)
			require.NoError(t, err)

			resp, err := http.Get("http://" + listenAddr)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, handlerResponse, string(body))

			err = server.Shutdown()
			require.NoError(t, err)

			assert.Eventually(t, func() bool {
				return readFromChannel(serverDone)
			}, time.Second, 10*time.Millisecond)

			assert.False(t, net.ServerIsAvailable(listenAddr))
		})

		context("WithTLS", func() {
			var tempDir string

			it.Before(func() {
				var err error
				tempDir, err = ioutil.TempDir("", "go-utils-server-test")
				require.NoError(t, err)
			})

			it.After(func() {
				_ = os.RemoveAll(tempDir)
			})

			it("creates a TLS server", func() {
				handlerResponse := "response from test handler"
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = fmt.Fprint(w, handlerResponse)
				})

				certPEMBlock, keyPEMBlock, err := certs.GenerateCert(certs.WithIPAddresses("127.0.0.1"))
				require.NoError(t, err)

				certPath := filepath.Join(tempDir, "cert")
				err = ioutil.WriteFile(certPath, certPEMBlock, 0644)
				require.NoError(t, err)

				keyPath := filepath.Join(tempDir, "key")
				err = ioutil.WriteFile(keyPath, keyPEMBlock, 0644)
				require.NoError(t, err)

				server := net.NewServer(listenAddr, handler, net.WithTLS(certPath, keyPath))

				serverDone := make(chan interface{})
				go func() {
					err := server.Start()
					require.NoError(t, err)
					close(serverDone)
				}()

				err = net.WaitForServerToBeAvailable(listenAddr, 5*time.Second)
				require.NoError(t, err)

				rootCAs := x509.NewCertPool()
				require.True(t, rootCAs.AppendCertsFromPEM(certPEMBlock))
				client := &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							RootCAs: rootCAs,
						},
					},
				}
				resp, err := client.Get("https://" + listenAddr)
				require.NoError(t, err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				assert.Equal(t, handlerResponse, string(body))

				err = server.Shutdown()
				require.NoError(t, err)

				assert.Eventually(t, func() bool {
					return readFromChannel(serverDone)
				}, time.Second, 10*time.Millisecond)

				assert.False(t, net.ServerIsAvailable(listenAddr))
			})
		})
	})
}

func readFromChannel(channel chan interface{}) bool {
	for {
		select {
		case <-channel:
			return true
		default:
			return false
		}
	}
}
