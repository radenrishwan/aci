package chttp

import (
	"io"
	"strconv"
)

type Response struct {
	Code int
	// you need to assign a headers map if you create response from [Response],
	// please use [NewResponse] instead to avoid nil headers
	Headers map[string]string
	Body    string
}

func NewResponse() *Response {
	return &Response{
		Code:    200,
		Headers: make(map[string]string),
	}
}

func NewTextResponse(text string) *Response {
	header := map[string]string{
		"Content-Type": "text/plain",
	}

	return &Response{
		Code:    200,
		Headers: header,
		Body:    text,
	}
}

func NewHTMLResponse(html string) *Response {
	header := map[string]string{
		"Content-Type": "text/html",
	}

	return &Response{
		Code:    200,
		Headers: header,
		Body:    html,
	}
}

func NewJSONResponse(json string) *Response {
	header := map[string]string{
		"Content-Type": "application/json",
	}

	return &Response{
		Code:    200,
		Headers: header,
		Body:    json,
	}
}

func (r *Response) AddHeader(key, value string) *Response {
	r.Headers[key] = value

	return r
}

func (r *Response) SetBody(body string) *Response {
	r.Body = body

	return r
}

func (r *Response) SetCode(code int) *Response {
	r.Code = code

	return r
}

func (r *Response) SetCookie(key, value, path string, maxAge int) *Response {
	r.Headers["Set-Cookie"] = key + "=" + value + "; Path=" + path + "; Max-Age=" + strconv.Itoa(maxAge)

	return r
}

func (r *Response) Write(conn io.Writer) error {
	if r == nil {
		r = NewResponse()
	}

	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}

	// check if header has a content-type
	if _, ok := r.Headers["Content-Type"]; !ok {
		r.Headers["Content-Type"] = "text/plain"
	}

	// add content length to Headers
	r.Headers["Content-Length"] = strconv.Itoa(len(r.Body))

	// check if code is 0
	if r.Code == 0 {
		r.Code = 200
	}

	_, err := conn.Write([]byte(
		"HTTP/1.1 " + strconv.Itoa(r.Code) + "\r\n" +
			headerString(r.Headers) +
			"\r\n" +
			r.Body,
	))

	return err
}

func headerString(headers map[string]string) string {
	var headerString string
	for key, value := range headers {
		headerString += key + ": " + value + "\r\n"
	}

	return headerString
}
