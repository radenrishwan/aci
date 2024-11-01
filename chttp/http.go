package chttp

import (
	"context"
	"io"
	"strings"
)

type Context struct {
	context.Context
	Request *Request
	Conn    io.ReadWriteCloser
}

type Handler func(c Context) *Response

type Route struct {
	Handler map[string]Handler
	path    string
}

func (r Route) methodIsExist(method string) bool {
	if _, ok := r.Handler[method]; ok {
		return true
	}

	return false
}

func (r Route) getHandler(method string) Handler {
	return r.Handler[method]
}

type Router struct {
	Handler  map[string]Route
	NotFound Handler
}

func NewRouter() *Router {

	return &Router{
		Handler: make(map[string]Route),
		NotFound: func(c Context) *Response {
			return NewTextResponse("404 Not Found").SetCode(404)
		},
	}
}

func (r Router) routeIsExist(path string) bool {
	_, ok := r.Handler[path]
	return ok
}

// adding handler to router, if you didnt set the method, it will use GET method as default
func (r *Router) HandleFunc(path string, handler Handler) *Router {
	// parse path
	method, path := parsePath(path)

	// check if route is exist
	ok := r.routeIsExist(path)
	if ok {
		existRoute := r.Handler[path]

		if existRoute.methodIsExist(method) {
			// replace handler
			existRoute.Handler[method] = handler
		}

		return r
	}

	// create new route
	route := Route{
		Handler: make(map[string]Handler),
		path:    path,
	}
	route.Handler[method] = handler

	r.Handler[path] = route

	return r
}

func (r *Router) Execute(conn io.ReadWriteCloser) error {
	// parse request
	req, err := NewRequest(conn)
	if err != nil {
		return err
	}

	// get route
	route, ok := r.Handler[req.Path]
	if !ok {
		resp := r.NotFound(Context{
			Context: context.Background(),
			Request: &req,
			conn:    conn,
		})

		resp.Write(conn)

		return nil
	}

	// get handler
	handler := route.getHandler(req.Method)

	// call handler
	resp := handler(Context{
		Context: context.Background(),
		Request: &req,
		conn:    conn,
	})

	// write response
	resp.Write(conn)

	return nil
}

func parsePath(uri string) (method, path string) {
	s := strings.Split(uri, " ")

	if len(s) > 2 {
		panic("invalid uri")
	}

	if len(s) == 1 {
		path = s[0]
		return "GET", path
	}

	method = strings.ToUpper(s[0])
	return method, s[1]
}
