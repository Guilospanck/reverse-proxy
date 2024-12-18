package http

import (
	"fmt"
	"io"
	"log"
	"net"
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

type CustomError interface {
	GetStatusCode() StatusCode
	GetBody() *string
}

type HTTPError struct {
	statusCode StatusCode
	body       *string
}

func (httpError *HTTPError) GetStatusCode() StatusCode {
	return httpError.statusCode
}

func (httpError *HTTPError) GetBody() *string {
	return httpError.body
}

type StatusCode int
type StatusText string

var (
	OkStatusCode                  StatusCode = 200
	CreatedStatusCode             StatusCode = 201
	BadRequestStatusCode          StatusCode = 400
	NotFoundStatusCode            StatusCode = 404
	LengthRequiredStatusCode      StatusCode = 411
	InternalServerErrorStatusCode StatusCode = 500
	NotImplementedStatusCode      StatusCode = 501
)

var StatusCodeToStatusText = map[StatusCode]StatusText{
	200: "OK",
	201: "Created",
	400: "Bad Request",
	404: "Not Found",
	411: "Length Required",
	500: "Internal Server Error",
	501: "Not Implemented",
}

const SupportedHTTPVersion string = "HTTP/1.1"

type RequestLine struct {
	method        Method
	requestTarget string
	httpVersion   string
}

type StatusLine struct {
	httpVersion string
	statusCode  *StatusCode
	statusText  StatusText
}

type Headers map[string]string

type HttpRequest struct {
	requestLine RequestLine
	headers     Headers
	body        *string
}

type HttpResponse struct {
	statusLine StatusLine
	headers    Headers
	body       *string
}

type HttpServer struct {
	req        HttpRequest
	res        HttpResponse
	c          net.Conn
	HandleFunc func(ResponseWriter, HttpRequest) error
}

type ResponseWriter interface {
	Write([]byte) (int, error)
	WriteStatusCode(statusCode StatusCode) error
}

func (server *HttpServer) ListenAndServe(port string) {
	l, err := net.Listen("tcp4", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		fmt.Printf("net listen error")
		log.Fatal(err)
	}
	defer (func() {
		err := l.Close()
		if err != nil {
			fmt.Printf("Listener close error")
		}
	})()

	for {
		c, err := l.Accept()
		fmt.Println("Accepted!")
		if err != nil {
			fmt.Printf("Listener accept error")
			fmt.Println(err)
			return
		}
		server.c = c

		go server.handleConnection()
	}
}

func NewHTTPServer(handleFunc func(ResponseWriter, HttpRequest) error) HttpServer {
	return HttpServer{HandleFunc: handleFunc}
}

func (sl *StatusLine) toString() string {
	statusCode := sl.statusCode
	if sl.statusCode == nil {
		statusCode = &OkStatusCode
	}
	return fmt.Sprintf("%s %d %s", sl.httpVersion, *statusCode, sl.statusText)
}

func (rl *RequestLine) toString() string {
	return fmt.Sprintf("%s %s %s", rl.method, rl.requestTarget, rl.httpVersion)
}

func (headers *Headers) toString() string {
	acc := []string{}
	for key, value := range *headers {
		acc = append(acc, fmt.Sprintf("%s: %s", key, value))
	}

	return strings.Join(acc, "\r\n")
}

func (request *HttpRequest) ToString() string {
	if request.body != nil {
		return fmt.Sprintf("%s\r\n%s\r\n\r\n%s\r\n", request.requestLine.toString(), request.headers.toString(), *request.body)
	}

	return fmt.Sprintf("%s\r\n%s\r\n", request.requestLine.toString(), request.headers.toString())
}

func (response *HttpResponse) toString() string {
	if response.body != nil {
		return fmt.Sprintf("%s\r\n%s\r\n\r\n%s\r\n", response.statusLine.toString(), response.headers.toString(), *response.body)
	}

	return fmt.Sprintf("%s\r\n%s\r\n", response.statusLine.toString(), response.headers.toString())
}

func (server *HttpServer) parseHTTPRequest(byteMessage []byte) *HttpResponse {
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
		return &HttpResponse{
			statusLine: StatusLine{
				httpVersion: SupportedHTTPVersion,
				statusCode:  &BadRequestStatusCode,
				statusText:  StatusCodeToStatusText[BadRequestStatusCode],
			},
			headers: map[string]string{"Accept": "*/*"},
			body:    nil,
		}
	}

	startLine := splitMessage[0]
	err := server.parseRequestLine(startLine)
	if err != nil {
		statusCode := err.GetStatusCode()
		body := err.GetBody()
		return &HttpResponse{
			statusLine: StatusLine{
				httpVersion: SupportedHTTPVersion,
				statusCode:  &statusCode,
				statusText:  StatusCodeToStatusText[statusCode],
			},
			headers: map[string]string{"Accept": "*/*"},
			body:    body,
		}
	}

	headersAndBody := splitMessage[1:]
	headersAndBodyLength := len(headersAndBody)

	var messageBodyIndex *int = nil
	for index, line := range headersAndBody {
		isLastValue := index == headersAndBodyLength-1
		isTheBodyNext := line == "" && !isLastValue && headersAndBody[index+1] != ""

		if isTheBodyNext {
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
		statusCode := err.GetStatusCode()
		body := err.GetBody()
		return &HttpResponse{
			statusLine: StatusLine{
				httpVersion: SupportedHTTPVersion,
				statusCode:  &statusCode,
				statusText:  StatusCodeToStatusText[statusCode],
			},
			headers: map[string]string{"Accept": "*/*"},
			body:    body,
		}
	}

	// check body
	if messageBodyIndex == nil {
		return nil
	}
	messageBody := headersAndBody[*messageBodyIndex]
	err = server.parseHTTPMessageBody(messageBody, messageBodyIndex)
	if err != nil {
		statusCode := err.GetStatusCode()
		body := err.GetBody()
		return &HttpResponse{
			statusLine: StatusLine{
				httpVersion: SupportedHTTPVersion,
				statusCode:  &statusCode,
				statusText:  StatusCodeToStatusText[statusCode],
			},
			headers: map[string]string{"Accept": "*/*"},
			body:    body,
		}
	}

	return nil
}

func (server *HttpServer) parseHTTPMessageBody(messageBody string, messageBodyStartIndex *int) CustomError {
	contentLengthHeader := server.req.headers["Content-Length"]
	// Have the message body but it's missing the content length header
	// (we are not looking for the transfer encoding atm)
	// See: https://httpwg.org/specs/rfc9112.html#message.body.length
	hasBodyButNotContentLengthHeader := contentLengthHeader == "" && messageBodyStartIndex != nil
	if hasBodyButNotContentLengthHeader {
		errMsg := "Missing Content-Length header"
		return &HTTPError{
			statusCode: LengthRequiredStatusCode,
			body:       &errMsg,
		}
	}

	server.req.body = &messageBody

	return nil
}

func (server *HttpServer) parseHTTPHeaders(array []string) CustomError {
	headers := make(map[string]string)
	for _, item := range array {
		if item == "" {
			continue
		}
		split := strings.Split(item, ":")
		fieldName := split[0]
		fieldValue := strings.Join(split[1:], ":")

		// https://httpwg.org/specs/rfc9112.html#rfc.section.5.1
		if len(strings.Split(fieldName, " ")) > 1 {
			errMsg := "Cannot have whitespace between field name and colon"
			return &HTTPError{
				statusCode: BadRequestStatusCode,
				body:       &errMsg,
			}
		}

		headers[fieldName] = strings.Trim(fieldValue, " ")
	}

	server.req.headers = headers

	return nil
}

func (server *HttpServer) parseHTTPMethod(str string) (Method, CustomError) {
	switch str {
	case string(get):
		return get, nil
	case string(post):
		return post, nil
	default:
		return "", &HTTPError{statusCode: NotImplementedStatusCode, body: nil}
	}
}

// https://httpwg.org/specs/rfc9112.html#request.target
func (server *HttpServer) parseRequestTarget(str string) (string, CustomError) {
	// Check for whitespaces
	split := strings.Split(str, " ")
	if len(split) > 1 {
		errMsg := "Request target should not contain whitespace"
		return "", &HTTPError{statusCode: BadRequestStatusCode, body: &errMsg}
	}

	return split[0], nil
}

func (server *HttpServer) parseHTTPVersion(str string) (string, CustomError) {
	split := strings.Split(str, "/")
	if len(split) != 2 {
		errMsg := "Wrong HTTP version"
		return "", &HTTPError{statusCode: BadRequestStatusCode, body: &errMsg}
	}

	protocol := split[0]
	version := split[1]

	if protocol != "HTTP" {
		errMsg := "Not an HTTP protocol"
		return "", &HTTPError{statusCode: BadRequestStatusCode, body: &errMsg}
	}

	if version != "1.1" {
		errMsg := "HTTP versin unsupported"
		return "", &HTTPError{statusCode: BadRequestStatusCode, body: &errMsg}
	}

	return str, nil
}

// request-line   = method SP request-target SP HTTP-version
func (server *HttpServer) parseRequestLine(requestLine string) CustomError {
	// split based on whitespace
	split := strings.Split(requestLine, " ")
	if len(split) != 3 {
		errorMsg := "Request line not parseable. Expected: method SP request-target SP HTTP-version"
		return &HTTPError{
			statusCode: BadRequestStatusCode,
			body:       &errorMsg,
		}
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

	server.req.requestLine.method = method
	server.req.requestLine.requestTarget = requestTarget
	server.req.requestLine.httpVersion = httpVersion

	return nil
}

func (server *HttpServer) WriteStatusCode(statusCode StatusCode) error {
	statusText, ok := StatusCodeToStatusText[statusCode]
	if !ok {
		return fmt.Errorf("Could not map this code to a valid status text")
	}

	server.res.statusLine.statusCode = &statusCode
	server.res.statusLine.statusText = statusText

	return nil
}

func (server *HttpServer) Write(message []byte) (int, error) {
	fmt.Printf("Writing message: %s\n", string(message))

	body := string(message)
	response := HttpResponse{
		statusLine: StatusLine{
			httpVersion: SupportedHTTPVersion,
			statusCode:  server.res.statusLine.statusCode,
			statusText:  server.res.statusLine.statusText,
		},
		headers: Headers{},
		body:    &body,
	}

	n, err := server.c.Write([]byte(response.toString()))
	if err != nil {
		body := err.Error()
		statusCode := InternalServerErrorStatusCode
		response := HttpResponse{
			statusLine: StatusLine{
				httpVersion: SupportedHTTPVersion,
				statusCode:  &statusCode,
				statusText:  StatusCodeToStatusText[statusCode],
			},
			headers: Headers{},
			body:    &body,
		}
		server.c.Write([]byte(response.toString()))
		return n, err
	}
	return n, nil
}

func (server *HttpServer) handleConnection() {
	fmt.Printf("Serving %s\n", server.c.RemoteAddr().String())

	packet := make([]byte, 4096)
	tmp := make([]byte, 4096)
	defer (func() {
		err := server.c.Close()
		if err != nil {
			server.WriteStatusCode(InternalServerErrorStatusCode)
			server.Write([]byte(err.Error()))
		}
	})()

	_, err := server.c.Read(tmp)
	if err != nil {
		if err != io.EOF {
			server.WriteStatusCode(InternalServerErrorStatusCode)
			server.Write([]byte(err.Error()))
		}
	}
	packet = append(packet, tmp...)

	error := server.parseHTTPRequest(packet)
	if error != nil {
		server.WriteStatusCode(BadRequestStatusCode)
		server.Write([]byte(error.toString()))
	}
	err = server.HandleFunc(server, server.req)
	if err != nil {
		if server.res.statusLine.statusCode == nil {
			server.WriteStatusCode(BadRequestStatusCode)
		}
		server.Write([]byte(err.Error()))
	}
}
