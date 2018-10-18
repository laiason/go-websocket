package http_server

import (
	"io"
	"net/http"
)

func HttpHandler(response http.ResponseWriter, request *http.Request)  {
	io.WriteString(response, "http hello, world!\n")
}
