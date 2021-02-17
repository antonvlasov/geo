package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/antonvlasov/geo/cache"
	"github.com/antonvlasov/geo/client"
)

func TestTelnet(t *testing.T) {
	port := 1200
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		t.Error(err)
	}
	server := NewTelnetServer()
	server.SetHandler("ECHO", echoHandler)
	go launchServer(t, server, addr)

	clientConn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%v", port), time.Second)
	if err != nil {
		t.Error(err)
	}
	defer clientConn.Close()
	request := "ECHO /field value\r\n"
	expected := strings.TrimSuffix(request, "\r\n")
	_, err = clientConn.Write([]byte(request))
	if err != nil {
		t.Error(err)
	}
	reader := bufio.NewReader(clientConn)
	recieved, err := reader.ReadBytes('\n')
	response := strings.TrimSuffix(string(recieved), "\r\n")

	if err != nil {
		t.Error(err)
	}
	if string(response) != expected {
		t.Errorf("expected %v, got %v", expected, string(response))
	}
}

func launchServer(t *testing.T, server TelnetServer, addr *net.TCPAddr) {
	err := server.ListenAndServe(addr)
	if err != nil {
		t.Error(err)
	}
}
func echoHandler(w io.Writer, req *RESTRequest) error {
	var response strings.Builder
	response.WriteString(req.Method)
	for i := range req.Args {
		response.WriteString(" ")
		response.WriteString(req.Args[i])
	}
	response.WriteString("\r\n")
	_, err := w.Write([]byte(response.String()))
	if err != nil {
		return err
	}
	return nil
}
func Run(port int) error {
	CacheServer := NewTelnetServer()
	cache := cache.NewCache()
	go cache.StartCleaner()
	handler := func(w io.Writer, req *RESTRequest) error {
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
func TestCacheHandler(t *testing.T) {
	port := 1205
	go Run(port)

	clientConn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%v", port), 0)
	if err != nil {
		t.Error(err)
	}
	resp := make([]byte, 4000)
	reader := bufio.NewReader(clientConn)

	err = client.Set(clientConn, []string{"name", "Anton"})
	if err != nil {
		t.Error(err)
	}
	n, err := reader.Read(resp)
	if err != nil {
		t.Error(err)
	}
	if string(resp[:n]) != "OK\r\n" {
		t.Errorf("expected %v, got %v", "OK", string(resp[:n]))
	}

	err = client.Set(clientConn, []string{"\\n\\r", "Anton"})
	if err != nil {
		t.Error(err)
	}
	n, err = reader.Read(resp)
	if err != nil {
		t.Error(err)
	}
	if string(resp[:n]) != "OK\r\n" {
		t.Errorf("expected %v, got %v", "OK", string(resp[:n]))
	}

	err = client.HSet(clientConn, []string{"hashmap", "hash1", "val1", "hash2", "val2"})
	if err != nil {
		t.Error(err)
	}
	n, err = reader.Read(resp)
	if err != nil {
		t.Error(err)
	}
	if string(resp[:n]) != "(integer) 2\r\n" {
		t.Errorf("expected %v, got %v", "(integer) 2", string(resp[:n]))
	}

	err = client.LPush(clientConn, []string{"list", "1", "2", "3"})
	if err != nil {
		t.Error(err)
	}
	n, err = reader.Read(resp)
	if err != nil {
		t.Error(err)
	}
	if string(resp[:n]) != "(integer) 3\r\n" {
		t.Errorf("expected %v, got %v", "(integer) 3", string(resp[:n]))
	}

	err = client.Keys(clientConn, []string{"*"})
	if err != nil {
		t.Error(err)
	}
	n, err = reader.Read(resp)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(resp[:n]))

	err = client.HGet(clientConn, []string{"hashmap", "hash1"})
	if err != nil {
		t.Error(err)
	}
	n, err = reader.Read(resp)
	if err != nil {
		t.Error(err)
	}
	if string(resp[:n]) != "val1\r\n" {
		t.Errorf("expected %v, got %v", "val1", string(resp[:n]))
	}
}

func TestMultConn(t *testing.T) {
	port := 1405
	go Run(port)

	nconn := 10
	connPool := make([]net.Conn, nconn)
	readerPool := make([]*bufio.Reader, nconn)
	var err error
	for i := 0; i < nconn; i++ {
		connPool[i], err = net.DialTimeout("tcp", fmt.Sprintf("localhost:%v", port), 0)
		if err != nil {
			t.Error(err)
		}
		readerPool[i] = bufio.NewReader(connPool[i])
	}
	var wg sync.WaitGroup
	wg.Add(nconn)
	for i := 0; i < nconn; i++ {
		go func(wg *sync.WaitGroup, i int) {
			resp := make([]byte, 4000)
			for j := 0; j < 1; j++ {
				key := fmt.Sprintf("key%v%v", i, j)
				value := fmt.Sprintf("value%v%v", i, j)
				err = client.Set(connPool[i], []string{key, value})
				if err != nil {
					t.Error(err)
				}
				n, err := readerPool[i].Read(resp)
				if err != nil {
					t.Error(err)
				}
				if string(resp[:n]) != "OK\r\n" {
					t.Errorf("expected %v, got %v", "OK", string(resp[:n]))
				}
				err = client.Get(connPool[i], []string{key})
				if err != nil {
					t.Error(err)
				}
				n, err = readerPool[i].Read(resp)
				if err != nil {
					t.Error(err)
				}
				if string(resp[:n]) != value+"\r\n" {
					t.Errorf("expected %v, got %v", value, string(resp[:n]))
				}
			}
			wg.Done()
		}(&wg, i)
	}
	wg.Wait()
}
