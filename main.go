/*
Test POST:

	curl -X POST http://0.0.0.0:3333 -d @test.json
	curl http://0.0.0.0:3333
*/
package main

import (
	"fmt"
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

	proxy := http.Proxy{}

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Printf("Listener accept error")
			fmt.Println(err)
			return
		}

		go proxy.HandleConnection(c)
	}
}
