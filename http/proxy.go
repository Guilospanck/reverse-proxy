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
		w.WriteStatusCode(NotFoundStatusCode)
		return fmt.Errorf("Path not found")
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	defer conn.Close()
	if err != nil {
		fmt.Println("Error dialing to hostname and port")
		return err
	}
	fmt.Fprintf(conn, req.ToString())

	scanner := bufio.NewScanner(conn)
	var response []string
	for scanner.Scan() {
		response = append(response, string(scanner.Bytes()))
	}
	/*
		HTTP-message:
					start-line CRLF
					*( field-line CRLF )
					CRLF
					[ message-body ]
	*/
	body := strings.Join(response[3:], "")

	fmt.Fprintf(w, body)

	return nil
}
