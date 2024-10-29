package http

import (
	"errors"
	"fmt"
	"strings"
)

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

type StatusLine struct {
	httpVersion string
	statusCode  int
	statusText  string
}

type Headers map[string]string

type HttpServer struct {
	requestLine RequestLine
	headers     Headers
	body        *string
}

type HttpResponse struct {
	statusLine StatusLine
	headers    Headers
	body       *string
}

func (sl *StatusLine) toString() string {
	return fmt.Sprintf("%s %d %s", sl.httpVersion, sl.statusCode, sl.statusText)
}

func (headers *Headers) toString() string {
	acc := []string{}
	for key, value := range *headers {
		acc = append(acc, fmt.Sprintf("%s: %s", key, value))
	}

	return strings.Join(acc, "\r\n")
}

func (response *HttpResponse) ToString() string {
	if response.body != nil {
		return fmt.Sprintf("%s\r\n%s\r\n\r\n%s\r\n", response.statusLine.toString(), response.headers.toString(), *response.body)
	}

	return fmt.Sprintf("%s\r\n%s\r\n", response.statusLine.toString(), response.headers.toString())
}

func (server *HttpServer) ParseMessage(byteMessage []byte) HttpResponse {
	var splitMessage []string

	// "A recipient MUST parse an HTTP message as a sequence of octets (sequence of bytes)"
	// See here: https://httpwg.org/specs/rfc9112.html#message.parsing
	message := string(byteMessage[:])

	// Remove leading null characters
	message = strings.Trim(message, "\x00")

	// Tries to split the components of HTTP message using CRLF
	splitMessage = strings.Split(message, "\r\n")
	// If not possible, tries using only LF
	if len(splitMessage) == 1 {
		splitMessage = strings.Split(message, "\n")
	}
	// If it is still not possible, it is because there is something wrong
	if len(splitMessage) == 1 {
		return HttpResponse{
			statusLine: StatusLine{
				httpVersion: "HTTP/1.1",
				statusCode:  400,
				statusText:  "Bad Request",
			},
			headers: map[string]string{"Accept": "*/*"},
			body:    nil,
		}
	}

	startLine := splitMessage[0]
	err := server.parseRequestLine(startLine)
	if err != nil {
		return HttpResponse{
			statusLine: StatusLine{
				httpVersion: "HTTP/1.1",
				statusCode:  400,
				statusText:  "Bad Request",
			},
			headers: map[string]string{"Accept": "*/*"},
			body:    nil,
		}
	}

	headersAndBody := splitMessage[1:]
	headersAndBodyLength := len(headersAndBody)

	var messageBodyIndex *int = nil
	for index, line := range headersAndBody {
		isLastValue := index == headersAndBodyLength-1

		if line == "" && !isLastValue && headersAndBody[index+1] != "" {
			messageBodyIndex = &index
			*messageBodyIndex++
			break
		}
	}

	// check headers
	var finalHeadersIndex *int = messageBodyIndex
	if messageBodyIndex == nil {
		finalHeadersIndex = &headersAndBodyLength
	}
	headers := headersAndBody[:*finalHeadersIndex-1]
	err = server.parseHTTPHeaders(headers)
	if err != nil {
		return HttpResponse{
			statusLine: StatusLine{
				httpVersion: "HTTP/1.1",
				statusCode:  400,
				statusText:  "Bad Request",
			},
			headers: map[string]string{"Accept": "*/*"},
			body:    nil,
		}
	}

	responseBody := `
		<!DOCTYPE html>
			<html>
			<body>

			<h1>This is a GET</h1>
			<p>Hell yeah</p>

			</body>
		</html>
	`

	// check body
	if messageBodyIndex == nil {
		return HttpResponse{
			statusLine: StatusLine{
				httpVersion: "HTTP/1.1",
				statusCode:  200,
				statusText:  "Ok",
			},
			headers: map[string]string{"Accept": "*/*"},
			body:    &responseBody,
		}
	}
	messageBody := headersAndBody[*messageBodyIndex]
	err = server.parseHTTPMessageBody(messageBody, messageBodyIndex)
	if err != nil {
		return HttpResponse{
			statusLine: StatusLine{
				httpVersion: "HTTP/1.1",
				statusCode:  400,
				statusText:  "Bad Request",
			},
			headers: map[string]string{"Accept": "*/*"},
			body:    nil,
		}
	}

	responseBody = `
		<!DOCTYPE html>
			<html>
			<body>

			<h1>This is a POST</h1>
			<p>Hell yeah</p>

			</body>
		</html>
	`

	return HttpResponse{
		statusLine: StatusLine{
			httpVersion: "HTTP/1.1",
			statusCode:  200,
			statusText:  "Ok",
		},
		headers: map[string]string{"Accept": "*/*"},
		body:    &responseBody,
	}
}

func (server *HttpServer) parseHTTPMessageBody(messageBody string, messageBodyStartIndex *int) error {
	contentLengthHeader := server.headers["Content-Length"]
	// Have the message body but it's missing the content length header
	// (we are not looking for the transfer encoding atm)
	// See: https://httpwg.org/specs/rfc9112.html#message.body.length
	if contentLengthHeader == "" && messageBodyStartIndex != nil {
		return errors.New("411 Length Required: missing Content-Length header")
	}

	server.body = &messageBody

	return nil
}

func (server *HttpServer) parseHTTPHeaders(array []string) error {
	headers := make(map[string]string)
	for _, item := range array {
		if item == "" {
			continue
		}
		split := strings.Split(item, ":")
		fieldName := split[0]
		fieldValue := strings.Join(split[1:], ":")

		if len(strings.Split(fieldName, " ")) > 1 {
			return errors.New("400 Bad Request: cannot have whitespace between field name and colon")
		}

		headers[fieldName] = strings.Trim(fieldValue, " ")

	}

	server.headers = headers

	return nil
}

func (server *HttpServer) parseHTTPMethod(str string) (Method, error) {
	switch str {
	case string(get):
		return get, nil
	case string(post):
		return post, nil
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
func (server *HttpServer) parseRequestLine(requestLine string) error {
	// split based on whitespace
	split := strings.Split(requestLine, " ")
	if len(split) != 3 {
		return errors.New("Request line not parseable. Expected: method SP request-target SP HTTP-version")
	}

	method, err := server.parseHTTPMethod(split[0])
	if err != nil {
		return err
	}

	requestTarget, err := server.parseRequestTarget(split[1])
	if err != nil {
		return err
	}

	httpVersion, err := server.parseHTTPVersion(split[2])
	if err != nil {
		return err
	}

	server.requestLine.method = method
	server.requestLine.requestTarget = requestTarget
	server.requestLine.httpVersion = httpVersion

	return nil
}

func NewHTTPServer() HttpServer {
	return HttpServer{}
}
