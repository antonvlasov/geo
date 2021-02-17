package main

import (
	"fmt"
	"io"
	"net"

	"github.com/antonvlasov/geo/cache"
	"github.com/antonvlasov/geo/server"
)

func Run(port int) error {
	CacheServer := server.NewTelnetServer()
	cache := cache.NewCache()
	go cache.StartCleaner()
	handler := func(w io.Writer, req *server.RESTRequest) error {
		response, err := cache.HandleRequest(req.Method, req.Args)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(response + "\r\n"))
		if err != nil {
			return err
		}
		return err
	}
	CacheServer.SetHandler("default", handler)

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}
	err = CacheServer.ListenAndServe(addr)
	if err != nil {
		fmt.Println(err)
	}
	return err
}
func main() {
	port := 7089
	Run(port)
}
