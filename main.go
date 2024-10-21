package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"reverse_proxy/http"
)

func main() {
	l, err := net.Listen("tcp4", ":8000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
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
	defer c.Close()

	_, err := c.Read(tmp)
	if err != nil {
		if err != io.EOF {
			fmt.Println("read error:", err)
		}
	}
	packet = append(packet, tmp...)
	response := httpServer.ParseMessage(packet)
	fmt.Println(response.ToString())

	_, err = c.Write([]byte(response.ToString()))
	if err != nil {
		fmt.Println(err.Error())
	}
}
