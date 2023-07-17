package anya

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestFindRoute(t *testing.T) {
	mockHandFunc := func(ctx *Context) {}
	s := NewHttpServer(":8081")

	testCases := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/user/home/*",
		},
		{
			method: http.MethodGet,
			path:   "/user/home/*/store",
		},
		{
			method: http.MethodGet,
			path:   "/user/:id",
		},
		{
			method: http.MethodGet,
			path:   "/user/home/*/store/phone",
		},
		{
			method: http.MethodPost,
			path:   "/",
		},
		{
			method: http.MethodPost,
			path:   "/:id",
		},
		{
			method: http.MethodPost,
			path:   "/:id/:number(\\d+)",
		},
	}

	for _, tc := range testCases {
		s.addRoute(tc.method, tc.path, mockHandFunc)
	}

	//find star child
	info, ok := s.findRoute(http.MethodGet, "/user/home/hello")
	assert.Equal(t, true, ok)
	assert.Equal(t, "*", info.n.path)

	//find param child
	info, ok = s.findRoute(http.MethodGet, "/user/114514")
	assert.Equal(t, true, ok)
	assert.Equal(t, "114514", info.params["id"])

	//find regex child
	info, ok = s.findRoute(http.MethodPost, "/114514/hj114514")
	assert.Equal(t, true, ok)

	assert.Equal(t, "114514", info.params["number"])

	//find normal child
	info, ok = s.findRoute(http.MethodGet, "/user/home")
	assert.Equal(t, true, ok)
	assert.Equal(t, "home", info.n.path)

}

func TestAddRoute(t *testing.T) {
	mockHandFunc := func(ctx *Context) {}
	s := NewHttpServer(":8081")

	testCases := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/user/home/*",
		},
		{
			method: http.MethodGet,
			path:   "/user/home/*/store",
		},
		{
			method: http.MethodGet,
			path:   "/user/:id",
		},
		{
			method: http.MethodGet,
			path:   "/user/home/*/store/phone",
		},
		{
			method: http.MethodPost,
			path:   "/",
		},
		{
			method: http.MethodPost,
			path:   "/:id",
		},
		{
			method: http.MethodPost,
			path:   "/:id/:number(\\d+)",
		},
	}

	tree := map[string]*node{
		http.MethodGet: {
			path: "/",
			children: map[string]*node{
				"user": {
					path:       "user",
					paramChild: &node{path: ":id", typ: nodeParam, param: "id", handler: mockHandFunc},
					children: map[string]*node{
						"home": {
							path: "home",
							starChild: &node{
								path:    "*",
								typ:     nodeStar,
								handler: mockHandFunc,
								children: map[string]*node{
									"store": {
										path:    "store",
										typ:     nodeStatic,
										handler: mockHandFunc,
										children: map[string]*node{
											"phone": {
												path:    "phone",
												typ:     nodeStatic,
												handler: mockHandFunc,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		http.MethodPost: {
			path:    "/",
			handler: mockHandFunc,
			paramChild: &node{
				path:    ":id",
				typ:     nodeParam,
				param:   "id",
				handler: mockHandFunc,
				regChild: &node{
					path:    ":number(\\d+)",
					typ:     nodeRegex,
					param:   "number",
					handler: mockHandFunc,
				},
			},
		},
	}

	for _, tc := range testCases {
		s.addRoute(tc.method, tc.path, mockHandFunc)
	}

	res, ok := compareTree(tree, s.m)
	assert.Equal(t, true, ok)
	assert.Equal(t, "", res)

	// invalid cases
	//重复注册
	err := s.addRoute(http.MethodPost, "/:id", mockHandFunc)
	assert.Equal(t, errPathRegistered(":id"), err)

	err = s.addRoute(http.MethodPost, "/", mockHandFunc)
	assert.Equal(t, errPathRegistered("/"), err)

	//非法路径
	err = s.addRoute(http.MethodPost, "user", mockHandFunc)
	assert.Equal(t, errInvalidPath("user"), err)

	err = s.addRoute(http.MethodPost, "/user//hello", mockHandFunc)
	assert.Equal(t, errInvalidPath("/user//hello"), err)

}

func compareTree(target, result map[string]*node) (string, bool) {
	for method, tRoot := range target {
		rRoot, ok := result[method]
		if !ok {
			return fmt.Sprintf("缺少方法树:%v", method), false
		}

		//compare tree
		s, res := compareNode(tRoot, rRoot)
		if !res {
			return s, false
		}
	}
	return "", true
}

func compareNode(t, r *node) (string, bool) {
	//先对比children
	var res bool
	var s string
	if t.starChild != nil {
		s, res = compareNode(t.starChild, r.starChild)
		if !res {
			return s, false
		}
	}

	if t.paramChild != nil {
		s, res = compareNode(t.paramChild, r.paramChild)
		if !res {
			return s, false
		}
	}

	if t.regChild != nil {
		s, res = compareNode(t.regChild, r.regChild)
		if !res {
			return s, false
		}
	}

	if len(t.children) > 0 {
		for path, child := range t.children {
			rChild, ok := r.children[path]
			if !ok {
				return fmt.Sprintf("缺少子节点:%v", path), false
			}

			s, res = compareNode(child, rChild)
			if !res {
				return s, false
			}
		}
	}

	switch {
	case t.typ == nodeParam || t.typ == nodeRegex:
		res = t.param == r.param
		if !res {
			return "param 不相等", false
		}
	}

	if t.handler != nil {
		res = reflect.ValueOf(t.handler).Pointer() == reflect.ValueOf(r.handler).Pointer()
		if !res {
			return "handler 不相等", false
		}
	}

	if len(t.mdls) > 0 {
		for i, mdl := range t.mdls {
			if reflect.ValueOf(r.mdls[i]).Pointer() != reflect.ValueOf(mdl).Pointer() {
				return "middleware 不相等", false
			}
		}
	}

	return "", true
}

func TestHandleFuncCompare(t *testing.T) {
	mockHandFunc := func(ctx *Context) {}
	type hdf struct {
		hdf HandleFunc
	}

	h1 := hdf{
		hdf: mockHandFunc,
	}

	h2 := hdf{
		hdf: mockHandFunc,
	}

	assert.Equal(t, reflect.ValueOf(h2.hdf).Pointer(), reflect.ValueOf(h1.hdf).Pointer())

}

func TestAddMdls(t *testing.T) {
	mockHandFunc := func(ctx *Context) {}
	testCases := []struct {
		method string
		path   string
		mdls   []Middleware
	}{
		{
			method: http.MethodGet,
			path:   "/*",
			mdls:   []Middleware{mdwRoute1},
		},
		{
			method: http.MethodGet,
			path:   "/user",
			mdls:   []Middleware{mdwRoute2},
		},
		{
			method: http.MethodGet,
			path:   "/user/home/*",
			mdls:   []Middleware{mdwRoute3},
		},
		{
			method: http.MethodGet,
			path:   "/user/home/*/store",
		},
		{
			method: http.MethodGet,
			path:   "/user/:id",
		},
		{
			method: http.MethodGet,
			path:   "/user/:id/profile",
		},
	}

	mTree := map[string]*node{
		http.MethodGet: {
			path: "/",
			children: map[string]*node{
				"user": {
					path: "user",
					paramChild: &node{
						path:    ":id",
						typ:     nodeParam,
						param:   "id",
						handler: mockHandFunc,
						mdls:    []Middleware{mdwRoute1, mdwRoute2},
						children: map[string]*node{
							"profile": {
								path:    "profile",
								handler: mockHandFunc,
								typ:     nodeStatic,
								mdls:    []Middleware{mdwRoute1, mdwRoute2},
							},
						}},
					children: map[string]*node{
						"home": {
							path: "home",
							starChild: &node{
								path:    "*",
								typ:     nodeStar,
								handler: mockHandFunc,
								children: map[string]*node{
									"store": {
										path:    "store",
										typ:     nodeStatic,
										handler: mockHandFunc,
										mdls:    []Middleware{mdwRoute1, mdwRoute2, mdwRoute3},
									},
								},
								mdls: []Middleware{mdwRoute1, mdwRoute2, mdwRoute3},
							},
							mdls: []Middleware{mdwRoute1, mdwRoute2},
						},
					},
					mdls: []Middleware{mdwRoute1, mdwRoute2},
				},
			},
			starChild: &node{
				path:       "*",
				handler:    mockHandFunc,
				children:   nil,
				starChild:  nil,
				paramChild: nil,
				param:      "",
				regChild:   nil,
				regExp:     nil,
				typ:        0,
				mdls:       []Middleware{mdwRoute1},
			},
		},
	}

	s := NewHttpServer(":8081")
	for _, tc := range testCases {
		s.addRoute(tc.method, tc.path, mockHandFunc, tc.mdls...)
	}
	s.mergeMdls()

	res, ok := compareTree(mTree, s.m)
	assert.Equal(t, true, ok)
	assert.Equal(t, "", res)
}

func mdwRoute1(handleFunc HandleFunc) HandleFunc {
	return func(ctx *Context) {
		fmt.Println("mdwRoute1 start------")
		handleFunc(ctx)
		fmt.Println("mdwRoute1 end--------")
	}
}

func mdwRoute2(handleFunc HandleFunc) HandleFunc {
	return func(ctx *Context) {
		fmt.Println("mdwRoute2 start------")
		handleFunc(ctx)
		fmt.Println("mdwRoute2 end--------")
	}
}

func mdwRoute3(handleFunc HandleFunc) HandleFunc {
	return func(ctx *Context) {
		fmt.Println("mdwRoute3 start------")
		handleFunc(ctx)
		fmt.Println("mdwRoute3 end--------")
	}
}

func TestMdlsEq(t *testing.T) {
	arr := []Middleware{mdwRoute1, mdwRoute2, mdwRoute3}
	root := rootHandleFunc
	for i := 2; i > -1; i-- {
		root = arr[i](root)
	}
	root(&Context{})
}

func compareMdls(tMdls, rMdls []Middleware) {
	for i, mdl := range tMdls {
		if reflect.ValueOf(rMdls[i]).Type() == reflect.ValueOf(mdl).Type() {
			fmt.Println("not equal")
		}
	}
}

func rootHandleFunc(*Context) {
	fmt.Println("root handle func")
}
