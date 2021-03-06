package testkit

import (
	"fmt"
	"net"
)

func GetIdleLocalPort(host string, from, to int) int {
	if from < 1 || to < from {
		return 0
	}

	for port := from; port <= to; port++ {
		listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
		if err == nil {
			listen.Close()
			return port
		}
	}

	return 0

}
