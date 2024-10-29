package http

import (
	"fmt"
	"io"
	"net"
)

type Proxy struct{}

func (proxy *Proxy) HandleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())

	httpServer := NewHTTPServer()

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

	// TODO: check the path from the endpoint and then choose the
	// appropriate server
	response := httpServer.ParseMessage(packet)

	_, err = c.Write([]byte(response.ToString()))
	if err != nil {
		fmt.Printf("Connection write error")
		fmt.Println(err.Error())
	}
}

func (proxy *Proxy) forward(url string) {

}
