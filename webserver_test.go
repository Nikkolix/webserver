package webserver

import (
	"net/http"
	"testing"
)

func TestNewWebServer(t *testing.T) {
	settings := Settings{
		UseHttps:        true,
		UseHttpRedirect: true,
		Hostname:        "localhost",
		HttpPort:        "80",
		HttpsPort:       "443",
		Root:            "root",
		IndexHtml:       "index.html",
		FallbackHtml:    "index.html",
		Logger:          nil,
		CertFile:        "./ssl/certificate.crt",
		KeyFile:         "./ssl/key.key",
	}
	webServer := NewWebServer(settings)

	webServer.NewHandlerBody(http.MethodPost, "/doors", func(rw http.ResponseWriter, req *http.Request, body []byte) {

	}, 1_000_000)

	err := webServer.Run()
	if err != nil {
		panic(err)
	}
}
