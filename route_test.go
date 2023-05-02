package anya

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

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
			path: "/",
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
