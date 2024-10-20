package http

import (
	"errors"
	"fmt"
	"strings"
)

/*
Some info: https://www.notion.so/HTTP-from-scratch-12180e3900fd801d99e4d105d2ddc7aa?pvs=4

My thoughts are:
- Create a simple http server from scratch, with only GET and POST methods
and only validating a couple of things;
- Start the server using TCP;
- Test it using curl, for example;
- Validate how the well-known libraries implement the http protocol.
*/

/*
HTTP-message   = start-line CRLF

	*( field-line CRLF )
	CRLF
	[ message-body ]

	start-line     = request-line (when server) / status-line (when client)
*/

type Method string

const (
	get  Method = "GET"
	post Method = "POST"
)

type RequestLine struct {
	method        Method
	requestTarget string
	httpVersion   string
}

type HttpServer struct {
	requestLine RequestLine
	headers     map[string]string
	body        *string
}

func NewHTTPServer() HttpServer {
	return HttpServer{}
}

func (server *HttpServer) ParseMessage(byteMessage []byte) {
	var splitMessage []string

	// "A recipient MUSt parse an HTTP message as a sequence of octets (sequence of bytes)"
	// See here: https://httpwg.org/specs/rfc9112.html#message.parsing
	message := string(byteMessage[:])

	// Tries to split the components of HTTP message using CRLF
	splitMessage = strings.Split(message, "\r\n")
	// If not possible, tries using only LF
	if len(splitMessage) == 1 {
		splitMessage = strings.Split(message, "\n")
	}
	// If it is still not possible, it is because there is something wrong
	if len(splitMessage) == 1 {
		panic("Something wrong with the HTTP message")
	}

	startLine := splitMessage[0]
	server.parseRequestLine(startLine)

	fmt.Printf("%v", server.requestLine)
}

func (server *HttpServer) parseHTTPMethod(str string) (Method, error) {
	fmt.Println(Method(str)) // GET
	fmt.Println(strings.Split(str, " "))
	fmt.Println(get) // GET
	fmt.Println(strings.Split(string(get), " "))
	fmt.Println(Method(str) == get) // false
	switch Method(str) {
	case get, post:
		return Method(str), nil
	default:
		return "", errors.New("501 Not Implemented")
	}

}

// https://httpwg.org/specs/rfc9112.html#request.target
func (server *HttpServer) parseRequestTarget(str string) (string, error) {
	// Check for whitespaces
	split := strings.Split(str, " ")
	if len(split) > 1 {
		return "", errors.New("400 Bad Request: request target should not contain whitespace")
	}

	return split[0], nil
}

func (server *HttpServer) parseHTTPVersion(str string) (string, error) {
	split := strings.Split(str, "/")
	if len(split) != 2 {
		return "", errors.New("400 Bad Request: wrong HTTP version")
	}

	protocol := split[0]
	version := split[1]

	if protocol != "HTTP" {
		return "", errors.New("400 Bad Request: not an HTTP protocol")
	}

	if version != "1.1" {
		return "", errors.New("400 Bad Request: HTTP version unsupported")
	}

	return str, nil
}

// request-line   = method SP request-target SP HTTP-version
func (server *HttpServer) parseRequestLine(requestLine string) {
	// split based on whitespace
	split := strings.Split(requestLine, " ")
	if len(split) != 3 {
		panic("Request line not parseable. Expected: method SP request-target SP HTTP-version")
	}

	method, err := server.parseHTTPMethod(split[0])
	if err != nil {
		panic(err)
	}

	requestTarget, err := server.parseRequestTarget(split[1])
	if err != nil {
		panic(err)
	}

	httpVersion, err := server.parseHTTPVersion(split[2])
	if err != nil {
		panic(err)
	}

	server.requestLine.method = method
	server.requestLine.requestTarget = requestTarget
	server.requestLine.httpVersion = httpVersion
}
