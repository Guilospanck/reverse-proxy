/*
Test POST:

	curl -X POST http://0.0.0.0:3333 -d @test.json
	curl http://0.0.0.0:3333
*/
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"reverse_proxy/http"
)

func main() {
	l, err := net.Listen("tcp4", "0.0.0.0:3333")
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
		if err != nil {
			fmt.Printf("Listener accept error")
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())

	httpServer := http.NewHTTPServer()

	packet := make([]byte, 4096)
	tmp := make([]byte, 4096)
	defer (func() {
		err := c.Close()
		if err != nil {
			fmt.Printf("Connection close error")
		}
	})()

	_, err := c.Read(tmp)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Connection read error:", err)
		}
	}
	packet = append(packet, tmp...)
	response := httpServer.ParseMessage(packet)

	_, err = c.Write([]byte(response.ToString()))
	if err != nil {
		fmt.Printf("Connection write error")
		fmt.Println(err.Error())
	}
}
