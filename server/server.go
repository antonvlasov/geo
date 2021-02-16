package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/antonvlasov/geo/cache"
)

type TelnetServer interface {
	SetHandler(method string, h func(w io.Writer, req *RESTRequest) error)
	HandleRequest(w io.Writer, request *RESTRequest) error
	ListenAndServe(addr *net.TCPAddr) error
}

type telnetServer struct {
	addr     *net.TCPAddr
	handlers map[string]func(w io.Writer, req *RESTRequest) error
}

func NewTelnetServer() TelnetServer {
	return &telnetServer{
		addr:     nil,
		handlers: make(map[string]func(w io.Writer, req *RESTRequest) error, 0),
	}
}

func Run(port int) error {
	server := NewTelnetServer()
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
	server.SetHandler("default", handler)

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}
	err = server.ListenAndServe(addr)
	if err != nil {
		fmt.Println(err)
	}
	return err
}
func (this *telnetServer) SetHandler(method string, h func(w io.Writer, req *RESTRequest) error) {
	this.handlers[method] = h
}
func (this *telnetServer) HandleRequest(w io.Writer, request *RESTRequest) error {
	if request.Method == "" {
		w.Write([]byte("\r\n"))
		return nil
	}
	if request.Method == "default" {
		return errors.New("method does not exist\n")
	}
	h, exists := this.handlers[request.Method]
	if !exists {
		h, exists = this.handlers["default"]
		if !exists {
			return errors.New("method does not exist\n")
		}
	}
	//handler should not send error to w, caller should do it
	err := h(w, request)
	if err != nil {
		return err
	}
	return nil
}

func (this *telnetServer) ListenAndServe(addr *net.TCPAddr) error {
	this.addr = addr
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			return err
		}
		defer conn.Close()
		connReader := bufio.NewReader(conn)
		for {
			bytes, err := connReader.ReadBytes('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			msg := strings.TrimSuffix(string(bytes), "\r\n")

			req, err := RESTParse(msg)
			if err != nil {
				_, err = conn.Write([]byte(err.Error() + "\r\n"))
				if err != nil {
					return err
				}
				continue
			}
			err = this.HandleRequest(conn, &req)
			if err != nil {
				_, err = conn.Write([]byte(err.Error() + "\r\n"))
				if err != nil {
					return err
				}
			}
		}
	}
}

type RESTRequest struct {
	Method string
	Args   []string
}

func RESTParse(request string) (req RESTRequest, err error) {
	request = strings.Trim(request, " ")
	parts := strings.Split(request, " ")
	return RESTRequest{
		Method: parts[0],
		Args:   parts[1:],
	}, nil
}
