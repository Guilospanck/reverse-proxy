package http

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type ReverseProxyServer struct{}

func (proxy *ReverseProxyServer) Serve() {
	proxyServer := NewHTTPServer(proxy.forward)
	go proxyServer.ListenAndServe("6000")
}

func (proxy *ReverseProxyServer) forward(w ResponseWriter, req HttpRequest) error {
	var port string
	switch req.requestLine.requestTarget {
	case "/a":
		port = "3000"
	case "/b":
		port = "4000"
	case "/c":
		port = "5000"
	default:
		return fmt.Errorf("Path not found")
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	defer conn.Close()
	if err != nil {
		fmt.Println("error dialing to hostname and port")
		return err
	}
	fmt.Fprintf(conn, req.ToString())

	// TODO: this is probably not the best way to forward the response
	// to the client
	scanner := bufio.NewScanner(conn)
	var response []string
	for scanner.Scan() {
		response = append(response, string(scanner.Bytes()))
	}
	res := strings.Join(response[3:], "")

	fmt.Fprintf(w, res)

	return nil
}
