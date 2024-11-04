package chttp

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type Context struct {
	context.Context
	Request *Request
	Conn    io.ReadWriteCloser
}

type Handler func(c Context) *Response
type ErrHandler func(c Context, err error) *Response

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
	Handler      map[string]Route
	NotFound     Handler
	ErrorHandler ErrHandler
}

func NewRouter() *Router {
	return &Router{
		Handler: make(map[string]Route),
		NotFound: func(c Context) *Response {
			return NewTextResponse("404 Not Found").SetCode(404)
		},
		ErrorHandler: func(c Context, err error) *Response {
			slog.Error("Error", "err", err)

			return NewTextResponse("500 Internal Server Error").SetCode(500)
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
		resp := r.ErrorHandler(Context{
			Context: context.Background(),
			Request: &req,
			Conn:    conn,
		}, err)

		resp.Write(conn)

		return err
	}

	// get route
	route, ok := r.Handler[req.Path]
	if !ok {
		resp := r.NotFound(Context{
			Context: context.Background(),
			Request: &req,
			Conn:    conn,
		})

		resp.Write(conn)

		return nil
	}

	// get handler
	handler := route.getHandler(req.Method)

	var resp *Response
	func() {
		defer func() {
			rc := recover()
			if rc != nil {
				err = rc.(error)
			}
		}()

		// execute handler
		resp = handler(Context{
			Context: context.Background(),
			Request: &req,
			Conn:    conn,
		})
	}()

	if err != nil {
		resp = r.ErrorHandler(Context{
			Context: context.Background(),
			Request: &req,
			Conn:    conn,
		}, err)

		resp.Write(conn)

		return err
	}

	resp.Write(conn)

	return nil
}

func (r *Router) ServeFile(path string, filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return errors.New("file not found")
	}

	fileType := http.DetectContentType(file)

	r.HandleFunc(path, func(c Context) *Response {
		return NewHTMLResponse(string(file)).SetCode(200).SetHeader("Content-Type", fileType)
	})

	return nil
}

func (r *Router) ServeDir(prefixPath string, filePath string) error {
	// check last character of the path
	if filePath[len(filePath)-1] == '/' {
		filePath = filePath[:len(filePath)-1]
	}

	if prefixPath[len(prefixPath)-1] == '/' {
		prefixPath = prefixPath[:len(prefixPath)-1]
	}

	// get all files in the directory
	files, err := os.ReadDir(filePath)
	if err != nil {
		return errors.New("directory not found")
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// read file
		output, err := os.ReadFile(filePath + "/" + file.Name())
		if err != nil {
			return errors.New("file not found")
		}

		fileType := http.DetectContentType(output)

		// handle file
		r.HandleFunc(prefixPath+"/"+file.Name(), func(c Context) *Response {
			return NewHTMLResponse(string(output)).SetCode(200).SetHeader("Content-Type", fileType)
		})
	}

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
