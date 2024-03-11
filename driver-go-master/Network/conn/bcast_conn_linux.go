//go:build linux
// +build linux

package conn

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

func DialBroadcastUDP(port int) net.PacketConn {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		fmt.Errorf("error creating socket: %w", err)
	}

	err = syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		fmt.Errorf("error setting SO_REUSEADDR: %w", err)
	}

	err = syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil {
		fmt.Errorf("error setting SO_BROADCAST: %w", err)
	}

	err = syscall.Bind(s, &syscall.SockaddrInet4{Port: port})
	if err != nil {
		fmt.Errorf("error binding socket: %w", err)
	}

	f := os.NewFile(uintptr(s), "")
	conn, err := net.FilePacketConn(f)
	f.Close() // Close the file regardless of the outcome
	if err != nil {
		fmt.Errorf("error creating FilePacketConn: %w", err)
	}

	return conn
}
