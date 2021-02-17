package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

// func Run(port int) error {
// 	CacheServer := server.NewTelnetServer()
// 	cache := cache.NewCache()
// 	go cache.StartCleaner()
// 	handler := func(w io.Writer, req *server.RESTRequest) error {
// 		response, err := cache.HandleRequest(req.Method, req.Args)
// 		if err != nil {
// 			return err
// 		}
// 		_, err = w.Write([]byte(response + "\r\n"))
// 		if err != nil {
// 			return err
// 		}
// 		return err
// 	}
// 	CacheServer.SetHandler("default", handler)

// 	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%v", port))
// 	if err != nil {
// 		return err
// 	}
// 	err = CacheServer.ListenAndServe(addr)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return err
// }
// func main() {
// 	port := 7089
// 	Run(port)
// }

func main() {
	slice := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	go func(slice []int) {
		for {
			time.Sleep(time.Second / 10)
			fmt.Println(slice)
		}
	}(slice)
	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		slice[i] = i * 10
	}
	time.Sleep(time.Second)

	// var serv mockServer
	// var err error
	// serv.L, err = net.Listen("tcp", ":1200")
	// go serv.serve()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// nconn := 10
	// connPool := make([]net.Conn, nconn)
	// readerPool := make([]*bufio.Reader, nconn)
	// for i := 0; i < nconn; i++ {
	// 	connPool[i], err = net.DialTimeout("tcp", "localhost:1200", 0)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	readerPool[i] = bufio.NewReader(connPool[i])
	// }
	// for i := 0; i < nconn; i++ {
	// 	buf := make([]byte, 4096)
	// 	err = client.Set(connPool[i], []string{"name", "Anton"})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	n, err := readerPool[i].Read(buf)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	if string(buf[:n]) != "SET name Anton\r\n" {
	// 		log.Fatal(fmt.Sprintf("expected SET name Anton\r\n, got %v", string(buf[:n])))
	// 	}
	// }

}

type mockServer struct {
	L    net.Listener
	data *string
}

func (s mockServer) Write(msg []byte) (int, error) {
	*s.data = string(msg)
	return len(msg), nil
}

func (s *mockServer) serve() {
	for {
		conn, err := s.L.Accept()
		defer conn.Close()
		go func(conn net.Conn) {
			if err != nil {
				log.Fatal(err)
			}
			reader := bufio.NewReader(conn)
			for {
				bytes, err := reader.ReadBytes('\n')
				if err != nil {
					log.Fatal(err)
				}
				_, err = conn.Write(bytes)
				if err != nil {
					log.Fatal(err)
				}
			}
		}(conn)
	}
}
