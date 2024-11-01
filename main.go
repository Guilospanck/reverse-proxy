/*
Test POST:

	curl -X POST http://0.0.0.0:3333 -d @test.json
	curl http://0.0.0.0:3333
*/
package main

import (
	"fmt"
	"reverse_proxy/http"
)

func main() {
	serverA := http.NewHTTPServer(func(w http.ResponseWriter, req http.HttpRequest) error {
		fmt.Println(req)
		_, err := fmt.Fprintf(w, "Server A")
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		return nil
	})
	serverB := http.NewHTTPServer(func(w http.ResponseWriter, req http.HttpRequest) error {
		_, err := fmt.Fprintf(w, "Server B")
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		return nil
	})
	serverC := http.NewHTTPServer(func(w http.ResponseWriter, req http.HttpRequest) error {
		_, err := fmt.Fprintf(w, "Server C")
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		return nil
	})

	go serverA.ListenAndServe("3000")
	go serverB.ListenAndServe("4000")
	go serverC.ListenAndServe("5000")

	proxy := http.ReverseProxyServer{}
	go proxy.Serve()

	for {

	}

}
