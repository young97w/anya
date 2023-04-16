package anya

import (
	"fmt"
	"net/http"
)

type Server interface {
	http.Handler
	addRoute(method string, handleFunc HandleFunc, middleware ...Middleware)
	Start() error
}

var _ Server = &HttpServer{}

type HttpServer struct {
	addr  string
	mdls  []Middleware
	tress map[string]*node
}

func NewHttpServer(addr string) *HttpServer {
	return &HttpServer{
		addr: addr,
	}
}

func (s *HttpServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("waku waku!"))
	fmt.Println("here")

}

func (s *HttpServer) addRoute(method string, handleFunc HandleFunc, middleware ...Middleware) {
	//TODO implement me
	panic("implement me")
}

func (s *HttpServer) Start() error {
	// pass s as an instance to tell http package to call method
	return http.ListenAndServe(s.addr, s)
}

type HandleFunc func(ctx *Context)
type Middleware func(handleFunc HandleFunc) HandleFunc
