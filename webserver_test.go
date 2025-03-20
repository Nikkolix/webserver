package webserver

import (
	"net/http"
	"testing"
)

func TestNewWebServer(t *testing.T) {
	settings := Settings{
		UseHttps:         true,
		UseHttpRedirect:  true,
		Hostname:         "localhost",
		HttpPort:         "80",
		HttpsPort:        "443",
		Root:             "root",
		Logger:           nil,
		CertFile:         "./ssl/certificate.crt",
		KeyFile:          "./ssl/privatekey.key",
		FallbackRedirect: "/index",
	}
	webServer := NewWebServer(settings)

	webServer.NewHandleFunc(http.MethodGet, "/index", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte("Hello World"))
		if err != nil {
			panic(err)
		}
	})

	err := webServer.Run()
	if err != nil {
		panic(err)
	}
}
