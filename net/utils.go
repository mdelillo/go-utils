package net

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

func GetFreeAddr() (string, error) {
	conn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	return conn.Addr().String(), nil
}

func ServerIsAvailable(address string) bool {
	conn, err := net.Dial("tcp", address)
	if err == nil {
		_ = tls.Client(conn, &tls.Config{InsecureSkipVerify: true}).Handshake()
		_ = conn.Close()

		return true
	}

	return false
}

func WaitForServerToBeAvailable(address string, timeout time.Duration) error {
	timeoutChan := time.After(timeout)

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("failed to connect to %s within %s", address, timeout)
		default:
			if ServerIsAvailable(address) {
				return nil
			}
		}
	}
}
