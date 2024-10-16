package webserver

import (
	"context"
	"encoding/json"
	"github.com/a-h/templ"
	"io"
	"net/http"
	"net/url"
	"reflect"
)

// handler struct implementation

func NewURLBodyHandler[T any](webServer *WebServer, method HTTPMethod,
	pattern string, handler func(http.ResponseWriter, *http.Request, T)) {
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		panic("T must be a struct")
	}

	webServer.NewHandler(method, pattern, func(rw http.ResponseWriter, req *http.Request) {
		bodyData, err := io.ReadAll(req.Body)
		if err != nil {
			webServer.settings.Logger.Fatalln(err)
		}

		err = req.Body.Close()
		if err != nil {
			webServer.settings.Logger.Fatalln(err)
		}

		query, err := url.ParseQuery(string(bodyData))
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

func NewHTMXTemplURLBodyHandler[T, C any](
	webServer *WebServer,
	component func(arg C) templ.Component,
	method HTTPMethod,
	pattern string,
	handler func(http.ResponseWriter, *http.Request, T) C,
) {
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		panic("T must be a struct")
	}

	webServer.NewHandler(method, pattern, func(rw http.ResponseWriter, req *http.Request) {
		bodyData, err := io.ReadAll(req.Body)
		if err != nil {
			webServer.settings.Logger.Fatalln(err)
		}

		err = req.Body.Close()
		if err != nil {
			webServer.settings.Logger.Fatalln(err)
		}

		query, err := url.ParseQuery(string(bodyData))
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

		err = component(handler(rw, req, *values)).Render(context.Background(), rw)
		if err != nil {
			webServer.settings.Logger.Fatalln(err)
		}
	})
}

// helper todo move elsewhere

func (webServer *WebServer) badRequest(rw http.ResponseWriter, msg string) {
	rw.WriteHeader(http.StatusBadRequest)
	_, err := rw.Write([]byte(msg))
	if err != nil {
		webServer.settings.Logger.Fatalln(err)
	}
}
