package anya

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Context struct {
	req              *http.Request
	resp             http.ResponseWriter
	params           map[string]string
	statusCode       int
	respBody         []byte
	MatchedRoute     string
	cachedQueryValue url.Values
}

func newContext(req *http.Request, resp http.ResponseWriter) *Context {
	return &Context{
		req:  req,
		resp: resp,
	}
}

// inputs: 1.body(json) 2.form value 3.path value 4. query value 5.header

type StringValue struct {
	val string
	err error
}

// BindJson write request body to val
func (ctx *Context) BindJson(val any) error {
	if ctx.req.Body == nil {
		return errBodyNil
	}
	decoder := json.NewDecoder(ctx.req.Body)
	return decoder.Decode(val)
}

// FormValue return value in form x-www-form-urlencoded
func (ctx *Context) FormValue(key string) StringValue {
	err := ctx.req.ParseForm()
	if err != nil {
		return StringValue{err: err}
	}
	// it's [][]string
	val, ok := ctx.req.Form[key]
	if !ok {
		return StringValue{err: errKeyNotExist(key)}
	}
	return StringValue{val: val[0]}
}

// QueryValue return value in path like: ?name=young
func (ctx *Context) QueryValue(key string) StringValue {
	if ctx.cachedQueryValue == nil {
		ctx.cachedQueryValue = ctx.req.URL.Query()
	}
	val, ok := ctx.cachedQueryValue[key]
	if !ok {
		return StringValue{err: errKeyNotExist(key)}
	}

	return StringValue{val: val[0]}
}

// PathValue return path value
func (ctx *Context) PathValue(key string) StringValue {
	if ctx.params == nil {
		return StringValue{err: errKeyNotExist(key)}
	}
	val, ok := ctx.params[key]
	if !ok {
		return StringValue{err: errKeyNotExist(key)}
	}
	return StringValue{val: val}
}

// HeaderValue return header value
func (ctx *Context) HeaderValue(key string) StringValue {
	val := ctx.req.Header.Get(key)
	return StringValue{val: val}
}

// output:

func (ctx *Context) RespJSON(statusCode int, val any) {
	ctx.statusCode = statusCode
	bytes, err := json.Marshal(val)
	if err != nil {
		ctx.statusCode = http.StatusNotFound
	}
	ctx.respBody = bytes
}

func (ctx *Context) RespJSONOk(val any) {
	ctx.statusCode = http.StatusOK
	bytes, err := json.Marshal(val)
	if err != nil {
		ctx.statusCode = http.StatusNotFound
	}
	ctx.respBody = bytes
}
