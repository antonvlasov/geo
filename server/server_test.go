package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

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
