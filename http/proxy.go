package http

type ReverseProxyServer struct{}

func (proxy *ReverseProxyServer) Start() {
	// proxyServer := NewHTTPServer()
	//
	// go proxyServer.ListenAndServe("6000")

}

func (proxy *ReverseProxyServer) forward(url string) {

}
