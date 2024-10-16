package webserver

import (
	"context"
	"encoding/json"
	"github.com/a-h/templ"
	"net/http"
	"net/url"
	"reflect"
)

func NewURLBodyHandler[T any](
	webServer *WebServer,
	method HTTPMethod,
	pattern string,
	handler func(http.ResponseWriter, *http.Request, T),
) {
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		panic("T must be a struct")
	}

	webServer.NewHandlerBody(method, pattern, func(rw http.ResponseWriter, req *http.Request, body []byte) {
		query, err := url.ParseQuery(string(body))
		if err != nil {
			webServer.settings.Logger.Fatalln(err)
		}

		values := new(T)

		t := reflect.TypeFor[T]()
		v := reflect.ValueOf(values).Elem()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			if !field.IsExported() {
				continue
			}

			value := query.Get(field.Name)

			dst := reflect.New(field.Type)

			if field.Type.Kind() == reflect.String {
				value = "\"" + value + "\""
			}

			err := json.Unmarshal([]byte(value), dst.Interface())
			if err != nil {
				panic(err)
			}

			v.Field(i).Set(dst.Elem())
		}

		handler(rw, req, *values)
	})
}

// htmx templ addon

func NewHTMXTemplURLBodyHandler[D, C any](
	webServer *WebServer,
	component func(arg C) templ.Component,
	method HTTPMethod,
	pattern string,
	handler func(http.ResponseWriter, *http.Request, D) C,
) {
	NewURLBodyHandler(webServer, method, pattern, func(rw http.ResponseWriter, req *http.Request, data D) {
		err := component(handler(rw, req, data)).Render(context.Background(), rw)
		if err != nil {
			webServer.settings.Logger.Fatalln(err)
		}
	})
}
