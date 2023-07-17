package anya

import (
	"log"
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
	s.serve(writer, request)
}

func (s *HttpServer) serve(writer http.ResponseWriter, request *http.Request) {
	ctx := newContext(request, writer)
	info, ok := s.findRoute(ctx.req.Method, ctx.req.URL.Path)
	if !ok || info.n == nil {
		ctx.resp.WriteHeader(http.StatusNotFound)
		ctx.resp.Write([]byte("4o4 not found"))
		return
	}

	ctx.params = info.params
	ctx.MatchedRoute = info.n.path
	//add middlewares
	root := info.n.handler
	for j := len(info.n.mdls) - 1; j > -1; j-- {
		root = info.n.mdls[j](root)
	}

	for i := len(s.mdls) - 1; i > -1; i-- {
		root = s.mdls[i](root)
	}

	root = flashResp(root)
	root(ctx)
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
	s.mergeMdls()
	return http.ListenAndServe(s.addr, s)
}

type HandleFunc func(ctx *Context)

func flashResp(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		next(ctx)
		if ctx.statusCode > 0 {
			ctx.resp.WriteHeader(ctx.statusCode)
		}
		_, err := ctx.resp.Write(ctx.respBody)
		if err != nil {
			log.Fatalln("response failed:", err)
		}
	}
}
