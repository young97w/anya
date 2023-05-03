package anya

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	s := NewHttpServer(":8085")
	s.Use(mdw1, mdw2, mdw3)

	s.addRoute(http.MethodPost, "/json", jsonData)    //pass
	s.addRoute(http.MethodPost, "/form", formData)    //pass
	s.addRoute(http.MethodGet, "/:id", pathData)      //pass
	s.addRoute(http.MethodGet, "/query", queryData)   //pass
	s.addRoute(http.MethodGet, "/header", headerData) //pass

	err := s.Start()
	if err != nil {
		t.Fatal(err)
	}
}

type testModel struct {
	Id   string
	Name string
}

func formData(ctx *Context) {
	data := ctx.FormValue("id")
	fmt.Println(data.val)
}

func jsonData(ctx *Context) {
	val := &testModel{}
	ctx.BindJson(val)
	ctx.RespJSONOk(val)
}

func pathData(ctx *Context) {
	data := ctx.PathValue("id")
	fmt.Println(data.val)
}

func queryData(ctx *Context) {
	data := ctx.QueryValue("id")
	fmt.Println(data)
}

func headerData(ctx *Context) {
	data := ctx.HeaderValue("jwt")
	ctx.RespJSONOk(data.val)
}

func mdw1(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		next(ctx)
		fmt.Println("server middleware1")
	}
}

func mdw2(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		next(ctx)
		fmt.Println("server middleware2")
	}
}

func mdw3(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		next(ctx)
		fmt.Println("server middleware3")
	}
}
