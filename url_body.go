package webserver

import (
	"context"
	"encoding/json"
	"github.com/a-h/templ"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

// parameter interface

type Parameter interface {
	Name() string
	Required() bool
	Valid(value string) bool
	Value(value string) any
}

// string parameter

type StringParameter struct {
	name     string
	required bool
}

func NewStringParameter(name string, required bool) *StringParameter {
	return &StringParameter{
		name:     name,
		required: required,
	}
}

func (p *StringParameter) Name() string {
	return p.name
}

func (p *StringParameter) Required() bool {
	return p.required
}

func (p *StringParameter) Valid(value string) bool {
	return true
}

func (p *StringParameter) Value(value string) any {
	return value
}

// int parameter

type IntParameter struct {
	name     string
	required bool
}

func NewIntParameter(name string, required bool) *IntParameter {
	return &IntParameter{
		name:     name,
		required: required,
	}
}

func (p *IntParameter) Name() string {
	return p.name
}

func (p *IntParameter) Required() bool {
	return p.required
}

func (p *IntParameter) Valid(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}

func (p *IntParameter) Value(value string) any {
	rt, _ := strconv.Atoi(value)
	return rt
}

// custom parameter

type CustomParameter[T any] struct {
	name     string
	required bool
	valid    func(string) bool
	value    func(string) T
}

func NewCustomParameter[T any](
	name string,
	required bool,
	valid func(string) bool,
	value func(string) T,
) *CustomParameter[T] {
	return &CustomParameter[T]{
		name:     name,
		required: required,
		valid:    valid,
		value:    value,
	}
}

func (c CustomParameter[T]) Name() string {
	return c.name
}

func (c CustomParameter[T]) Required() bool {
	return c.required
}

func (c CustomParameter[T]) Valid(value string) bool {
	return c.valid(value)
}

func (c CustomParameter[T]) Value(value string) any {
	return c.value(value)
}

// handler parameter implementation

func (webServer *WebServer) NewHandlerURLBody(method HTTPMethod, pattern string, handler func(http.ResponseWriter, *http.Request, map[string]any), parameters ...Parameter) {
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

		values := make(map[string]any)

		for _, param := range parameters {
			has := query.Has(param.Name())

			if !has {
				if param.Required() {
					webServer.badRequest(rw, param.Name()+" is required")
				}
				continue
			}

			value := query.Get(param.Name())

			if !param.Valid(value) {
				webServer.badRequest(rw, value+" is invalid for parameter "+param.Name())
			}

			values[param.Name()] = param.Value(value)
		}

		handler(rw, req, values)
	})
}

// handler struct implementation

// method HTTPMethod, pattern string,

type URLBodyHandler struct {
	handle func(rw http.ResponseWriter, req *http.Request)
}

func NewURLBodyHandler[T any](webServer *WebServer, handler func(http.ResponseWriter, *http.Request, T)) *URLBodyHandler {
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		panic("T must be a struct")
	}

	out := new(URLBodyHandler)

	out.handle = func(rw http.ResponseWriter, req *http.Request) {
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
	}

	return out
}

func (h URLBodyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.handle(rw, req)
}

// helper todo move elsewhere

func HTMXTemplURLBodyHandler[T any](
	webServer *WebServer,
	component func(arg T) templ.Component,
	handler func(http.ResponseWriter, *http.Request, map[string]any) T,
) func(http.ResponseWriter, *http.Request, map[string]any) {
	return func(rw http.ResponseWriter, req *http.Request, values map[string]any) {
		err := component(handler(rw, req, values)).Render(context.Background(), rw)
		if err != nil {
			webServer.settings.Logger.Fatalln(err)
		}
	}
}

func (webServer *WebServer) badRequest(rw http.ResponseWriter, msg string) {
	rw.WriteHeader(http.StatusBadRequest)
	_, err := rw.Write([]byte(msg))
	if err != nil {
		webServer.settings.Logger.Fatalln(err)
	}
}
