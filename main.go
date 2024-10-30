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
	serverA := http.NewHTTPServer(func(hr1 http.HttpResponse, hr2 http.HttpRequest) error {
		fmt.Println("Hello from server A")
		fmt.Println(hr1)
		fmt.Println(hr2)
		return nil
	})
	serverB := http.NewHTTPServer(func(hr1 http.HttpResponse, hr2 http.HttpRequest) error {
		fmt.Println("Hello from server B")
		fmt.Println(hr1)
		fmt.Println(hr2)
		return nil
	})
	serverC := http.NewHTTPServer(func(hr1 http.HttpResponse, hr2 http.HttpRequest) error {
		fmt.Println("Hello from server C")
		fmt.Println(hr1)
		fmt.Println(hr2)
		return nil
	})

	go serverA.ListenAndServe("3000")
	go serverB.ListenAndServe("4000")
	go serverC.ListenAndServe("5000")

	for {

	}

}
