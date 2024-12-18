package chttp

import (
	"io"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Version string
	Body    string
	Args    map[string]string
	Headers map[string]string
	Cookie  map[string]string
}

func (r *Request) GetHeader(key string) string {
	return r.Headers[key]
}

func (r *Request) GetArgs(arg string) string {
	return r.Args[arg]
}

func NewRequest(conn io.ReadWriteCloser) (request Request, err error) {
	buf := make([]byte, 1024)

	_, err = conn.Read(buf)
	if err != nil {
		return request, err
	}

	stringBuf := string(buf)

	sp := strings.Split(stringBuf, "\r\n")

	requestLine := strings.Split(sp[0], " ")
	request.Method = strings.ToUpper(requestLine[0])
	request.Version = requestLine[2]

	request.Headers = make(map[string]string)
	for i := 1; i < len(sp); i++ {
		if sp[i] == "" {
			request.Body = sp[i+1]
			break
		}

		header := strings.Split(sp[i], ": ")
		request.Headers[header[0]] = header[1]
	}

	// check if cookie exists in Headers
	if request.Headers["Cookie"] != "" {
		request.Cookie = parseCookie(request.Headers["Cookie"])
	}

	// parse args
	request.Path, request.Args = parseArgs(requestLine[1])

	return request, err
}

func NewRequestFromBuffer(buf []byte) (request Request, err error) {
	stringBuf := string(buf)

	sp := strings.Split(stringBuf, "\r\n")

	requestLine := strings.Split(sp[0], " ")
	request.Method = strings.ToUpper(requestLine[0])
	request.Version = requestLine[2]

	request.Headers = make(map[string]string)
	for i := 1; i < len(sp); i++ {
		if sp[i] == "" {
			request.Body = sp[i+1]
			break
		}

		header := strings.Split(sp[i], ": ")
		request.Headers[header[0]] = header[1]
	}

	// check if cookie exists in Headers
	if request.Headers["Cookie"] != "" {
		request.Cookie = parseCookie(request.Headers["Cookie"])
	}

	// parse args
	request.Path, request.Args = parseArgs(requestLine[1])

	return request, err
}

func parseCookie(cookie string) map[string]string {
	cookieMap := make(map[string]string)
	cookies := strings.Split(cookie, "; ")

	for _, c := range cookies {
		cookie := strings.Split(c, "=")
		cookieMap[cookie[0]] = cookie[1]
	}

	return cookieMap
}

func parseArgs(uri string) (string, map[string]string) {
	s := strings.Split(uri, "?")
	result := make(map[string]string)

	if len(s) == 1 {
		return s[0], result
	}

	args := strings.Split(s[1], "&")

	for _, args := range args {
		arg := strings.Split(args, "=")
		if len(arg) == 1 {
			result[arg[0]] = ""
			continue
		}

		result[arg[0]] = arg[1]
	}

	return s[0], result
}
