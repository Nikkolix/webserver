package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

func NewJsonBodyHandler[T any](
	webServer *WebServer,
	method HTTPMethod,
	pattern string,
	handler func(error, http.ResponseWriter, *http.Request, T),
) error {
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		return fmt.Errorf("webserver.NewURLBodyHandler expects a struct type T")
	}

	webServer.NewHandlerBody(
		method,
		pattern,
		func(rw http.ResponseWriter, req *http.Request, body []byte) {
			values := new(T)
			err := json.Unmarshal(body, values)
			handler(err, rw, req, *values)
		},
	)

	return nil
}
