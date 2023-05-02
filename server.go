package anya

import (
	"fmt"
	"net/http"
)

type Server interface {
	http.Handler
	addRoute(method string, path string, handleFunc HandleFunc, middleware ...Middleware) error
	Start() error
}

//var _ Server = &HttpServer{}

type HttpServer struct {
	addr  string
	mdls  []Middleware
	tress map[string]*node
	router
}

func NewHttpServer(addr string) *HttpServer {
	return &HttpServer{
		addr:   addr,
		router: router{m: make(map[string]*node, 8)},
	}
}

func (s *HttpServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("waku waku!"))
	fmt.Println("here")

}

func (s *HttpServer) GET(path string, handleFunc HandleFunc, mdls ...Middleware) {
	s.addRoute(http.MethodGet, path, handleFunc, mdls...)
}

func (s *HttpServer) POST(path string, handleFunc HandleFunc, mdls ...Middleware) {
	s.addRoute(http.MethodPost, path, handleFunc, mdls...)
}

func (s *HttpServer) DELETE(path string, handleFunc HandleFunc, mdls ...Middleware) {
	s.addRoute(http.MethodDelete, path, handleFunc, mdls...)
}

func (s *HttpServer) OPTIONS(path string, handleFunc HandleFunc, mdls ...Middleware) {
	s.addRoute(http.MethodOptions, path, handleFunc, mdls...)
}

func (s *HttpServer) Use(mdls ...Middleware) {
	if s.mdls == nil {
		s.mdls = make([]Middleware, 0, 4)
	}

	s.mdls = append(s.mdls, mdls...)
}

func (s *HttpServer) Start() error {
	// pass s as an instance to tell http package to call method
	return http.ListenAndServe(s.addr, s)
}

type HandleFunc func(ctx *Context)
type Middleware func(handleFunc HandleFunc) HandleFunc
