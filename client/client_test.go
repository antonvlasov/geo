package client

import (
	"bufio"
	"log"
	"net"
	"testing"
)

func TestClient(t *testing.T) {
	var serv mockServer
	var err error
	serv.L, err = net.Listen("tcp", ":1200")
	go serv.serve()
	if err != nil {
		t.Error(err)
	}
	clientConn, err := net.DialTimeout("tcp", "localhost:1200", 0)
	if err != nil {
		t.Error(err)
	}

	err = Set(clientConn, []string{"name", "Anton"})
	if err != nil {
		t.Error(err)
	}
	reader := bufio.NewReader(clientConn)
	bytes, err := reader.ReadBytes('\n')
	if err != nil {
		t.Error(err)
	}
	if string(bytes) != "SET name Anton\r\n" {
		t.Errorf("expected SET name Anton\r\n, got %v", string(bytes))
	}
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
	}
}
