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

	webServer.NewHandlerURLBody(http.MethodGet, "/index", func(rw http.ResponseWriter, req *http.Request, values map[string]any) {

	})

	err := webServer.Run()
	if err != nil {
		panic(err)
	}
}
